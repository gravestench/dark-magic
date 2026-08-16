#!/bin/sh

set -eu
. "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)/common.sh"
. "$REALM_SCRIPT_DIR/postgres.sh"

realm_usage() {
    cat <<'EOF'
Usage: scripts/realm/fresh-install.sh [--yes] [--no-start] [--no-build]

Stop Realm, archive its private data/TLS identity, optionally reset an explicitly
authorized local PostgreSQL database, and start a clean Realm.

PostgreSQL reset requires BOTH:
  DARK_MAGIC_REALM_POSTGRES_URL=<local database URL>
  DARK_MAGIC_REALM_ALLOW_DATABASE_RESET=1

The script asks for confirmation unless --yes is supplied. File-backed state is
moved under the runtime backups directory rather than permanently deleted.
EOF
}

realm_yes=0
realm_no_start=0
realm_no_build=0
while [ "$#" -gt 0 ]; do
    case "$1" in
        --yes) realm_yes=1 ;;
        --no-start) realm_no_start=1 ;;
        --no-build) realm_no_build=1 ;;
        -h|--help) realm_usage; exit 0 ;;
        *) realm_usage >&2; realm_fail "unknown argument: $1" ;;
    esac
    shift
done

[ -n "${DARK_MAGIC_REALM_POSTGRES_URL:-}" ] || realm_fail "Realm fresh install requires DARK_MAGIC_REALM_POSTGRES_URL"
realm_assert_safe_data_directory
realm_ensure_layout

if [ "$realm_yes" -ne 1 ]; then
    printf 'Fresh-install Realm data at %s' "$REALM_DATA_DIR"
    if [ -n "${DARK_MAGIC_REALM_POSTGRES_URL:-}" ]; then
        printf ' and the configured local PostgreSQL database'
    fi
    printf '? [y/N] '
    IFS= read -r realm_answer
    case "$realm_answer" in
        y|Y|yes|YES) ;;
        *) realm_say "cancelled"; exit 0 ;;
    esac
fi

"$REALM_SCRIPT_DIR/down.sh"

realm_stamp=$(date -u '+%Y%m%dT%H%M%SZ')
realm_backup="$REALM_BACKUP_DIR/install-$realm_stamp-$$"
mkdir -p "$realm_backup"
chmod 700 "$realm_backup"
if [ -d "$REALM_DATA_DIR" ] && [ -n "$(ls -A "$REALM_DATA_DIR" 2>/dev/null)" ]; then
    mv "$REALM_DATA_DIR" "$realm_backup/data"
    realm_say "archived previous file state at $realm_backup/data"
fi
mkdir -p "$REALM_DATA_DIR"
chmod 700 "$REALM_DATA_DIR"

if [ -n "${DARK_MAGIC_REALM_POSTGRES_URL:-}" ]; then
    [ "${DARK_MAGIC_REALM_ALLOW_DATABASE_RESET:-0}" = 1 ] || \
        realm_fail "refusing PostgreSQL reset without DARK_MAGIC_REALM_ALLOW_DATABASE_RESET=1"
    realm_require_command psql
    realm_postgres_up
    realm_probe=$(psql "$DARK_MAGIC_REALM_POSTGRES_URL" -X -A -t -v ON_ERROR_STOP=1 -c \
        "SELECT current_database() || '|' || CASE WHEN inet_server_addr() IS NULL OR inet_server_addr() <<= '127.0.0.0/8'::inet OR inet_server_addr() = '::1'::inet THEN 'local' ELSE 'remote' END")
    realm_database=${realm_probe%%|*}
    realm_locality=${realm_probe#*|}
    [ "$realm_locality" = local ] || realm_fail "refusing to reset a non-local PostgreSQL server"
    case "$realm_database" in
        ''|postgres|template0|template1) realm_fail "refusing to reset protected PostgreSQL database: $realm_database" ;;
    esac
    realm_postgres_dump "$realm_backup/realm-postgres.dump"
    realm_say "archived PostgreSQL state at $realm_backup/realm-postgres.dump"
    psql "$DARK_MAGIC_REALM_POSTGRES_URL" -X -v ON_ERROR_STOP=1 -1 -c \
        'DROP SCHEMA public CASCADE; CREATE SCHEMA public'
    realm_say "reset local PostgreSQL database $realm_database"
fi

if [ "$realm_no_start" -eq 1 ]; then
    realm_postgres_down
    realm_say "fresh data directory is ready; Realm was not started"
    exit 0
fi
if [ "$realm_no_build" -eq 1 ]; then
    exec "$REALM_SCRIPT_DIR/up.sh" --no-build
fi
exec "$REALM_SCRIPT_DIR/up.sh"
