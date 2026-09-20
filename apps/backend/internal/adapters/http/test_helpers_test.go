package http

import (
	"context"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

type testAuthenticator struct{ accountID string }

type testDeletionAuthenticator struct{ testAuthenticator }

func (testDeletionAuthenticator) ScheduleAccountDeletion(context.Context, string) (time.Time, error) {
	return time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC), nil
}

var testAllowedOrigins = []string{"http://localhost:8082"}

func (a testAuthenticator) Authenticate(context.Context, string) (string, error) {
	if a.accountID == "" {
		return "", tournaments.ErrUnauthenticated
	}
	return a.accountID, nil
}

func (a testAuthenticator) GetCurrentSession(context.Context, string) (tournaments.CurrentSession, error) {
	if a.accountID == "" {
		return tournaments.CurrentSession{}, tournaments.ErrUnauthenticated
	}
	return tournaments.CurrentSession{
		AccountID:         a.accountID,
		Username:          "person",
		LastTeamName:      "Barrio Norte",
		IdleExpiresAt:     "2026-08-09T12:00:00Z",
		AbsoluteExpiresAt: "2026-08-09T12:00:00Z",
	}, nil
}

func (testAuthenticator) RevokeSession(context.Context, string) error { return nil }

type testTournamentRepository struct {
	items         []tournaments.Item
	recentItems   []tournaments.Item
	followVisible bool
}

func (r testTournamentRepository) List(_ context.Context, _ string, _ tournaments.Relationship, _ string, _ int) ([]tournaments.Item, error) {
	return r.items, nil
}

func (r testTournamentRepository) ListRecent(context.Context, string) ([]tournaments.Item, error) {
	return r.recentItems, nil
}

func (r testTournamentRepository) Follow(_ context.Context, _ string, _ string) (bool, error) {
	return r.followVisible, nil
}

func (r testTournamentRepository) Unfollow(context.Context, string, string) error { return nil }

type testCreationRepository struct {
	created           tournaments.Tournament
	started           tournaments.Tournament
	startErr          error
	elimination       tournaments.Tournament
	eliminationErr    error
	administrators    []string
	administratorsErr error
	cancelled         tournaments.Tournament
	cancelErr         error
	team              tournaments.Team
	teamErr           error
	removeErr         error
	transferErr       error
	withdrawn         tournaments.Tournament
	withdrawErr       error
	result            tournaments.Tournament
	resultErr         error
	completed         tournaments.Tournament
	completeErr       error
	invitation        tournaments.TeamInvitation
	invitationErr     error
	registration      tournaments.TeamRegistration
	registrationErr   error
	revokeInviteErr   error
	saveInviteErr     error
}

func (r testCreationRepository) Create(context.Context, string, tournaments.CreateInput) (tournaments.Tournament, error) {
	return r.created, nil
}

func (r testCreationRepository) AddTeam(context.Context, string, string, tournaments.TeamInput) (tournaments.Team, error) {
	return r.team, r.teamErr
}

func (r testCreationRepository) RemoveTeam(context.Context, string, string, string) error {
	return r.removeErr
}

func (r testCreationRepository) SaveTeamInvitation(context.Context, string, string, tournaments.InvitationTokenHash) error {
	return r.saveInviteErr
}

func (r testCreationRepository) RevokeTeamInvitation(context.Context, string, string) error {
	return r.revokeInviteErr
}

func (r testCreationRepository) InspectTeamInvitation(context.Context, tournaments.InvitationTokenHash) (tournaments.TeamInvitation, error) {
	return r.invitation, r.invitationErr
}

func (r testCreationRepository) JoinTeamInvitation(context.Context, string, tournaments.InvitationTokenHash, tournaments.TeamInput) (tournaments.TeamRegistration, error) {
	return r.registration, r.registrationErr
}

func (r testCreationRepository) WithdrawTeam(context.Context, string, string, string) (tournaments.Tournament, error) {
	return r.withdrawn, r.withdrawErr
}

func (r testCreationRepository) Start(context.Context, string, string, tournaments.StartInput) (tournaments.Tournament, error) {
	return r.started, r.startErr
}

func (r testCreationRepository) StartElimination(context.Context, string, string) (tournaments.Tournament, error) {
	return r.elimination, r.eliminationErr
}

