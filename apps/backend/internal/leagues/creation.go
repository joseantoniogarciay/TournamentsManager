package leagues

import (
	"context"
	"errors"
	"regexp"
	"sort"
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
	// ErrLeagueCompletionConflict indica que la liga no puede cerrarse todavía.
	ErrLeagueCompletionConflict = errors.New("liga no se puede finalizar")
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
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Sport           string     `json:"sport"`
	Format          string     `json:"format"`
	State           string     `json:"state"`
	RoundRobinLegs  int        `json:"roundRobinLegs"`
	Teams           []Team     `json:"teams"`
	Matches         []Match    `json:"matches"`
	Standings       []Standing `json:"standings"`
	ChampionTeamIDs []string   `json:"championTeamIds"`
}

// Standing es una fila calculada por el dominio, no un dato introducido por la clientela.
type Standing struct {
	Position       int    `json:"position"`
	TeamID         string `json:"teamId"`
	Played         int    `json:"played"`
	Won            int    `json:"won"`
	Drawn          int    `json:"drawn"`
	Lost           int    `json:"lost"`
	GoalsFor       int    `json:"goalsFor"`
	GoalsAgainst   int    `json:"goalsAgainst"`
	GoalDifference int    `json:"goalDifference"`
	Points         int    `json:"points"`
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
	Complete(context.Context, string, string) (League, error)
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
	league, err := s.repository.Create(ctx, accountID, normalizedInput(input))
	return s.withStandings(league), err
}

// Start valida la configuración final y genera el calendario de la liga.
func (s CreationService) Start(ctx context.Context, accountID, leagueID string, input StartInput) (League, error) {
	if input.RoundRobinLegs != 1 && input.RoundRobinLegs != 2 {
		return League{}, ErrInvalidLeagueInput
	}
	league, err := s.repository.Start(ctx, accountID, leagueID, input)
	return s.withStandings(league), err
}

// Cancel conserva una liga visible, pero impide que continúe su ciclo deportivo.
func (s CreationService) Cancel(ctx context.Context, accountID, leagueID string) (League, error) {
	league, err := s.repository.Cancel(ctx, accountID, leagueID)
	return s.withStandings(league), err
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
	league, err := s.repository.RecordResult(ctx, accountID, leagueID, matchID, input)
	return s.withStandings(league), err
}

// Complete cierra explícitamente la liga y hace oficiales sus co-campeones.
func (s CreationService) Complete(ctx context.Context, accountID, leagueID string) (League, error) {
	league, err := s.repository.Complete(ctx, accountID, leagueID)
	return s.withStandings(league), err
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

// GetPublic devuelve la proyección pública de una liga visible.
func (s CreationService) GetPublic(ctx context.Context, leagueID string) (League, error) {
	league, err := s.repository.GetPublic(ctx, leagueID)
	return s.withStandings(league), err
}

func (s CreationService) withStandings(league League) League {
	if league.State != "in_progress" && league.State != "completed" {
		league.Standings = []Standing{}
		return league
	}
	league.Standings = calculateStandings(league)
	return league
}

func calculateStandings(league League) []Standing {
	byTeam := make(map[string]*Standing, len(league.Teams))
	for _, team := range league.Teams {
		byTeam[team.ID] = &Standing{TeamID: team.ID}
	}
	for _, match := range league.Matches {
		if match.State != "completed" || match.HomeScore == nil || match.AwayScore == nil {
			continue
		}
		home, away := byTeam[match.HomeTeamID], byTeam[match.AwayTeamID]
		if home == nil || away == nil {
			continue
		}
		home.Played, away.Played = home.Played+1, away.Played+1
		home.GoalsFor, home.GoalsAgainst = home.GoalsFor+*match.HomeScore, home.GoalsAgainst+*match.AwayScore
		away.GoalsFor, away.GoalsAgainst = away.GoalsFor+*match.AwayScore, away.GoalsAgainst+*match.HomeScore
		switch {
		case *match.HomeScore > *match.AwayScore:
			home.Won, home.Points, away.Lost = home.Won+1, home.Points+3, away.Lost+1
		case *match.HomeScore < *match.AwayScore:
			away.Won, away.Points, home.Lost = away.Won+1, away.Points+3, home.Lost+1
		default:
			home.Drawn, home.Points, away.Drawn, away.Points = home.Drawn+1, home.Points+1, away.Drawn+1, away.Points+1
		}
	}
	standings := make([]Standing, 0, len(league.Teams))
	for _, team := range league.Teams {
		standing := byTeam[team.ID]
		if standing == nil {
			continue
		}
		standing.GoalDifference = standing.GoalsFor - standing.GoalsAgainst
		standings = append(standings, *standing)
	}
	groups := map[int][]Standing{}
	for _, standing := range standings {
		groups[standing.Points] = append(groups[standing.Points], standing)
	}
	points := make([]int, 0, len(groups))
	for pointTotal := range groups {
		points = append(points, pointTotal)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(points)))
	ordered := make([]Standing, 0, len(standings))
	for _, pointTotal := range points {
		group := rankTiedStandings(groups[pointTotal], league)
		base := len(ordered)
		for i := range group {
			group[i].Position += base
		}
		ordered = append(ordered, group...)
	}
	return ordered
}

