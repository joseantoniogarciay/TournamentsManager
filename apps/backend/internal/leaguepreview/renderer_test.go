package leaguepreview

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testLeagueID = "018f1a4b-7197-7db4-8e4d-5d528526188d"

func TestHandlerRendersLeagueMetadataAtCanonicalURL(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "api.fasttourney.test" {
			t.Errorf("host = %q, want API host", r.Host)
		}
		if r.URL.Path != "/v1/leagues/"+testLeagueID {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"name":"Liga <final>","state":"in_progress","teams":[{},{}]}`)
	}))
	defer api.Close()

	handler := newTestHandler(t, api.URL)
	request := httptest.NewRequest(http.MethodGet, "/league/"+testLeagueID, nil)
	request.Host = "fasttourney.test"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("X-Robots-Tag"); got != "noindex, nofollow, noarchive" {
		t.Errorf("X-Robots-Tag = %q", got)
	}
	body := response.Body.String()
	for _, expected := range []string{
		"Liga &lt;final&gt; | FastTourney",
		`content="Football league · 2 teams · In progress on FastTourney."`,
		`property="og:url" content="https://fasttourney.test/league/` + testLeagueID + `"`,
		`property="og:image" content="https://fasttourney.test/fasttourney-league-preview.png"`,
		`name="twitter:card" content="summary_large_image"`,
		`rel="canonical" href="https://fasttourney.test/league/` + testLeagueID + `"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response does not contain %q: %s", expected, body)
		}
	}
}

func TestHandlerReturnsNotFoundForUnavailableLeague(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.NotFoundHandler())
	defer api.Close()
	handler := newTestHandler(t, api.URL)
	request := httptest.NewRequest(http.MethodGet, "/league/"+testLeagueID, nil)
	request.Host = "fasttourney.test"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", response.Code)
	}
}

func TestHandlerDoesNotCaptureNestedLeagueRoutes(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t, "http://127.0.0.1:1/v1")
	request := httptest.NewRequest(http.MethodGet, "/league/"+testLeagueID+"/standings", nil)
	request.Host = "fasttourney.test"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", response.Code)
	}
}

func newTestHandler(t *testing.T, apiURL string) http.Handler {
	t.Helper()
	webRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(webRoot, "index.html"), []byte(`<!doctype html><html><head><title>FastTourney</title><meta name="description" content="Home"><link rel="canonical" href="https://fasttourney.test/"></head><body><div id="root"></div></body></html>`), 0o600); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{
		WebRoot: webRoot, PublicBaseURL: "https://fasttourney.test", PublicHost: "fasttourney.test", APIBaseURL: apiURL + "/v1", APIHost: "api.fasttourney.test",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
