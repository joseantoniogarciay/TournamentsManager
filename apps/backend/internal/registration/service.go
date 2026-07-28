// Package registration contiene el caso de uso de alta local.
package registration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrVerificationInvalid indica un token vencido, usado o inválido.
var ErrVerificationInvalid = errors.New("verificación inválida")

const (
	passwordMemory      = 19 * 1024
	passwordIterations  = 2
	passwordParallelism = 1
	passwordKeyLength   = 32
)

// Input es la información validada que recibe el caso de uso.
type Input struct {
	Email    string
	Username string
	Password string
}

// Repository persiste la cuenta pendiente, su credencial y su verificación.
type Repository interface {
	CreatePending(context.Context, Input, string, []byte) (bool, error)
	VerifyAndCreateSession(context.Context, []byte, []byte) (Session, error)
}

// Session describe una sesión creada durante la verificación.
type Session struct {
	AccountID, Username string
	IdleExpiresAt       string
}

// Mailer entrega el enlace de verificación mediante el adaptador configurado.
type Mailer interface {
	SendVerification(context.Context, string, string) error
}

// Verify consume una verificación y crea una sesión opaca.
func (s Service) Verify(ctx context.Context, token string) (Session, string, error) {
	verificationHash := sha256.Sum256([]byte("registration-verification:" + token))
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return Session{}, "", err
	}
	sessionToken := base64.RawURLEncoding.EncodeToString(secret)
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	session, err := s.repository.VerifyAndCreateSession(ctx, verificationHash[:], sessionHash[:])
	if err != nil {
		return Session{}, "", err
	}
	return session, sessionToken, nil
}

// Service coordina el alta sin revelar si un email ya está registrado.
type Service struct {
	repository Repository
	mailer     Mailer
}

// NewService construye el caso de uso con sus puertos explícitos.
func NewService(repository Repository, mailer Mailer) Service {
	return Service{repository: repository, mailer: mailer}
}

// Register crea una cuenta pendiente. La respuesta no diferencia un email ya
// existente para no convertir el endpoint en un oráculo de cuentas.
func (s Service) Register(ctx context.Context, input Input) error {
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("generar hash de contraseña: %w", err)
	}

	token, tokenHash, err := newVerificationToken()
	if err != nil {
		return fmt.Errorf("generar token de verificación: %w", err)
	}

	created, err := s.repository.CreatePending(ctx, input, passwordHash, tokenHash)
	if err != nil {
		return fmt.Errorf("crear cuenta pendiente: %w", err)
	}
	if !created {
		return nil
	}
	if err := s.mailer.SendVerification(ctx, input.Email, token); err != nil {
		return fmt.Errorf("enviar verificación: %w", err)
	}
	return nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", passwordMemory, passwordIterations, passwordParallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func newVerificationToken() (string, []byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(secret)
	hash := sha256.Sum256([]byte("registration-verification:" + token))
	return token, hash[:], nil
}

// NormalizeInput aplica las normalizaciones aceptadas antes de persistir.
func NormalizeInput(input Input) Input {
	input.Email = strings.TrimSpace(input.Email)
	input.Username = strings.TrimSpace(input.Username)
	return input
}
