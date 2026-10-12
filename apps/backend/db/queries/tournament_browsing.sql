-- name: ListPersonalMatches :one
WITH candidates AS (
 SELECT m.*,t.name AS tournament_name,t.sport,t.created_at,t.venue AS tournament_venue,s.type AS stage_type,
 h.name AS home_name,a.name AS away_name,
 COALESCE((SELECT max(c.changed_at) FROM match_result_changes c WHERE c.match_id=m.id),t.created_at) AS resolved_at
 FROM tournament_team_accounts own
 JOIN tournament_teams team ON team.id=own.team_id
 JOIN tournaments t ON t.id=own.tournament_id
 JOIN matches m ON m.tournament_id=t.id AND (m.home_team_id=own.team_id OR m.away_team_id=own.team_id)
 JOIN tournament_stages s ON s.id=m.stage_id
 JOIN tournament_teams h ON h.id=m.home_team_id
 JOIN tournament_teams a ON a.id=m.away_team_id
 WHERE own.account_id=sqlc.arg(account_id)::uuid
 AND (m.state='completed' OR (m.state='pending' AND t.state='in_progress' AND s.state='in_progress' AND team.withdrawn_at IS NULL))
), ordered AS (
 SELECT *,row_number() OVER (ORDER BY
 CASE WHEN sqlc.arg(view)::text='pending' THEN NULLIF(scheduled_date,'') END ASC NULLS LAST,
 CASE WHEN sqlc.arg(view)::text='pending' THEN NULLIF(scheduled_time,'') END ASC NULLS LAST,
 CASE WHEN sqlc.arg(view)::text='results' THEN resolved_at END DESC,
 id DESC) AS ordinal FROM candidates
 WHERE state=CASE WHEN sqlc.arg(view)::text='pending' THEN 'pending' ELSE 'completed' END
), selected AS (
 SELECT * FROM ordered WHERE ordinal>COALESCE((SELECT ordinal FROM ordered WHERE id=sqlc.narg(cursor_id)::uuid),0)
 ORDER BY ordinal LIMIT sqlc.arg(page_size)::int
)
SELECT jsonb_build_object('hasMatches',EXISTS(SELECT 1 FROM candidates),'items',COALESCE((
 SELECT jsonb_agg(jsonb_build_object(
 'tournamentId',tournament_id,'tournamentName',tournament_name,'sport',sport,'stageType',stage_type,
 'homeTeamName',home_name,'awayTeamName',away_name,'match',jsonb_strip_nulls(jsonb_build_object(
 'id',id,'stageId',stage_id,'round',round_number,'sequence',sequence,'groupNumber',NULLIF(group_number,0),
 'homeTeamId',home_team_id,'awayTeamId',away_team_id,'homeSourceKind',home_source_kind,'awaySourceKind',away_source_kind,
 'homeSourceMatchId',home_source_match_id,'awaySourceMatchId',away_source_match_id,'winnerTeamId',winner_team_id,
 'state',state,'homeScore',home_score,'awayScore',away_score,'homePenalties',home_penalties,'awayPenalties',away_penalties,
 'resultType',result_type,'incident',incident,'scheduledDate',NULLIF(scheduled_date,''),'scheduledTime',NULLIF(scheduled_time,''),
 'timeZone',NULLIF(time_zone,''),'venue',NULLIF(COALESCE(NULLIF(venue,''),tournament_venue),''),
 'sets',COALESCE((SELECT jsonb_agg(jsonb_build_object('homeScore',ms.home_score,'awayScore',ms.away_score) ORDER BY ms.set_number) FROM match_sets ms WHERE ms.match_id=selected.id),'[]'::jsonb)
 ))) ORDER BY ordinal) FROM selected),'[]'::jsonb)) AS payload;

-- name: ListTournamentEditions :one
WITH RECURSIVE ancestors AS (
 SELECT id,repetition_source_id FROM tournaments WHERE id=sqlc.arg(tournament_id)::uuid
 UNION
 SELECT t.id,t.repetition_source_id FROM tournaments t JOIN ancestors a ON t.id=a.repetition_source_id
), family AS (
 SELECT t.id,t.name,t.state,t.created_at FROM tournaments t JOIN ancestors a ON t.id=a.id WHERE a.repetition_source_id IS NULL
 UNION
 SELECT t.id,t.name,t.state,t.created_at FROM tournaments t JOIN family f ON t.repetition_source_id=f.id
), ordered AS (
 SELECT *,row_number() OVER (ORDER BY created_at DESC,id DESC) AS ordinal FROM family
), selected AS (
 SELECT * FROM ordered WHERE ordinal>COALESCE((SELECT ordinal FROM ordered WHERE id=sqlc.narg(cursor_id)::uuid),0)
 ORDER BY ordinal LIMIT sqlc.arg(page_size)::int
)
SELECT jsonb_build_object('exists',EXISTS(SELECT 1 FROM ancestors),'items',COALESCE((
 SELECT jsonb_agg(jsonb_build_object('id',id,'name',name,'state',state,'createdAt',created_at,'champions',
 COALESCE((SELECT jsonb_agg(team.name ORDER BY team.position) FROM tournament_champions c JOIN tournament_teams team ON team.id=c.team_id WHERE c.tournament_id=selected.id),'[]'::jsonb)) ORDER BY ordinal)
 FROM selected),'[]'::jsonb)) AS payload;
