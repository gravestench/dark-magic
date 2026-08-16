#!/bin/sh

# Functions for the isolated PostgreSQL development dependency. Callers source
# common.sh first so configuration and safety policy are shared.

realm_postgres_enabled() {
    [ "$REALM_MANAGE_POSTGRES" = 1 ]
}

realm_postgres_require_tools() {
    for realm_tool in initdb pg_ctl pg_isready psql createdb pg_dump; do
        realm_require_command "$realm_tool"
    done
}

realm_postgres_up() {
    realm_postgres_enabled || return 0
    [ -n "${DARK_MAGIC_REALM_POSTGRES_URL:-}" ] || realm_fail "managed PostgreSQL requires DARK_MAGIC_REALM_POSTGRES_URL"
    [ -n "$REALM_POSTGRES_PASSWORD" ] || realm_fail "managed PostgreSQL requires DARK_MAGIC_REALM_POSTGRES_PASSWORD"
    realm_postgres_require_tools
    mkdir -p "$REALM_POSTGRES_DIR" "$REALM_POSTGRES_SOCKET" "$REALM_LOG_DIR"
    chmod 700 "$REALM_POSTGRES_DIR" "$REALM_POSTGRES_SOCKET" "$REALM_LOG_DIR"
    if [ ! -f "$REALM_POSTGRES_DATA/PG_VERSION" ]; then
        realm_say "initializing local PostgreSQL data"
        realm_password_file="$REALM_POSTGRES_DIR/.initial-password.$$"
        printf '%s\n' "$REALM_POSTGRES_PASSWORD" > "$realm_password_file"
        chmod 600 "$realm_password_file"
        if ! initdb -D "$REALM_POSTGRES_DATA" --username="$REALM_POSTGRES_USER" --pwfile="$realm_password_file" \
            --auth-local=trust --auth-host=scram-sha-256 --encoding=UTF8 --no-locale >/dev/null; then
            rm -f "$realm_password_file"
            realm_fail "initializing local PostgreSQL"
        fi
        rm -f "$realm_password_file"
    fi
    if pg_ctl -D "$REALM_POSTGRES_DATA" status >/dev/null 2>&1; then
        return 0
    fi
    if pg_isready -h "$REALM_POSTGRES_HOST" -p "$REALM_POSTGRES_PORT" >/dev/null 2>&1; then
        realm_fail "PostgreSQL endpoint $REALM_POSTGRES_HOST:$REALM_POSTGRES_PORT is already owned by another cluster"
    fi
    realm_say "starting local PostgreSQL on $REALM_POSTGRES_HOST:$REALM_POSTGRES_PORT"
    pg_ctl -D "$REALM_POSTGRES_DATA" -l "$REALM_POSTGRES_LOG" \
        -o "-h $REALM_POSTGRES_HOST -p $REALM_POSTGRES_PORT -k $REALM_POSTGRES_SOCKET" \
        -w -t 30 start >/dev/null
    realm_attempt=0
    while ! pg_isready -h "$REALM_POSTGRES_HOST" -p "$REALM_POSTGRES_PORT" -U "$REALM_POSTGRES_USER" >/dev/null 2>&1; do
        [ "$realm_attempt" -lt 30 ] || realm_fail "local PostgreSQL did not become ready"
        sleep 1
        realm_attempt=$((realm_attempt + 1))
    done
    if ! PGPASSWORD="$REALM_POSTGRES_PASSWORD" psql -h "$REALM_POSTGRES_HOST" -p "$REALM_POSTGRES_PORT" \
        -U "$REALM_POSTGRES_USER" -d postgres -X -A -t -v ON_ERROR_STOP=1 \
        -c "SELECT 1 FROM pg_database WHERE datname = '$REALM_POSTGRES_DATABASE'" | grep -q 1; then
        PGPASSWORD="$REALM_POSTGRES_PASSWORD" createdb -h "$REALM_POSTGRES_HOST" -p "$REALM_POSTGRES_PORT" \
            -U "$REALM_POSTGRES_USER" "$REALM_POSTGRES_DATABASE"
        realm_say "created local PostgreSQL database $REALM_POSTGRES_DATABASE"
    fi
}

realm_postgres_down() {
    realm_postgres_enabled || return 0
    [ -f "$REALM_POSTGRES_DATA/PG_VERSION" ] || return 0
    realm_require_command pg_ctl
    if pg_ctl -D "$REALM_POSTGRES_DATA" status >/dev/null 2>&1; then
        realm_say "stopping local PostgreSQL"
        pg_ctl -D "$REALM_POSTGRES_DATA" -m fast -w -t 30 stop >/dev/null
    fi
}

realm_postgres_dump() {
    realm_destination=$1
    realm_require_command pg_dump
    PGPASSWORD="$REALM_POSTGRES_PASSWORD" pg_dump "$DARK_MAGIC_REALM_POSTGRES_URL" \
        --format=custom --file="$realm_destination"
    chmod 600 "$realm_destination"
}
