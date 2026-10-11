package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationInvitationsChooseUnlinkedTeamsAndRespectCapacity(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "claims-owner@example.test", "claimsowner", "correct password")
	first := createVerifiedLocalAccount(ctx, t, pool, "claims-first@example.test", "claimsfirst", "correct password")
	second := createVerifiedLocalAccount(ctx, t, pool, "claims-second@example.test", "claimssecond", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	source, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Normal tournament", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Owner"}, {Name: "Free"}}})
	if err != nil {
		t.Fatal(err)
	}
	token, err := service.CreateTeamInvitation(ctx, owner, source.ID)
	if err != nil {
		t.Fatal(err)
	}
	invitation, err := service.InspectTeamInvitation(ctx, token)
	if err != nil || !invitation.AllowsExistingTeams || len(invitation.AvailableTeamNames) != 1 || invitation.AvailableTeamNames[0] != "Free" {
		t.Fatalf("options: %#v %v", invitation, err)
	}
	var wg sync.WaitGroup
	successes := make(chan tournaments.TeamRegistration, 2)
	failures := make(chan error, 2)
	for _, account := range []string{first, second} {
		wg.Go(func() {
			value, err := service.JoinTeamInvitation(ctx, account, token, tournaments.TeamInput{Name: " free "})
			if err != nil {
				failures <- err
			} else {
				successes <- value
			}
		})
	}
	wg.Wait()
	close(successes)
	close(failures)
	if len(successes) != 1 || len(failures) != 1 {
		t.Fatal("claim race did not produce one winner")
	}
	for value := range successes {
		if value.Team.ID != source.Teams[1].ID {
			t.Fatal("claim created another team")
		}
	}
	for err := range failures {
		if !errors.Is(err, tournaments.ErrTournamentTeamOccupied) {
			t.Fatal(err)
		}
	}
	projection, err := service.GetPublic(ctx, source.ID)
	if err != nil || len(projection.Teams) != 2 {
		t.Fatal("claim changed composition")
	}
	invitation, err = service.InspectTeamInvitation(ctx, token)
	if err != nil || len(invitation.AvailableTeamNames) != 0 {
		t.Fatalf("occupied option remains: %#v %v", invitation, err)
	}
	var losingAccount string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM accounts WHERE id IN ($1,$2) AND NOT EXISTS (SELECT 1 FROM tournament_team_accounts a WHERE a.account_id=accounts.id AND a.tournament_id=$3)`, first, second, source.ID).Scan(&losingAccount); err != nil {
		t.Fatal(err)
	}
	if _, err := service.JoinTeamInvitation(ctx, losingAccount, token, tournaments.TeamInput{Name: "New"}); err != nil {
		t.Fatal(err)
	}
	projection, err = service.GetPublic(ctx, source.ID)
	if err != nil || len(projection.Teams) != 3 {
		t.Fatal("new entrant not added")
	}
	if _, err := service.JoinTeamInvitation(ctx, losingAccount, token, tournaments.TeamInput{Name: "Again"}); !errors.Is(err, tournaments.ErrTournamentInvitationConflict) {
		t.Fatalf("duplicate account: %v", err)
	}
	if _, err := service.JoinTeamInvitation(ctx, owner, token, tournaments.TeamInput{Name: "Extra"}); !errors.Is(err, tournaments.ErrTournamentInvitationConflict) {
		t.Fatalf("owner joined twice: %v", err)
	}
	if err := service.RevokeTeamInvitation(ctx, owner, source.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.JoinTeamInvitation(ctx, losingAccount, token, tournaments.TeamInput{Name: "Anything"}); !errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
		t.Fatalf("revoked: %v", err)
	}
	// Full rosters still accept an account for a pre-existing free team.
	teams := make([]tournaments.TeamInput, 64)
	for i := range teams {
		teams[i].Name = fmt.Sprintf("Team %02d", i)
	}
	full, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Full roster", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	fullToken, err := service.CreateTeamInvitation(ctx, owner, full.ID)
	if err != nil {
		t.Fatal(err)
	}
	invitation, err = service.InspectTeamInvitation(ctx, fullToken)
	if err != nil || len(invitation.AvailableTeamNames) != 63 {
		t.Fatalf("full options: %d %v", len(invitation.AvailableTeamNames), err)
	}
	joined, err := service.JoinTeamInvitation(ctx, first, fullToken, tournaments.TeamInput{Name: "Team 01"})
	if err != nil || joined.Team.ID != full.Teams[1].ID {
		t.Fatalf("full claim: %#v %v", joined, err)
	}
	if _, err := service.JoinTeamInvitation(ctx, second, fullToken, tournaments.TeamInput{Name: "No room"}); !errors.Is(err, tournaments.ErrTournamentInvitationConflict) {
		t.Fatalf("capacity exceeded: %v", err)
	}
	for i := 2; i < 64; i++ {
		var account string
		if err := pool.QueryRow(ctx, `INSERT INTO accounts (email,locale,state,username,verified_at) VALUES ($1,'es','verified',$2,now()) RETURNING id::text`, fmt.Sprintf("claim-%d@example.test", i), fmt.Sprintf("claim%d", i)).Scan(&account); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO tournament_team_accounts (tournament_id,team_id,account_id) VALUES ($1,$2,$3)`, full.ID, full.Teams[i].ID, account); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.CreateTeamInvitation(ctx, owner, full.ID); !errors.Is(err, tournaments.ErrTournamentInvitationConflict) {
		t.Fatalf("full occupied invite: %v", err)
	}
	if _, err := service.InspectTeamInvitation(ctx, fullToken); !errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
		t.Fatalf("full occupied inspection: %v", err)
	}
}
