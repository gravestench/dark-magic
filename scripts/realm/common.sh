#!/bin/sh

# Shared local Realm lifecycle configuration. Public scripts source this file;
# it is not intended to be run directly.

set -eu
umask 077

REALM_SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
REALM_REPOSITORY=$(CDPATH= cd -- "$REALM_SCRIPT_DIR/../.." && pwd -P)

# realm_say gives lifecycle messages a stable prefix so operators can distinguish them from child-process output.
realm_say() {
    printf '%s\n' "realm: $*"
}

# realm_fail preserves the same prefix on stderr and terminates immediately, preventing callers from continuing after
# a failed safety check.
realm_fail() {
    printf '%s\n' "realm: $*" >&2
    exit 1
}

# realm_require_command gives callers one consistent dependency error; lifecycle
# entry points invoke it before the operation that needs the executable.
realm_require_command() {
    command -v "$1" >/dev/null 2>&1 || realm_fail "required command is unavailable: $1"
}

# Resolve the platform's ordinary configuration directory unless the caller supplies an isolated location. Tests and
# disposable acceptance runs rely on the override to avoid reading or changing an operator's real configuration.
if [ -n "${DARK_MAGIC_CONFIG_DIR:-}" ]; then
    REALM_CONFIG_DIR=$DARK_MAGIC_CONFIG_DIR
elif [ "$(uname -s)" = Darwin ]; then
    REALM_CONFIG_DIR="${HOME:?}/Library/Application Support/dark-magic"
elif [ -n "${APPDATA:-}" ]; then
    REALM_CONFIG_DIR="$APPDATA/dark-magic"
else
    REALM_CONFIG_DIR="${XDG_CONFIG_HOME:-${HOME:?}/.config}/dark-magic"
fi
REALM_ENV_FILE=${DARK_MAGIC_REALM_ENV_FILE:-"$REALM_CONFIG_DIR/realm.env"}

# The checked-in template makes first use approachable, but a custom environment path is always treated as
# operator-owned and therefore is never created or overwritten here.
mkdir -p "$REALM_CONFIG_DIR"
chmod 700 "$REALM_CONFIG_DIR"
if [ ! -e "$REALM_ENV_FILE" ] && [ "$REALM_ENV_FILE" = "$REALM_CONFIG_DIR/realm.env" ]; then
    cp "$REALM_REPOSITORY/internal/app/envconfig/templates/realm.env" "$REALM_ENV_FILE"
    chmod 600 "$REALM_ENV_FILE"
fi

