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

// ErrRefreshInvalid indica un refresh vencido, revocado, usado o inválido.
var ErrRefreshInvalid = errors.New("refresh inválido")

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
	IsUsernameAvailable(context.Context, string) (bool, error)
	VerifyAndCreateSession(context.Context, []byte, []byte, []byte, []byte) (Session, error)
	RotateSessionTokens(context.Context, []byte, []byte, []byte) (Session, error)
}

// Refresh rota un refresh opaco y emite los siguientes tokens de la sesión.
func (s Service) Refresh(ctx context.Context, token string) (Session, string, string, error) {
	access, refresh := make([]byte, 32), make([]byte, 32)
	if _, err := rand.Read(access); err != nil {
		return Session{}, "", "", err
	}
	if _, err := rand.Read(refresh); err != nil {
		return Session{}, "", "", err
	}
	accessToken, refreshToken := base64.RawURLEncoding.EncodeToString(access), base64.RawURLEncoding.EncodeToString(refresh)
	oldHash := sha256.Sum256([]byte("refresh:" + token))
	accessHash := sha256.Sum256([]byte("session:" + accessToken))
	refreshHash := sha256.Sum256([]byte("refresh:" + refreshToken))
	session, err := s.repository.RotateSessionTokens(ctx, oldHash[:], accessHash[:], refreshHash[:])
	if err != nil {
		return Session{}, "", "", ErrRefreshInvalid
	}
	return session, accessToken, refreshToken, nil
}

// Session describe una sesión creada durante la verificación.
type Session struct {
	AccountID, Username             string
	IdleExpiresAt, RefreshExpiresAt string
}

// Mailer entrega el enlace de verificación mediante el adaptador configurado.
type Mailer interface {
	SendVerification(context.Context, string, string) error
}

// Verify consume una verificación y crea una sesión opaca.
func (s Service) Verify(ctx context.Context, token, previousSessionToken string) (Session, string, string, error) {
	verificationHash := sha256.Sum256([]byte("registration-verification:" + token))
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return Session{}, "", "", err
	}
	sessionToken := base64.RawURLEncoding.EncodeToString(secret)
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	refreshSecret := make([]byte, 32)
	if _, err := rand.Read(refreshSecret); err != nil {
		return Session{}, "", "", err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(refreshSecret)
	refreshHash := sha256.Sum256([]byte("refresh:" + refreshToken))
	var previousSessionHash []byte
	if previousSessionToken != "" {
		hash := sha256.Sum256([]byte("session:" + previousSessionToken))
		previousSessionHash = hash[:]
	}
	session, err := s.repository.VerifyAndCreateSession(ctx, verificationHash[:], sessionHash[:], refreshHash[:], previousSessionHash)
	if err != nil {
		return Session{}, "", "", err
	}
	return session, sessionToken, refreshToken, nil
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

// UsernameAvailable consulta la disponibilidad actual sin reservar el nombre.
// El alta sigue siendo la autoridad para garantizar la unicidad bajo concurrencia.
func (s Service) UsernameAvailable(ctx context.Context, username string) (bool, error) {
	return s.repository.IsUsernameAvailable(ctx, username)
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
