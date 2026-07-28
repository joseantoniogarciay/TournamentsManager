// Package smtp entrega correo mediante SMTP.
package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"net/url"
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
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return Mailer{}, fmt.Errorf("URL pública inválida")
	}
	return Mailer{address: address, from: from, baseURL: parsedURL}, nil
}

// SendVerification entrega un enlace que el próximo corte consumirá mediante API.
func (m Mailer) SendVerification(ctx context.Context, recipient, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	verificationURL := *m.baseURL
	verificationURL.Path = "/verify-registration"
	query := verificationURL.Query()
	query.Set("token", token)
	verificationURL.RawQuery = query.Encode()

	message := []byte("To: " + recipient + "\r\n" +
		"From: " + m.from + "\r\n" +
		"Subject: Verifica tu cuenta de TournamentsManager\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		"Verifica tu cuenta: " + verificationURL.String() + "\r\n")
	if err := smtp.SendMail(m.address, nil, m.from, []string{recipient}, message); err != nil {
		return fmt.Errorf("SMTP: %w", err)
	}
	return nil
}