func (r testCreationRepository) Cancel(context.Context, string, string) (tournaments.Tournament, error) {
	return r.cancelled, r.cancelErr
}

func (testCreationRepository) AssignAdministrator(context.Context, string, string, string) error {
	return nil
}

func (r testCreationRepository) ListAdministrators(context.Context, string, string) ([]string, error) {
	return r.administrators, r.administratorsErr
}

func (r testCreationRepository) RemoveAdministrator(context.Context, string, string, string) error {
	return r.removeErr
}

func (r testCreationRepository) TransferOwnership(context.Context, string, string, string) error {
	return r.transferErr
}

func (r testCreationRepository) RecordResult(context.Context, string, string, string, tournaments.MatchResultInput) (tournaments.Tournament, error) {
	return r.result, r.resultErr
}

func (r testCreationRepository) Complete(context.Context, string, string) (tournaments.Tournament, error) {
	return r.completed, r.completeErr
}

func (testCreationRepository) GetPublic(context.Context, string) (tournaments.Tournament, error) {
	return tournaments.Tournament{}, nil
}

type testRegistrationRepository struct {
	available    bool
	loginAccount registration.LocalAccount
	loginDraft   **registration.Draft
	loginError   error
	loginSession registration.Session
}

func (r testRegistrationRepository) CreatePending(context.Context, registration.Input, string, []byte) (bool, error) {
	return false, nil
}

func (r testRegistrationRepository) IsUsernameAvailable(context.Context, string) (bool, error) {
	return r.available, nil
}

func (testRegistrationRepository) SearchUsernames(context.Context, string) ([]string, error) {
	return []string{}, nil
}

func (r testRegistrationRepository) VerifyAndCreateSession(context.Context, []byte, []byte, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrVerificationInvalid
}

func (r testRegistrationRepository) RotateSessionTokens(context.Context, []byte, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrRefreshInvalid
}
func (r testRegistrationRepository) CreatePasswordReset(context.Context, string, []byte) (string, registration.Locale, bool, error) {
	return "", "", false, nil
}
func (r testRegistrationRepository) InspectPasswordReset(context.Context, []byte) (string, error) {
	return "", registration.ErrPasswordResetInvalid
}
func (r testRegistrationRepository) ConsumePasswordReset(context.Context, []byte, string, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrPasswordResetInvalid
}
func (r testRegistrationRepository) FindLocalAccountForLogin(context.Context, string) (registration.LocalAccount, error) {
	if r.loginAccount.ID == "" {
		return registration.LocalAccount{}, registration.ErrLoginInvalid
	}
	return r.loginAccount, nil
}
func (r testRegistrationRepository) CreateLocalLoginSession(_ context.Context, _ string, _, _ []byte, draft *registration.Draft) (registration.Session, error) {
	if r.loginDraft != nil {
		*r.loginDraft = draft
	}
	return r.loginSession, r.loginError
}

type testFederatedRepository struct {
	challengeErr        error
	authenticateErr     error
	authenticateDraft   **federated.Draft
	addWithTicketErr    error
	removeWithTicketErr error
	reauthenticationErr error
}

type testGoogleVerifier struct{ identity federated.Identity }

func (v testGoogleVerifier) Verify(context.Context, string) (federated.Identity, error) {
	return v.identity, nil
}

func (r testFederatedRepository) CreateChallenge(context.Context, []byte, time.Time) (string, error) {
	return "019abcde-1111-7111-8111-111111111111", r.challengeErr
}
func (r testFederatedRepository) AuthenticateGoogle(_ context.Context, _ string, _ []byte, _ federated.Identity, _ *federated.Registration, draft *federated.Draft, _, _ []byte) (federated.Session, error) {
	if r.authenticateDraft != nil {
		*r.authenticateDraft = draft
	}
	return federated.Session{}, r.authenticateErr
}
func (testFederatedRepository) AddGoogleIdentity(context.Context, string, string, []byte, federated.Identity) error {
	return nil
}
func (r testFederatedRepository) ReauthenticateGoogle(context.Context, string, string, string, []byte, federated.Identity, []byte) error {
	return r.reauthenticationErr
}
