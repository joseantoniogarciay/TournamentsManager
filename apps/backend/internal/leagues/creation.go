package leagues

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
)

var (
	// ErrInvalidLeagueInput indicates invalid creation or start data.
	ErrInvalidLeagueInput = errors.New("liga inválida")
	// ErrLeagueForbidden indicates that the account does not organize the league.
	ErrLeagueForbidden = errors.New("liga no autorizada")
	// ErrLeagueConflict indicates that the start transition is invalid.
	ErrLeagueConflict = errors.New("liga no se puede iniciar")
	// ErrLeagueCancellationConflict indicates that the league cannot be cancelled from its current state.
	ErrLeagueCancellationConflict = errors.New("liga no se puede cancelar")
	// ErrLeagueAdministratorConflict indicates that the requested administrator assignment is invalid.
	ErrLeagueAdministratorConflict = errors.New("administradora de liga inválida")
	// ErrLeagueOwnershipTransferConflict indicates that the requested ownership transfer is invalid.
	ErrLeagueOwnershipTransferConflict = errors.New("transferencia de propiedad inválida")
	// ErrLeagueTeamConflict indicates that the roster cannot accept the requested team.
	ErrLeagueTeamConflict = errors.New("equipo de liga inválido")
	// ErrLeagueWithdrawalConflict indicates that a team cannot be withdrawn now.
	ErrLeagueWithdrawalConflict = errors.New("equipo no se puede retirar")
	// ErrMatchResultForbidden indicates that the account does not administer league results.
	ErrMatchResultForbidden = errors.New("resultado no autorizado")
	// ErrMatchResultConflict indicates that the league cannot accept results in its current state.
	ErrMatchResultConflict = errors.New("resultado no se puede registrar")
	// ErrLeagueCompletionConflict indicates that the league cannot be completed yet.
	ErrLeagueCompletionConflict = errors.New("liga no se puede finalizar")
)

// TeamInput represents a team submitted during creation.
type TeamInput struct{ Name string }

// CreateInput contains the minimum data for a published league.
type CreateInput struct {
	Name  string
	Teams []TeamInput
}

// StartInput defines the rules frozen when the league starts.
type StartInput struct{ RoundRobinLegs int }

// MatchResultInput represents a simple football score.
type MatchResultInput struct{ HomeScore, AwayScore int }

// Team represents a persisted league team.
type Team struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Position  int    `json:"position"`
	Withdrawn bool   `json:"withdrawn"`
}

// Match represents a generated league match.
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

// League is the projection of a visible league.
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

// Standing is a domain-calculated row, not data entered by clients.
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

// CreationRepository persists and queries a league's initial lifecycle.
type CreationRepository interface {
	Create(context.Context, string, CreateInput) (League, error)
	AddTeam(context.Context, string, string, TeamInput) (Team, error)
	RemoveTeam(context.Context, string, string, string) error
	WithdrawTeam(context.Context, string, string, string) (League, error)
	Start(context.Context, string, string, StartInput) (League, error)
	Cancel(context.Context, string, string) (League, error)
	AssignAdministrator(context.Context, string, string, string) error
	ListAdministrators(context.Context, string, string) ([]string, error)
	RemoveAdministrator(context.Context, string, string, string) error
	TransferOwnership(context.Context, string, string, string) error
	RecordResult(context.Context, string, string, string, MatchResultInput) (League, error)
	Complete(context.Context, string, string) (League, error)
	GetPublic(context.Context, string) (League, error)
}

// AddTeam adds a team while the league remains unstarted.
func (s CreationService) AddTeam(ctx context.Context, accountID, leagueID string, input TeamInput) (Team, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 100 {
		return Team{}, ErrInvalidLeagueInput
	}
	return s.repository.AddTeam(ctx, accountID, leagueID, TeamInput{Name: name})
}

// RemoveTeam removes a team only while the roster retains the valid minimum.
func (s CreationService) RemoveTeam(ctx context.Context, accountID, leagueID, teamID string) error {
	return s.repository.RemoveTeam(ctx, accountID, leagueID, teamID)
}

