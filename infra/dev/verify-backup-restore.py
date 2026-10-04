"""Verify a dev backup in a disposable PostgreSQL with no network or live data mount."""
import argparse
import json
import os
import re
import subprocess
import uuid

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--backup', required=True, help='pgBackRest backup label')
args = parser.parse_args()
if not re.fullmatch(r'\d{8}-\d{6}F(?:_\d{8}-\d{6}[DI])?', args.backup):
    parser.error('invalid pgBackRest backup label')

def run(*args, **kwargs):
    return subprocess.run(args, check=True, text=True, **kwargs)

def output(*args):
    return run(*args, stdout=subprocess.PIPE).stdout.strip()

container = 'tournaments-manager-dev-postgres-1'
metadata = json.loads(output('docker', 'inspect', container))[0]
private_env = dict(item.split('=', 1) for item in metadata['Config']['Env'])
repo = next(m['Source'] for m in metadata['Mounts'] if m['Destination'] == '/var/lib/pgbackrest')
info = json.loads(output('docker','exec','--user','postgres',container,'pgbackrest','--stanza=fasttourney-dev','--output=json','info'))[0]
if info['status']['code'] != 0:
    raise SystemExit('dev backup repository is not healthy')
backup = args.backup
if not any(item['label'] == backup and not item['error'] for item in info['backup']):
    raise SystemExit('requested backup is unavailable or invalid')
query = "SELECT current_database(), pg_is_in_recovery(), (SELECT max(version_id) FROM goose_db_version WHERE is_applied), (SELECT count(*) FROM accounts), (SELECT count(*) FROM tournaments), (SELECT count(*) FROM matches), (SELECT count(*) FROM match_result_changes), (SELECT count(*) FROM legal_account_acceptances);"
expected = output('docker','exec','--user','postgres',container,'psql','-U',private_env['POSTGRES_USER'],'-d',private_env['POSTGRES_DB'],'-At','-v','ON_ERROR_STOP=1','-c',query)
name = 'fasttourney-dev-restore-check-' + uuid.uuid4().hex[:12]
env = os.environ.copy()
for key in ('PGBACKREST_REPO1_CIPHER_PASS','POSTGRES_USER','POSTGRES_DB'):
    env[key] = private_env[key]
run('docker','volume','create',name,stdout=subprocess.DEVNULL)
try:
    script = '''set -eu
chown postgres:postgres /restore
gosu postgres pgbackrest --stanza=fasttourney-dev --pg1-path=/restore --set="$BACKUP_LABEL" --type=immediate --target-action=promote restore
gosu postgres postgres -D /restore -c archive_mode=off -c listen_addresses="" -c unix_socket_directories=/tmp -c port=55432 >/tmp/postgres-restore.log 2>&1 &
postgres_pid=$!
trap 'kill "$postgres_pid" 2>/dev/null || true; wait "$postgres_pid" 2>/dev/null || true' EXIT
attempt=0
until gosu postgres pg_isready -h /tmp -p 55432 -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; do
 attempt=$((attempt+1)); test "$attempt" -lt 60 || { tail -20 /tmp/postgres-restore.log; exit 1; }; sleep 1
done
attempt=0
until test "$(gosu postgres psql -h /tmp -p 55432 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc 'SELECT pg_is_in_recovery()')" = f; do
 attempt=$((attempt+1)); test "$attempt" -lt 60 || exit 1; sleep 1
done
gosu postgres psql -h /tmp -p 55432 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At -v ON_ERROR_STOP=1 -c "$VERIFY_QUERY"
'''
    result = run('docker','run','--rm','--name',name,'--network','none',
       '--env','PGBACKREST_REPO1_CIPHER_PASS','--env','POSTGRES_USER','--env','POSTGRES_DB',
       '--env','BACKUP_LABEL='+backup,'--env','VERIFY_QUERY='+query,
       '--mount','type=bind,src='+repo+',dst=/var/lib/pgbackrest,readonly',
       '--mount','type=volume,src='+name+',dst=/restore',metadata['Image'],
       'sh','-ec',script,env=env,stdout=subprocess.PIPE)
    actual = result.stdout.strip().splitlines()[-1]
    print('live aggregates:', expected)
    print('restored aggregates:', actual)
    if actual.split('|')[1] != 'f':
        raise SystemExit('restored PostgreSQL has not completed recovery')
    if actual != expected:
        raise SystemExit('restored aggregates differ from live dev; check backup age or concurrent writes')
    print('backup:',backup)
    print('database|in_recovery|schema|accounts|tournaments|matches|result_history|legal_acceptances')
    print(actual)
    print('isolated restore matches live aggregates; repository mounted read-only; no network or active data mount')
finally:
    subprocess.run(['docker','rm','-f',name],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    run('docker','volume','rm',name,stdout=subprocess.DEVNULL)