# realm_load_environment implements the narrow dotenv syntax accepted by the Go
# configuration loader. Already exported values take precedence, including
# caller overrides and an earlier duplicate entry in the same file.
realm_load_environment() {
    realm_environment_path=$1
    [ -f "$realm_environment_path" ] || realm_fail "environment file is unavailable: $realm_environment_path"
    while IFS= read -r realm_line || [ -n "$realm_line" ]; do
        realm_line=$(printf '%s' "$realm_line" | tr -d '\r')
        case "$realm_line" in ''|'#'*) continue ;; esac
        case "$realm_line" in export\ *) realm_line=${realm_line#export } ;; esac
        realm_key=${realm_line%%=*}
        realm_value=${realm_line#*=}
        case "$realm_key" in
            ''|[0-9]*|*[!A-Za-z0-9_]*)
                realm_fail "invalid variable name in $realm_environment_path: $realm_key"
                ;;
        esac

        # eval is limited to a validated identifier and checks only whether the
        # name is already set; the file value is never evaluated as shell source.
        eval "realm_is_exported=\${$realm_key+x}"
        [ "$realm_is_exported" = x ] && continue

        case "$realm_value" in
            \"*\")
                realm_value=${realm_value#\"}
                realm_value=${realm_value%\"}
                realm_value=$(printf '%b' "$realm_value")
                ;;
            \'*\')
                realm_value=${realm_value#\'}
                realm_value=${realm_value%\'}
                ;;
        esac
        export "$realm_key=$realm_value"
    done < "$realm_environment_path"
}

realm_load_environment "$REALM_ENV_FILE"

REALM_RUNTIME_DIR=${DARK_MAGIC_REALM_RUNTIME_DIR:-"$REALM_REPOSITORY/.local/realm"}
REALM_DATA_DIR=${DARK_MAGIC_REALM_DATA:-"$REALM_RUNTIME_DIR/data"}
REALM_BIN_DIR="$REALM_RUNTIME_DIR/bin"
REALM_LOG_DIR="$REALM_RUNTIME_DIR/logs"
REALM_BACKUP_DIR="$REALM_RUNTIME_DIR/backups"
REALM_PID_FILE="$REALM_RUNTIME_DIR/realm.pid"
REALM_LOG_FILE=${DARK_MAGIC_REALM_LOG_FILE:-"$REALM_LOG_DIR/realm.log"}
REALM_OPERATOR_TOKEN_FILE=${DARK_MAGIC_REALM_OPERATOR_TOKEN_FILE:-"$REALM_RUNTIME_DIR/operator-token"}
REALM_OPERATOR_LISTEN=${DARK_MAGIC_REALM_OPERATOR_LISTEN:-"127.0.0.1:6113"}
REALM_BIN="$REALM_BIN_DIR/realm"
REALM_WORKER_BIN=${DARK_MAGIC_REALM_WORKER:-"$REALM_BIN_DIR/server"}
REALM_LISTEN=${DARK_MAGIC_REALM_LISTEN:-"127.0.0.1:6112"}
REALM_URL=${DARK_MAGIC_REALM_URL:-"https://$REALM_LISTEN"}
REALM_LOG_LEVEL=${DARK_MAGIC_REALM_LOG_LEVEL:-debug}
REALM_START_TIMEOUT=${DARK_MAGIC_REALM_START_TIMEOUT_SECONDS:-30}
REALM_MANAGE_POSTGRES=${DARK_MAGIC_REALM_MANAGE_POSTGRES:-0}
REALM_POSTGRES_HOST=${DARK_MAGIC_REALM_POSTGRES_HOST:-127.0.0.1}
REALM_POSTGRES_PORT=${DARK_MAGIC_REALM_POSTGRES_PORT:-55432}
REALM_POSTGRES_DATABASE=${DARK_MAGIC_REALM_POSTGRES_DATABASE:-realm}
REALM_POSTGRES_USER=${DARK_MAGIC_REALM_POSTGRES_USER:-realm}
REALM_POSTGRES_PASSWORD=${DARK_MAGIC_REALM_POSTGRES_PASSWORD:-}
REALM_POSTGRES_DIR="$REALM_RUNTIME_DIR/postgres"
REALM_POSTGRES_DATA="$REALM_POSTGRES_DIR/data"
REALM_POSTGRES_SOCKET=${DARK_MAGIC_REALM_POSTGRES_SOCKET:-"/tmp/dark-magic-realm-postgres-$REALM_POSTGRES_PORT"}
REALM_POSTGRES_LOG="$REALM_LOG_DIR/postgres.log"

# realm_validate_managed_postgres confines the optional managed cluster to predictable local inputs. External
# PostgreSQL URLs remain opaque because their host and credentials belong to the external database operator.
realm_validate_managed_postgres() {
    [ "$REALM_MANAGE_POSTGRES" = 1 ] || return 0
    case "$REALM_POSTGRES_HOST" in
        127.0.0.1|localhost|::1) ;;
        *) realm_fail "managed PostgreSQL must bind to a loopback host" ;;
    esac
    case "$REALM_POSTGRES_PORT" in
        ''|*[!0-9]*) realm_fail "managed PostgreSQL port must be numeric" ;;
    esac
    [ "$REALM_POSTGRES_PORT" -ge 1 ] && [ "$REALM_POSTGRES_PORT" -le 65535 ] || \
        realm_fail "managed PostgreSQL port is outside 1-65535"
    for realm_identifier in "$REALM_POSTGRES_DATABASE" "$REALM_POSTGRES_USER"; do
        case "$realm_identifier" in
            ''|[0-9]*|*[!A-Za-z0-9_]*)
                realm_fail "managed PostgreSQL names must be simple SQL identifiers"
                ;;
        esac
    done
    case "$REALM_POSTGRES_PASSWORD" in
        *'
'*) realm_fail "managed PostgreSQL password cannot contain a newline" ;;
    esac
    case "$REALM_POSTGRES_SOCKET" in
        ''|*[!A-Za-z0-9_./-]*)
            realm_fail "managed PostgreSQL socket path contains unsupported characters"
            ;;
    esac
}