// WithdrawTeam applies the accepted 3-0 rule to every match of a withdrawn team.
func (s CreationService) WithdrawTeam(ctx context.Context, accountID, leagueID, teamID string) (League, error) {
	league, err := s.repository.WithdrawTeam(ctx, accountID, leagueID, teamID)
	return s.withStandings(league), err
}

// CreationService coordinates league creation, retrieval, and start.
type CreationService struct{ repository CreationRepository }

// NewCreationService builds the league-creation use case.
func NewCreationService(repository CreationRepository) CreationService {
	return CreationService{repository: repository}
}

// Create validates and persists a published league without a schedule.
func (s CreationService) Create(ctx context.Context, accountID string, input CreateInput) (League, error) {
	if !validCreateInput(input) {
		return League{}, ErrInvalidLeagueInput
	}
	league, err := s.repository.Create(ctx, accountID, normalizedInput(input))
	return s.withStandings(league), err
}

// Start validates the final configuration and generates the league schedule.
func (s CreationService) Start(ctx context.Context, accountID, leagueID string, input StartInput) (League, error) {
	if input.RoundRobinLegs != 1 && input.RoundRobinLegs != 2 {
		return League{}, ErrInvalidLeagueInput
	}
	league, err := s.repository.Start(ctx, accountID, leagueID, input)
	return s.withStandings(league), err
}

// Cancel keeps a league visible but prevents its sporting lifecycle from continuing.
func (s CreationService) Cancel(ctx context.Context, accountID, leagueID string) (League, error) {
	league, err := s.repository.Cancel(ctx, accountID, leagueID)
	return s.withStandings(league), err
}

// AssignAdministrator directly assigns a verified account identified by username.
func (s CreationService) AssignAdministrator(ctx context.Context, accountID, leagueID, username string) error {
	if !usernamePattern.MatchString(username) {
		return ErrInvalidLeagueInput
	}
	return s.repository.AssignAdministrator(ctx, accountID, leagueID, username)
}

// ListAdministrators returns the delegated accounts the league owner can manage.
func (s CreationService) ListAdministrators(ctx context.Context, accountID, leagueID string) ([]string, error) {
	return s.repository.ListAdministrators(ctx, accountID, leagueID)
}

// RemoveAdministrator removes a delegated administrator identified by username.
func (s CreationService) RemoveAdministrator(ctx context.Context, accountID, leagueID, username string) error {
	if !usernamePattern.MatchString(username) {
		return ErrInvalidLeagueInput
	}
	return s.repository.RemoveAdministrator(ctx, accountID, leagueID, username)
}

// TransferOwnership immediately changes the league organizer to another verified account.
func (s CreationService) TransferOwnership(ctx context.Context, accountID, leagueID, username string) error {
	if !usernamePattern.MatchString(username) {
		return ErrInvalidLeagueInput
	}
	return s.repository.TransferOwnership(ctx, accountID, leagueID, username)
}

// RecordResult immediately applies or corrects a score from an authorized account.
func (s CreationService) RecordResult(ctx context.Context, accountID, leagueID, matchID string, input MatchResultInput) (League, error) {
	if input.HomeScore < 0 || input.AwayScore < 0 {
		return League{}, ErrInvalidLeagueInput
	}
	league, err := s.repository.RecordResult(ctx, accountID, leagueID, matchID, input)
	return s.withStandings(league), err
}

// Complete explicitly closes the league and makes its co-champions official.
func (s CreationService) Complete(ctx context.Context, accountID, leagueID string) (League, error) {
	league, err := s.repository.Complete(ctx, accountID, leagueID)
	return s.withStandings(league), err
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

// GetPublic returns the public projection of a visible league.
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

// CalculateStandings exposes the domain projection for atomic completion operations.
func CalculateStandings(league League) []Standing { return calculateStandings(league) }

func rankTiedStandings(group []Standing, league League) []Standing {
	// The mini-table considers the whole tied group, not just the pair that the
	// algorithm compares. This keeps tiebreaking consistent with three or more teams.
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
