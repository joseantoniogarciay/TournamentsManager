// Package registration contiene el caso de uso de alta local.
package registration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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

// ErrPasswordResetInvalid indica un token de restablecimiento inválido, vencido o consumido.
var ErrPasswordResetInvalid = errors.New("restablecimiento inválido")

// ErrLoginInvalid no distingue email, contraseña ni estado para evitar enumerar cuentas.
var ErrLoginInvalid = errors.New("credenciales inválidas")

const (
	passwordMemory      = 19 * 1024
	passwordIterations  = 2
	passwordParallelism = 1
	passwordKeyLength   = 32
)

// Locale identifica una preferencia de idioma admitida para la cuenta.
type Locale string

const (
	// LocaleSpanish representa español.
	LocaleSpanish Locale = "es"
	// LocaleEnglish representa inglés.
	LocaleEnglish Locale = "en"
	// LocaleItalian representa italiano.
	LocaleItalian Locale = "it"
	// LocaleFrench representa francés.
	LocaleFrench Locale = "fr"
)

// Input es la información validada que recibe el caso de uso.
type Input struct {
	Email    string
	Locale   Locale
	Username string
	Password string
	Draft    *Draft
}

// Draft representa un borrador completo que cruza la frontera del alta.
type Draft struct {
	Name  string
	Teams []string
}

// Repository persiste la cuenta pendiente, su credencial y su verificación.
type Repository interface {
	CreatePending(context.Context, Input, string, []byte) (bool, error)
	IsUsernameAvailable(context.Context, string) (bool, error)
	SearchUsernames(context.Context, string) ([]string, error)
	VerifyAndCreateSession(context.Context, []byte, []byte, []byte, []byte) (Session, error)
	RotateSessionTokens(context.Context, []byte, []byte, []byte) (Session, error)
	CreatePasswordReset(context.Context, string, []byte) (string, Locale, bool, error)
	InspectPasswordReset(context.Context, []byte) (string, error)
	ConsumePasswordReset(context.Context, []byte, string, []byte, []byte) (Session, error)
	FindLocalAccountForLogin(context.Context, string) (LocalAccount, error)
	CreateLocalLoginSession(context.Context, string, []byte, []byte) (Session, error)
	RenewLoginVerification(context.Context, string, []byte) (string, Locale, error)
}

// LocalAccount reúne los datos de una cuenta necesarios para la autenticación local.
type LocalAccount struct {
	ID, Email, Username, PasswordHash string
	Locale                            Locale
	Verified                          bool
}

// LoginResult comunica una sesión creada o la renovación de una verificación pendiente.
type LoginResult struct {
	Session Session
	Pending bool
	Access  string
	Refresh string
}

// Login verifica una credencial local y crea una sesión, o reenvía la verificación pendiente.
func (s Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	account, err := s.repository.FindLocalAccountForLogin(ctx, strings.TrimSpace(email))
	if err != nil || !VerifyPassword(password, account.PasswordHash) {
		return LoginResult{}, ErrLoginInvalid
	}
	if !account.Verified {
		token, hash, err := newVerificationToken()
		if err != nil {
			return LoginResult{}, err
		}
		recipient, locale, err := s.repository.RenewLoginVerification(ctx, account.ID, hash)
		if err != nil {
			return LoginResult{}, err
		}
		if err := s.mailer.SendVerification(ctx, recipient, locale, token); err != nil {
			return LoginResult{}, err
		}
		return LoginResult{Pending: true}, nil
	}
	access, refresh := make([]byte, 32), make([]byte, 32)
	if _, err := rand.Read(access); err != nil {
		return LoginResult{}, err
	}
	if _, err := rand.Read(refresh); err != nil {
		return LoginResult{}, err
	}
	accessToken, refreshToken := base64.RawURLEncoding.EncodeToString(access), base64.RawURLEncoding.EncodeToString(refresh)
	accessHash := sha256.Sum256([]byte("session:" + accessToken))
	refreshHash := sha256.Sum256([]byte("refresh:" + refreshToken))
	session, err := s.repository.CreateLocalLoginSession(ctx, account.ID, accessHash[:], refreshHash[:])
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Session: session, Access: accessToken, Refresh: refreshToken}, nil
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
	SendVerification(context.Context, string, Locale, string) error
	SendPasswordReset(context.Context, string, Locale, string) error
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

