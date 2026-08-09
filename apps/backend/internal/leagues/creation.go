package leagues

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidLeagueInput indica datos de creación o inicio inválidos.
	ErrInvalidLeagueInput = errors.New("liga inválida")
	// ErrLeagueForbidden indica que la cuenta no organiza la liga.
	ErrLeagueForbidden = errors.New("liga no autorizada")
	// ErrLeagueConflict indica que la transición de inicio no es válida.
	ErrLeagueConflict = errors.New("liga no se puede iniciar")
	// ErrLeagueCancellationConflict indica que la liga no puede cancelarse desde su estado actual.
	ErrLeagueCancellationConflict = errors.New("liga no se puede cancelar")
	// ErrLeagueAdministratorConflict indica que no se puede asignar la administración solicitada.
	ErrLeagueAdministratorConflict = errors.New("administradora de liga inválida")
	// ErrLeagueTeamConflict indica que la composición ya no admite el equipo solicitado.
	ErrLeagueTeamConflict = errors.New("equipo de liga inválido")
	// ErrMatchResultForbidden indica que la cuenta no administra resultados de la liga.
	ErrMatchResultForbidden = errors.New("resultado no autorizado")
	// ErrMatchResultConflict indica que la liga no admite resultados en su estado actual.
	ErrMatchResultConflict = errors.New("resultado no se puede registrar")
)

// TeamInput representa un equipo enviado durante la creación.
type TeamInput struct{ Name string }

// CreateInput reúne los datos mínimos de una liga publicada.
type CreateInput struct {
	Name  string
	Teams []TeamInput
}

// StartInput fija las reglas que se congelan al iniciar.
type StartInput struct{ RoundRobinLegs int }

// MatchResultInput representa el marcador simple de fútbol.
type MatchResultInput struct{ HomeScore, AwayScore int }

// Team representa un equipo persistido de una liga.
type Team struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

// Match representa un partido generado de una liga.
type Match struct {
	ID          string `json:"id"`
	RoundNumber int    `json:"round"`
	Sequence    int    `json:"sequence"`
	HomeTeamID  string `json:"homeTeamId"`
	AwayTeamID  string `json:"awayTeamId"`
	State       string `json:"state"`
	HomeScore   *int   `json:"homeScore,omitempty"`
	AwayScore   *int   `json:"awayScore,omitempty"`
}

// League es la proyección de una liga visible.
type League struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Sport   string  `json:"sport"`
	Format  string  `json:"format"`
	State   string  `json:"state"`
	Teams   []Team  `json:"teams"`
	Matches []Match `json:"matches"`
}

// CreationRepository persiste y consulta el ciclo inicial de una liga.
type CreationRepository interface {
	Create(context.Context, string, CreateInput) (League, error)
	AddTeam(context.Context, string, string, TeamInput) (Team, error)
	RemoveTeam(context.Context, string, string, string) error
	Start(context.Context, string, string, StartInput) (League, error)
	Cancel(context.Context, string, string) (League, error)
	AssignAdministrator(context.Context, string, string, string) error
	RecordResult(context.Context, string, string, string, MatchResultInput) (League, error)
	GetPublic(context.Context, string) (League, error)
}

// AddTeam incorpora un equipo mientras la liga siga sin empezar.
func (s CreationService) AddTeam(ctx context.Context, accountID, leagueID string, input TeamInput) (Team, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 100 {
		return Team{}, ErrInvalidLeagueInput
	}
	return s.repository.AddTeam(ctx, accountID, leagueID, TeamInput{Name: name})
}

// RemoveTeam elimina un equipo solo mientras la composición conserve el mínimo válido.
func (s CreationService) RemoveTeam(ctx context.Context, accountID, leagueID, teamID string) error {
	return s.repository.RemoveTeam(ctx, accountID, leagueID, teamID)
}

// CreationService coordina creación, consulta e inicio de ligas.
type CreationService struct{ repository CreationRepository }

// NewCreationService construye el caso de uso de creación de ligas.
func NewCreationService(repository CreationRepository) CreationService {
	return CreationService{repository: repository}
}

// Create valida y persiste una liga publicada sin calendario.
func (s CreationService) Create(ctx context.Context, accountID string, input CreateInput) (League, error) {
	if !validCreateInput(input) {
		return League{}, ErrInvalidLeagueInput
	}
	return s.repository.Create(ctx, accountID, normalizedInput(input))
}

// Start valida la configuración final y genera el calendario de la liga.
func (s CreationService) Start(ctx context.Context, accountID, leagueID string, input StartInput) (League, error) {
	if input.RoundRobinLegs != 1 && input.RoundRobinLegs != 2 {
		return League{}, ErrInvalidLeagueInput
	}
	return s.repository.Start(ctx, accountID, leagueID, input)
}

// Cancel conserva una liga visible, pero impide que continúe su ciclo deportivo.
func (s CreationService) Cancel(ctx context.Context, accountID, leagueID string) (League, error) {
	return s.repository.Cancel(ctx, accountID, leagueID)
}

// AssignAdministrator asigna directamente una cuenta verificada identificada por username.
func (s CreationService) AssignAdministrator(ctx context.Context, accountID, leagueID, username string) error {
	if !usernamePattern.MatchString(username) {
		return ErrInvalidLeagueInput
	}
	return s.repository.AssignAdministrator(ctx, accountID, leagueID, username)
}

// RecordResult aplica o corrige inmediatamente un marcador de una cuenta autorizada.
func (s CreationService) RecordResult(ctx context.Context, accountID, leagueID, matchID string, input MatchResultInput) (League, error) {
	if input.HomeScore < 0 || input.AwayScore < 0 {
		return League{}, ErrInvalidLeagueInput
	}
	return s.repository.RecordResult(ctx, accountID, leagueID, matchID, input)
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

// GetPublic devuelve la proyección pública de una liga visible.
func (s CreationService) GetPublic(ctx context.Context, leagueID string) (League, error) {
	return s.repository.GetPublic(ctx, leagueID)
}
func validCreateInput(input CreateInput) bool {
	if len(strings.TrimSpace(input.Name)) == 0 || len(input.Name) > 140 || len(input.Teams) < 2 || len(input.Teams) > 64 {
		return false
	}
	seen := map[string]bool{}
	for _, team := range input.Teams {
		name := strings.TrimSpace(team.Name)
		if name == "" || len(name) > 100 || seen[strings.ToLower(name)] {
			return false
		}
		seen[strings.ToLower(name)] = true
	}
	return true
}
func normalizedInput(input CreateInput) CreateInput {
	input.Name = strings.TrimSpace(input.Name)
	for i := range input.Teams {
		input.Teams[i].Name = strings.TrimSpace(input.Teams[i].Name)
	}
	return input
}