// CalculateStandings expone la proyección de dominio para operaciones atómicas de cierre.
func CalculateStandings(league League) []Standing { return calculateStandings(league) }

func rankTiedStandings(group []Standing, league League) []Standing {
	// La mini-clasificación considera a todo el grupo empatado, no solo a la
	// pareja que el algoritmo esté comparando. Así el desempate es consistente
	// también con tres o más equipos.
	head := headToHead(group, league.Matches)
	sort.SliceStable(group, func(i, j int) bool {
		left, right := group[i], group[j]
		leftHead, rightHead := head[left.TeamID], head[right.TeamID]
		if league.RoundRobinLegs == 2 {
			if comparison := compareStanding(leftHead, rightHead); comparison != 0 {
				return comparison > 0
			}
			if comparison := compareGeneralStanding(left, right); comparison != 0 {
				return comparison > 0
			}
			return false
		}
		if comparison := compareGeneralStanding(left, right); comparison != 0 {
			return comparison > 0
		}
		return compareStanding(leftHead, rightHead) > 0
	})
	position := 1
	for i := range group {
		if i > 0 && !sameTiedStandingRank(group[i-1], group[i], head, league.RoundRobinLegs) {
			position = i + 1
		}
		group[i].Position = position
	}
	return group
}

func headToHead(group []Standing, matches []Match) map[string]Standing {
	result := make(map[string]Standing, len(group))
	inGroup := make(map[string]bool, len(group))
	for _, standing := range group {
		result[standing.TeamID] = Standing{TeamID: standing.TeamID}
		inGroup[standing.TeamID] = true
	}
	for _, match := range matches {
		if match.State != "completed" || match.HomeScore == nil || match.AwayScore == nil || !inGroup[match.HomeTeamID] || !inGroup[match.AwayTeamID] {
			continue
		}
		home, away := result[match.HomeTeamID], result[match.AwayTeamID]
		home.GoalsFor, home.GoalsAgainst = home.GoalsFor+*match.HomeScore, home.GoalsAgainst+*match.AwayScore
		away.GoalsFor, away.GoalsAgainst = away.GoalsFor+*match.AwayScore, away.GoalsAgainst+*match.HomeScore
		switch {
		case *match.HomeScore > *match.AwayScore:
			home.Points += 3
		case *match.HomeScore < *match.AwayScore:
			away.Points += 3
		default:
			home.Points++
			away.Points++
		}
		home.GoalDifference, away.GoalDifference = home.GoalsFor-home.GoalsAgainst, away.GoalsFor-away.GoalsAgainst
		result[match.HomeTeamID], result[match.AwayTeamID] = home, away
	}
	return result
}

func compareStanding(left, right Standing) int {
	for _, pair := range [][2]int{{left.Points, right.Points}, {left.GoalDifference, right.GoalDifference}, {left.GoalsFor, right.GoalsFor}} {
		if pair[0] != pair[1] {
			return pair[0] - pair[1]
		}
	}
	return 0
}

func compareGeneralStanding(left, right Standing) int {
	if left.GoalDifference != right.GoalDifference {
		return left.GoalDifference - right.GoalDifference
	}
	return left.GoalsFor - right.GoalsFor
}

func sameTiedStandingRank(left, right Standing, head map[string]Standing, legs int) bool {
	if legs == 2 {
		return compareStanding(head[left.TeamID], head[right.TeamID]) == 0 && compareGeneralStanding(left, right) == 0
	}
	return compareGeneralStanding(left, right) == 0 && compareStanding(head[left.TeamID], head[right.TeamID]) == 0
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
