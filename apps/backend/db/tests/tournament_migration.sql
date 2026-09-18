-- Run with psql against an isolated test database with the two deployment roles.
-- All fixtures and DDL are rolled back, including the test schema.
\set ON_ERROR_STOP on
BEGIN;
CREATE SCHEMA bracket_migration_fixture AUTHORIZATION tournaments_manager_dev_schema_owner;
SET LOCAL search_path = bracket_migration_fixture;
SET LOCAL ROLE tournaments_manager_dev_schema_owner;
\ir ../schema/initial_schema.sql

INSERT INTO accounts (id,email,locale,state,username,verified_at)
VALUES ('019abcde-1111-7111-8111-111111111111','migration@example.test','es','verified','migration_owner',now());
INSERT INTO leagues (id,organizer_account_id,name,state,round_robin_legs,published_at)
VALUES ('019abcde-2222-7222-8222-222222222222','019abcde-1111-7111-8111-111111111111','Existing league','completed',2,now());
INSERT INTO league_teams (id,league_id,name,name_normalized,position) VALUES
('019abcde-3333-7333-8333-333333333333','019abcde-2222-7222-8222-222222222222','A','a',1),
('019abcde-4444-7444-8444-444444444444','019abcde-2222-7222-8222-222222222222','B','b',2);
INSERT INTO matches(id,league_id,round_number,sequence,home_team_id,away_team_id,state,home_score,away_score)
VALUES ('019abcde-5555-7555-8555-555555555555','019abcde-2222-7222-8222-222222222222',1,1,'019abcde-3333-7333-8333-333333333333','019abcde-4444-7444-8444-444444444444','completed',2,1);
INSERT INTO match_result_changes(match_id,changed_by_account_id,home_score,away_score)
VALUES ('019abcde-5555-7555-8555-555555555555','019abcde-1111-7111-8111-111111111111',2,1);
INSERT INTO league_champions(league_id,team_id)
VALUES ('019abcde-2222-7222-8222-222222222222','019abcde-3333-7333-8333-333333333333');
INSERT INTO account_notifications(account_id,kind,league_id)
VALUES ('019abcde-1111-7111-8111-111111111111','league_administrator_assigned','019abcde-2222-7222-8222-222222222222');

\ir ../migrations/00005_migrate_leagues_to_tournament_stages.sql

DO $$
DECLARE wrong_stage uuid;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM tournament_stages WHERE type='league' AND state='completed' AND round_robin_legs=2) THEN
        RAISE EXCEPTION 'Existing league configuration was not preserved';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM matches m JOIN tournament_stages s ON s.id=m.stage_id AND s.tournament_id=m.tournament_id WHERE m.home_score=2 AND m.away_score=1 AND m.home_source_kind='seeded_team') THEN
        RAISE EXCEPTION 'Existing match was not preserved';
    END IF;
    IF (SELECT count(*) FROM tournament_champions)<>1 OR (SELECT count(*) FROM match_result_changes)<>1 THEN
        RAISE EXCEPTION 'Champion or history was lost';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM account_notifications WHERE kind='tournament_administrator_assigned') THEN
        RAISE EXCEPTION 'Notification was not migrated';
    END IF;
    IF NOT has_table_privilege('tournaments_manager_dev_app','tournament_stages','SELECT,INSERT,UPDATE,DELETE') THEN
        RAISE EXCEPTION 'Runtime cannot access stages';
    END IF;
    INSERT INTO tournaments(id,organizer_account_id,name,published_at)
    VALUES ('019abcde-6666-7666-8666-666666666666','019abcde-1111-7111-8111-111111111111','Other tournament',now());
    INSERT INTO tournament_stages(tournament_id,position,type)
    VALUES ('019abcde-6666-7666-8666-666666666666',1,'single_elimination') RETURNING id INTO wrong_stage;
    BEGIN
        UPDATE matches SET stage_id=wrong_stage;
        RAISE EXCEPTION 'Cross-tournament stage was allowed';
    EXCEPTION WHEN foreign_key_violation THEN NULL;
    END;
    BEGIN
        UPDATE matches SET home_source_kind='winner',home_source_match_id=id;
        RAISE EXCEPTION 'Self-reference was allowed';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        UPDATE matches SET home_source_kind='group_ranking';
        RAISE EXCEPTION 'Unimplemented source was allowed';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        UPDATE matches SET home_score=NULL;
        RAISE EXCEPTION 'Completed match without score was allowed';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
END $$;
ROLLBACK;