realm_validate_managed_postgres

# realm_ensure_layout creates every private lifecycle directory before binaries, logs, or credentials are written.
# Reapplying restrictive permissions also repairs directories created under an overly permissive umask.
realm_ensure_layout() {
    mkdir -p "$REALM_RUNTIME_DIR" "$REALM_DATA_DIR" "$REALM_BIN_DIR" "$REALM_LOG_DIR" "$REALM_BACKUP_DIR"
    chmod 700 "$REALM_RUNTIME_DIR" "$REALM_DATA_DIR" "$REALM_BIN_DIR" "$REALM_LOG_DIR" "$REALM_BACKUP_DIR"
}

# realm_read_pid accepts only a numeric first line, making empty, truncated, or
# nonnumeric ownership records stale rather than signal targets.
realm_read_pid() {
    [ -f "$REALM_PID_FILE" ] || return 1
    IFS= read -r realm_pid < "$REALM_PID_FILE" || return 1
    case "$realm_pid" in
        ''|*[!0-9]*) return 1 ;;
    esac
    printf '%s\n' "$realm_pid"
}

# realm_process_matches requires both liveness and the configured binary path in
# the process command, reducing the risk that PID reuse targets unrelated work.
realm_process_matches() {
    realm_pid=$1
    kill -0 "$realm_pid" 2>/dev/null || return 1
    realm_command=$(ps -p "$realm_pid" -o command= 2>/dev/null || true)
    case "$realm_command" in
        *"$REALM_BIN"*) return 0 ;;
        *) return 1 ;;
    esac
}

# realm_running_pid composes PID parsing and ownership validation for callers that need a trustworthy live process.
realm_running_pid() {
    realm_pid=$(realm_read_pid) || return 1
    realm_process_matches "$realm_pid" || return 1
    printf '%s\n' "$realm_pid"
}

# realm_remove_stale_pid removes only an invalid ownership record; it never signals the process named by that record.
realm_remove_stale_pid() {
    if [ -f "$REALM_PID_FILE" ] && ! realm_running_pid >/dev/null 2>&1; then
        rm -f "$REALM_PID_FILE"
    fi
}

# realm_mail_mode resolves the implicit compatibility defaults in priority order so SMTP configuration is selected
# automatically without overriding an explicit account-mail policy.
realm_mail_mode() {
    if [ -n "${DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE:-}" ]; then
        printf '%s\n' "$DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE"
    elif [ -n "${DARK_MAGIC_REALM_SMTP_ADDRESS:-}" ]; then
        printf '%s\n' smtp
    elif [ -n "${DARK_MAGIC_REALM_POSTGRES_URL:-}" ]; then
        printf '%s\n' auto-verify
    else
        printf '%s\n' disabled
    fi
}

# realm_storage_name centralizes the operator-facing storage label without exposing the credential-bearing URL.
realm_storage_name() {
    printf '%s\n' postgresql
}

# realm_workers_name translates the disable flag into a concise, non-secret configuration summary.
realm_workers_name() {
    if [ "${DARK_MAGIC_REALM_DISABLE_WORKERS:-0}" = 1 ]; then
        printf '%s\n' disabled
    else
        printf '%s\n' enabled
    fi
}

# realm_assert_safe_data_directory rejects broad and self-referential paths before fresh-install moves any state.
realm_assert_safe_data_directory() {
    [ -n "$REALM_DATA_DIR" ] || realm_fail "data directory is empty"
    case "$REALM_DATA_DIR" in
        /|"$REALM_REPOSITORY"|"$REALM_RUNTIME_DIR"|"${HOME:-/nonexistent}")
            realm_fail "refusing unsafe Realm data directory: $REALM_DATA_DIR"
            ;;
        "$REALM_BACKUP_DIR"|"$REALM_BACKUP_DIR"/*)
            realm_fail "Realm data directory cannot be inside its backup directory"
            ;;
    esac
}
