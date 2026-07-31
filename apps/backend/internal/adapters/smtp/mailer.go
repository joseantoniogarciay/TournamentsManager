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
func (m Mailer) SendVerification(ctx context.Context, recipient, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	verificationURL := *m.baseURL
	verificationURL.Path = "/link/confirm"
	query := verificationURL.Query()
	query.Set("token", token)
	verificationURL.RawQuery = query.Encode()

	message, err := verificationMessage(recipient, m.from, verificationURL.String())
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

func verificationMessage(recipient, from, verificationURL string) ([]byte, error) {
	var message bytes.Buffer
	boundary := "fasttourney-verification"
	message.WriteString("To: " + recipient + "\r\n")
	message.WriteString("From: " + from + "\r\n")
	message.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", "Verifica tu cuenta de Fast Tourney") + "\r\n")
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
	if _, err := fmt.Fprintf(plain, "Verifica tu cuenta de Fast Tourney\r\n\r\nAbre este enlace y confirma tu cuenta:\r\n%s\r\n\r\nSi no has creado una cuenta, puedes ignorar este correo.\r\n", verificationURL); err != nil {
		return nil, err
	}

	htmlPart, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/html; charset=UTF-8"}})
	if err != nil {
		return nil, err
	}
	view := struct{ VerificationURL string }{VerificationURL: verificationURL}
	if err := verificationHTML.Execute(htmlPart, view); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return message.Bytes(), nil
}

var verificationHTML = template.Must(template.New("verification").Parse(`<!doctype html>
<html lang="es"><body style="margin:0;background:#f8fafc;color:#101828;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding:32px 16px"><tr><td align="center">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e2e8f0;border-radius:20px;overflow:hidden">
      <tr><td style="padding:28px 32px;background:linear-gradient(135deg,#5b4bff,#e84a8a);color:#ffffff"><strong style="font-size:20px">Fast Tourney</strong><br><span style="font-size:13px;opacity:.9">Tus torneos, a tu ritmo</span></td></tr>
      <tr><td style="padding:32px"><h1 style="margin:0 0 12px;font-size:26px;line-height:1.2">Verifica tu cuenta</h1><p style="margin:0 0 24px;line-height:1.55">Solo queda confirmar tu correo para empezar a organizar tus torneos.</p>
      <a href="{{.VerificationURL}}" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#5b4bff;color:#ffffff;text-decoration:none;font-weight:700">Verificar mi cuenta</a>
      <p style="margin:28px 0 0;color:#667085;font-size:13px;line-height:1.5">Si no has creado una cuenta de Fast Tourney, puedes ignorar este correo.</p></td></tr>
    </table>
  </td></tr></table>
</body></html>`))
