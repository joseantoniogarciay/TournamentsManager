// Package smtp entrega correo mediante SMTP.
package smtp

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/url"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

// Mailer es el adaptador SMTP del correo de verificación local.
type Mailer struct {
	address string
	from    string
	baseURL *url.URL
}

// NewMailer construye un adaptador SMTP sin credenciales para Mailpit local.
func NewMailer(address, from, baseURL string) (Mailer, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Host == "" || !validPublicURL(parsedURL) {
		return Mailer{}, fmt.Errorf("URL pública inválida")
	}
	return Mailer{address: address, from: from, baseURL: parsedURL}, nil
}

// SendVerification entrega un enlace HTTPS que la persona confirma explícitamente
// en el cliente. El GET nunca modifica el estado de la cuenta.
func (m Mailer) SendVerification(ctx context.Context, recipient string, locale registration.Locale, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	verificationURL := *m.baseURL
	verificationURL.Path = "/link/confirm"
	query := verificationURL.Query()
	query.Set("token", token)
	verificationURL.RawQuery = query.Encode()

	message, err := verificationMessage(locale, recipient, m.from, verificationURL.String())
	if err != nil {
		return err
	}
	if err := smtp.SendMail(m.address, nil, m.from, []string{recipient}, message); err != nil {
		return fmt.Errorf("SMTP: %w", err)
	}
	return nil
}

func validPublicURL(parsedURL *url.URL) bool {
	if parsedURL.Scheme == "https" {
		return true
	}
	if parsedURL.Scheme != "http" {
		return false
	}
	hostname := strings.ToLower(parsedURL.Hostname())
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}

func verificationMessage(locale registration.Locale, recipient, from, verificationURL string) ([]byte, error) {
	localizedCopy, ok := localizedVerificationCopy(locale)
	if !ok {
		return nil, fmt.Errorf("locale de email no admitido: %q", locale)
	}

	var message bytes.Buffer
	boundary := "fasttourney-verification"
	message.WriteString("To: " + recipient + "\r\n")
	message.WriteString("From: " + from + "\r\n")
	message.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", localizedCopy.Subject) + "\r\n")
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n")

	writer := multipart.NewWriter(&message)
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, err
	}
	plain, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/plain; charset=UTF-8"}})
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(plain, "%s\r\n\r\n%s\r\n%s\r\n\r\n%s\r\n", localizedCopy.Title, localizedCopy.PlainPrompt, verificationURL, localizedCopy.Ignore); err != nil {
		return nil, err
	}

	htmlPart, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/html; charset=UTF-8"}})
	if err != nil {
		return nil, err
	}
	view := struct {
		verificationCopy
		VerificationURL string
	}{verificationCopy: localizedCopy, VerificationURL: verificationURL}
	if err := verificationHTML.Execute(htmlPart, view); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return message.Bytes(), nil
}

type verificationCopy struct {
	Language    string
	Subject     string
	Tagline     string
	Title       string
	Description string
	Button      string
	Ignore      string
	PlainPrompt string
}

func localizedVerificationCopy(locale registration.Locale) (verificationCopy, bool) {
	switch locale {
	case registration.LocaleSpanish:
		return verificationCopy{"es", "Verifica tu cuenta de Fast Tourney", "Tus torneos, a tu ritmo", "Verifica tu cuenta", "Solo queda confirmar tu correo para empezar a organizar tus torneos.", "Verificar mi cuenta", "Si no has creado una cuenta de Fast Tourney, puedes ignorar este correo.", "Abre este enlace y confirma tu cuenta:"}, true
	case registration.LocaleEnglish:
		return verificationCopy{"en", "Verify your Fast Tourney account", "Your tournaments, your way", "Verify your account", "Confirm your email to start organizing your tournaments.", "Verify my account", "If you did not create a Fast Tourney account, you can ignore this email.", "Open this link to confirm your account:"}, true
	case registration.LocaleItalian:
		return verificationCopy{"it", "Verifica il tuo account Fast Tourney", "I tuoi tornei, a modo tuo", "Verifica il tuo account", "Conferma la tua email per iniziare a organizzare i tuoi tornei.", "Verifica il mio account", "Se non hai creato un account Fast Tourney, puoi ignorare questa email.", "Apri questo link per confermare il tuo account:"}, true
	case registration.LocaleFrench:
		return verificationCopy{"fr", "Vérifiez votre compte Fast Tourney", "Vos tournois, à votre rythme", "Vérifiez votre compte", "Confirmez votre adresse e-mail pour commencer à organiser vos tournois.", "Vérifier mon compte", "Si vous n'avez pas créé de compte Fast Tourney, vous pouvez ignorer cet e-mail.", "Ouvrez ce lien pour confirmer votre compte :"}, true
	default:
		return verificationCopy{}, false
	}
}

var verificationHTML = template.Must(template.New("verification").Parse(`<!doctype html>
<html lang="{{.Language}}"><body style="margin:0;background:#f8fafc;color:#101828;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding:32px 16px"><tr><td align="center">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e2e8f0;border-radius:20px;overflow:hidden">
      <tr><td style="padding:28px 32px;background:linear-gradient(135deg,#5b4bff,#e84a8a);color:#ffffff"><strong style="font-size:20px">Fast Tourney</strong><br><span style="font-size:13px;opacity:.9">{{.Tagline}}</span></td></tr>
      <tr><td style="padding:32px"><h1 style="margin:0 0 12px;font-size:26px;line-height:1.2">{{.Title}}</h1><p style="margin:0 0 24px;line-height:1.55">{{.Description}}</p>
      <a href="{{.VerificationURL}}" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#5b4bff;color:#ffffff;text-decoration:none;font-weight:700">{{.Button}}</a>
      <p style="margin:28px 0 0;color:#667085;font-size:13px;line-height:1.5">{{.Ignore}}</p></td></tr>
    </table>
  </td></tr></table>
</body></html>`))
