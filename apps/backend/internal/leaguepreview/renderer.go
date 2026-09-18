// Package leaguepreview renders the small server-side document needed when a
// public league URL is shared. It deliberately consumes the public HTTP
// projection instead of reaching PostgreSQL or the application internals.
package leaguepreview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const (
	cacheControl           = "no-store"
	maxShellBytes          = 5 << 20
	maximumTournamentTitle = 120
)

var leagueIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// Config defines the one public host and its local, authenticated API route.
// APIBaseURL is deliberately loopback Caddy, so its edge-token policy remains
// the only credential boundary between this process and the API.
type Config struct {
	WebRoot       string
	PublicBaseURL string
	PublicHost    string
	APIBaseURL    string
	APIHost       string
}

type publicTournament struct {
	Name  string            `json:"name"`
	State string            `json:"state"`
	Teams []json.RawMessage `json:"teams"`
}

// NewHandler returns the canonical /tournament/{id} renderer. It is intentionally
// useful to browsers as well as crawlers: no User-Agent sniffing means the URL
// is stable and a person can still hydrate the normal static application.
func NewHandler(config Config, client *http.Client) (http.Handler, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serve(config, client, w, r)
	}), nil
}

func (config Config) validate() error {
	if config.WebRoot == "" || config.PublicHost == "" || config.APIHost == "" {
		return errors.New("web root and hosts are required")
	}
	publicURL, err := url.Parse(config.PublicBaseURL)
	if err != nil || publicURL.Scheme != "https" || publicURL.Host == "" {
		return fmt.Errorf("invalid public URL: %q", config.PublicBaseURL)
	}
	apiURL, err := url.Parse(config.APIBaseURL)
	if err != nil || (apiURL.Scheme != "http" && apiURL.Scheme != "https") || !isLoopbackHost(apiURL.Hostname()) {
		return fmt.Errorf("API URL must use loopback: %q", config.APIBaseURL)
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	address, err := netip.ParseAddr(host)
	return err == nil && address.IsLoopback()
}

func serve(config Config, client *http.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || hostWithoutPort(r.Host) != config.PublicHost {
		http.NotFound(w, r)
		return
	}
	leagueID, ok := leagueIDFromPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	league, status, err := fetchTournament(r.Context(), client, config, leagueID)
	if err != nil {
		http.Error(w, "Tournament preview is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	if status == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}
	if status != http.StatusOK {
		http.Error(w, "Tournament preview is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	shell, err := readShell(config.WebRoot)
	if err != nil {
		http.Error(w, "Tournament preview is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	page, err := render(shell, league, canonicalURL(config.PublicBaseURL, r.URL.Path))
	if err != nil {
		http.Error(w, "Tournament preview is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	_, _ = w.Write(page)
}

func leagueIDFromPath(requestPath string) (string, bool) {
	const prefix = "/tournament/"
	if !strings.HasPrefix(requestPath, prefix) {
		return "", false
	}
	leagueID := strings.ToLower(strings.TrimPrefix(requestPath, prefix))
	return leagueID, leagueIDPattern.MatchString(leagueID)
}

func fetchTournament(ctx context.Context, client *http.Client, config Config, leagueID string) (publicTournament, int, error) {
	// #nosec G704 -- validate permits only a loopback API URL controlled by launchd.
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(config.APIBaseURL, "/")+"/tournaments/"+leagueID, nil)
	if err != nil {
		return publicTournament{}, 0, err
	}
	request.Host = config.APIHost
	// #nosec G704 -- request above can only target the validated loopback Caddy route.
	response, err := client.Do(request)
	if err != nil {
		return publicTournament{}, 0, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return publicTournament{}, response.StatusCode, nil
	}
	var league publicTournament
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&league); err != nil {
		return publicTournament{}, 0, err
	}
	return league, response.StatusCode, nil
}

func readShell(webRoot string) ([]byte, error) {
	// #nosec G304 -- WebRoot is launchd configuration; the basename is constant.
	file, err := os.Open(path.Join(webRoot, "index.html"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return io.ReadAll(io.LimitReader(file, maxShellBytes+1))
}

func render(shell []byte, league publicTournament, canonical string) ([]byte, error) {
	if len(shell) > maxShellBytes {
		return nil, errors.New("static shell exceeds maximum size")
	}
	document, err := html.Parse(bytes.NewReader(shell))
	if err != nil {
		return nil, err
	}
	head := findElement(document, "head")
	if head == nil {
		return nil, errors.New("static shell has no head")
	}
	title := truncate(strings.TrimSpace(league.Name), maximumTournamentTitle) + " | FastTourney"
	description := leagueDescription(league)
	imageURL := canonicalURL(canonicalBaseURL(canonical), "/fasttourney-league-preview.png")
	setTitle(head, title)
	setMeta(head, "name", "description", description)
	setMeta(head, "name", "robots", "noindex, nofollow, noarchive")
	setMeta(head, "property", "og:type", "website")
	setMeta(head, "property", "og:title", title)
	setMeta(head, "property", "og:description", description)
	setMeta(head, "property", "og:url", canonical)
	setMeta(head, "property", "og:image", imageURL)
	setMeta(head, "property", "og:image:width", "1200")
	setMeta(head, "property", "og:image:height", "630")
	setMeta(head, "property", "og:image:alt", "FastTourney football league")
	setMeta(head, "name", "twitter:card", "summary_large_image")
	setMeta(head, "name", "twitter:title", title)
	setMeta(head, "name", "twitter:description", description)
	setMeta(head, "name", "twitter:image", imageURL)
	setCanonical(head, canonical)

	var output bytes.Buffer
	if err := html.Render(&output, document); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func canonicalBaseURL(canonical string) string {
	parsed, err := url.Parse(canonical)
	if err != nil {
		return canonical
	}
	return parsed.Scheme + "://" + parsed.Host
}

func canonicalURL(baseURL, requestPath string) string {
	return strings.TrimRight(baseURL, "/") + requestPath
}

func leagueDescription(league publicTournament) string {
	teamLabel := "teams"
	if len(league.Teams) == 1 {
		teamLabel = "team"
	}
	return fmt.Sprintf("Football league · %d %s · %s on FastTourney.", len(league.Teams), teamLabel, stateLabel(league.State))
}

func stateLabel(state string) string {
	switch state {
	case "published":
		return "Published"
	case "in_progress":
		return "In progress"
	case "completed":
		return "Completed"
	case "cancelled":
		return "Cancelled"
	default:
		return "Available"
	}
}

func truncate(value string, maximum int) string {
	if utf8.RuneCountInString(value) <= maximum {
		return value
	}
	runes := []rune(value)
	return string(runes[:maximum-1]) + "…"
}

func findElement(node *html.Node, tag string) *html.Node {
	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func setTitle(head *html.Node, value string) {
	title := findElement(head, "title")
	if title == nil {
		title = &html.Node{Type: html.ElementNode, Data: "title"}
		head.AppendChild(title)
	}
	title.FirstChild = nil
	title.LastChild = nil
	title.AppendChild(&html.Node{Type: html.TextNode, Data: value})
}

func setMeta(head *html.Node, attribute, key, content string) {
	for child := head.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "meta" && attributeValue(child, attribute) == key {
			setAttribute(child, "content", content)
			return
		}
	}
	meta := &html.Node{Type: html.ElementNode, Data: "meta"}
	setAttribute(meta, attribute, key)
	setAttribute(meta, "content", content)
	head.AppendChild(meta)
}

func setCanonical(head *html.Node, href string) {
	for child := head.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "link" && attributeValue(child, "rel") == "canonical" {
			setAttribute(child, "href", href)
			return
		}
	}
	link := &html.Node{Type: html.ElementNode, Data: "link"}
	setAttribute(link, "rel", "canonical")
	setAttribute(link, "href", href)
	head.AppendChild(link)
}

func attributeValue(node *html.Node, key string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == key {
			return attribute.Val
		}
	}
	return ""
}

func setAttribute(node *html.Node, key, value string) {
	for index := range node.Attr {
		if node.Attr[index].Key == key {
			node.Attr[index].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}

func hostWithoutPort(host string) string {
	if parsedHost, _, err := strings.Cut(host, ":"); err {
		return parsedHost
	}
	return host
}