// SearchUsernames devuelve usernames públicos de cuentas verificadas.
func (s Service) SearchUsernames(ctx context.Context, query string) ([]string, error) {
	return s.repository.SearchUsernames(ctx, query)
}

// Register crea una cuenta pendiente. La respuesta no diferencia un email ya
// existente para no convertir el endpoint en un oráculo de cuentas.
func (s Service) Register(ctx context.Context, input Input) error {
	passwordHash, err := HashPassword(input.Password)
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
	if err := s.mailer.SendVerification(ctx, input.Email, input.Locale, token); err != nil {
		return fmt.Errorf("enviar verificación: %w", err)
	}
	return nil
}

// RequestPasswordReset crea y entrega un enlace sin revelar si el email existe.
func (s Service) RequestPasswordReset(ctx context.Context, email string) error {
	token, hash, err := newPasswordResetToken()
	if err != nil {
		return err
	}
	recipient, locale, created, err := s.repository.CreatePasswordReset(ctx, strings.TrimSpace(email), hash)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	return s.mailer.SendPasswordReset(ctx, recipient, locale, token)
}

// InspectPasswordReset obtiene el email de un enlace válido sin consumirlo.
func (s Service) InspectPasswordReset(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte("password-reset:" + token))
	email, err := s.repository.InspectPasswordReset(ctx, hash[:])
	if err != nil {
		return "", ErrPasswordResetInvalid
	}
	return email, nil
}

// ResetPassword consume el enlace, cambia la credencial y emite una sesión nueva.
func (s Service) ResetPassword(ctx context.Context, token, password string) (Session, string, string, error) {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return Session{}, "", "", err
	}
	resetHash := sha256.Sum256([]byte("password-reset:" + token))
	access, refresh := make([]byte, 32), make([]byte, 32)
	if _, err := rand.Read(access); err != nil {
		return Session{}, "", "", err
	}
	if _, err := rand.Read(refresh); err != nil {
		return Session{}, "", "", err
	}
	accessToken, refreshToken := base64.RawURLEncoding.EncodeToString(access), base64.RawURLEncoding.EncodeToString(refresh)
	accessHash, refreshHash := sha256.Sum256([]byte("session:"+accessToken)), sha256.Sum256([]byte("refresh:"+refreshToken))
	session, err := s.repository.ConsumePasswordReset(ctx, resetHash[:], passwordHash, accessHash[:], refreshHash[:])
	if err != nil {
		return Session{}, "", "", ErrPasswordResetInvalid
	}
	return session, accessToken, refreshToken, nil
}

// HashPassword crea un verificador Argon2id con los parámetros vigentes.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", passwordMemory, passwordIterations, passwordParallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

// VerifyPassword compara una contraseña con un verificador Argon2id vigente.
// Un formato desconocido se considera una credencial inválida, no un error interno.
func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil || memory == 0 || iterations == 0 || parallelism == 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) != int(passwordKeyLength) {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, passwordKeyLength)
	return subtle.ConstantTimeCompare(actual, expected) == 1
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

func newPasswordResetToken() (string, []byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(secret)
	hash := sha256.Sum256([]byte("password-reset:" + token))
	return token, hash[:], nil
}

// NormalizeInput aplica las normalizaciones aceptadas antes de persistir.
func NormalizeInput(input Input) Input {
	input.Email = strings.TrimSpace(input.Email)
	input.Username = strings.TrimSpace(input.Username)
	if input.Draft != nil {
		input.Draft.Name = strings.TrimSpace(input.Draft.Name)
		for index := range input.Draft.Teams {
			input.Draft.Teams[index] = strings.TrimSpace(input.Draft.Teams[index])
		}
	}
	return input
}

// IsSupportedLocale indica si el locale puede persistirse y seleccionar contenido localizado.
func IsSupportedLocale(locale Locale) bool {
	switch locale {
	case LocaleSpanish, LocaleEnglish, LocaleItalian, LocaleFrench:
		return true
	default:
		return false
	}
}
