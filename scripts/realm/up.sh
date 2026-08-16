#!/bin/sh

set -eu
. "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)/common.sh"
. "$REALM_SCRIPT_DIR/postgres.sh"

realm_usage() {
    cat <<'EOF'
Usage: scripts/realm/up.sh [--no-build] [--foreground] [--print-config]

Build and start a local Realm plus its supervised game-worker executable.
Configuration is read from the Realm environment file, then from exported
DARK_MAGIC_REALM_* and MPQ_DIRECTORY overrides. Set DARK_MAGIC_REALM_ENV_FILE
to select a different file.

  --no-build      reuse binaries already under the runtime directory
  --foreground    run Realm in this terminal instead of creating a PID/log
  --print-config  print non-secret resolved configuration and exit
EOF
}

realm_no_build=0
realm_foreground=0
realm_print_config=0
while [ "$#" -gt 0 ]; do
    case "$1" in
        --no-build) realm_no_build=1 ;;
        --foreground) realm_foreground=1 ;;
        --print-config) realm_print_config=1 ;;
        -h|--help) realm_usage; exit 0 ;;
        *) realm_usage >&2; realm_fail "unknown argument: $1" ;;
    esac
    shift
done

realm_mode=$(realm_mail_mode)
[ -n "${DARK_MAGIC_REALM_POSTGRES_URL:-}" ] || realm_fail "Realm requires DARK_MAGIC_REALM_POSTGRES_URL"
if [ "$realm_mode" = smtp ] && [ -z "${DARK_MAGIC_REALM_SMTP_ADDRESS:-}" ]; then
    realm_fail "SMTP mail mode requires DARK_MAGIC_REALM_SMTP_ADDRESS"
fi

if [ "$realm_print_config" -eq 1 ]; then
    printf '%s\n' \
        "repository=$REALM_REPOSITORY" \
        "runtime_dir=$REALM_RUNTIME_DIR" \
        "data_dir=$REALM_DATA_DIR" \
        "listen=$REALM_LISTEN" \
        "url=$REALM_URL" \
        "storage=$(realm_storage_name)" \
        "postgres_management=$([ "$REALM_MANAGE_POSTGRES" = 1 ] && printf managed || printf external)" \
        "mail_mode=$realm_mode" \
        "workers=$(realm_workers_name)" \
		"operator_api=$REALM_OPERATOR_LISTEN" \
        "log_file=$REALM_LOG_FILE"
    exit 0
fi

realm_require_command go
realm_require_command curl
realm_ensure_layout
realm_postgres_up
realm_remove_stale_pid
if realm_pid=$(realm_running_pid); then
    realm_say "already running (pid $realm_pid, $REALM_URL)"
    exit 0
fi
if curl --silent --show-error --insecure --fail --max-time 1 "$REALM_URL/v1/status" >/dev/null 2>&1; then
    realm_fail "configured endpoint is already serving a Realm not owned by this runtime: $REALM_URL"
fi

realm_build() {
    realm_package=$1
    realm_output=$2
    realm_temporary="$realm_output.build.$$"
    trap 'rm -f "$realm_temporary"' EXIT HUP INT TERM
    (cd "$REALM_REPOSITORY" && go build -o "$realm_temporary" "$realm_package")
    chmod 700 "$realm_temporary"
    mv "$realm_temporary" "$realm_output"
    trap - EXIT HUP INT TERM
}

if [ "$realm_no_build" -eq 0 ]; then
    realm_say "building Realm"
    realm_build ./cmd/realm "$REALM_BIN"
    if [ "${DARK_MAGIC_REALM_DISABLE_WORKERS:-0}" != 1 ] && [ -z "${DARK_MAGIC_REALM_WORKER:-}" ]; then
        realm_say "building game worker"
        realm_build ./cmd/server "$REALM_WORKER_BIN"
    fi
fi
[ -x "$REALM_BIN" ] || realm_fail "Realm binary is unavailable: $REALM_BIN"

set -- "$REALM_BIN" \
    --env-file "$REALM_ENV_FILE" \
    --admin-shell=false \
    --listen "$REALM_LISTEN" \
    --data-dir "$REALM_DATA_DIR" \
    --log-level "$REALM_LOG_LEVEL" \
	--operator-token-file "$REALM_OPERATOR_TOKEN_FILE" \
	--operator-listen "$REALM_OPERATOR_LISTEN" \
    --account-base-url "${DARK_MAGIC_REALM_ACCOUNT_URL:-$REALM_URL}" \
    --account-mail-mode "$realm_mode" \
    --worker-control-listen "${DARK_MAGIC_REALM_WORKER_CONTROL_LISTEN:-127.0.0.1:0}" \
    --worker-game-listen "${DARK_MAGIC_REALM_WORKER_GAME_LISTEN:-0.0.0.0:0}"

if [ "${DARK_MAGIC_REALM_DISABLE_WORKERS:-0}" != 1 ]; then
    [ -x "$REALM_WORKER_BIN" ] || realm_fail "game-worker binary is unavailable: $REALM_WORKER_BIN"
    set -- "$@" --worker-executable "$REALM_WORKER_BIN"
fi
if [ -n "${DARK_MAGIC_REALM_GAME_HOST:-}" ]; then
    set -- "$@" --worker-game-advertise-host "$DARK_MAGIC_REALM_GAME_HOST"
fi
if [ "${DARK_MAGIC_REALM_SMTP_REQUIRE_TLS:-0}" = 1 ]; then
    set -- "$@" --smtp-require-tls
fi

if [ -z "${MPQ_DIRECTORY:-}" ] && [ "${DARK_MAGIC_REALM_DISABLE_WORKERS:-0}" != 1 ]; then
    realm_say "warning: MPQ_DIRECTORY is unset; workers cannot load a production game world"
fi

if [ "$realm_foreground" -eq 1 ]; then
    realm_say "starting in foreground at $REALM_URL"
    exec "$@"
fi

printf '\n[%s] starting Realm\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" >> "$REALM_LOG_FILE"
nohup "$@" </dev/null >> "$REALM_LOG_FILE" 2>&1 &
realm_pid=$!
realm_pid_temporary="$REALM_PID_FILE.tmp.$$"
printf '%s\n' "$realm_pid" > "$realm_pid_temporary"
mv "$realm_pid_temporary" "$REALM_PID_FILE"

realm_elapsed=0
while [ "$realm_elapsed" -lt "$REALM_START_TIMEOUT" ]; do
    if ! kill -0 "$realm_pid" 2>/dev/null; then
        rm -f "$REALM_PID_FILE"
        tail -n 40 "$REALM_LOG_FILE" >&2 || true
        realm_fail "Realm exited before becoming ready"
    fi
    if curl --silent --show-error --insecure --fail --max-time 1 "$REALM_URL/v1/status" >/dev/null 2>&1; then
        # Recheck after the request. A conflicting Realm may answer while the
        # process we just spawned is exiting because it could not bind.
        if ! kill -0 "$realm_pid" 2>/dev/null; then
            rm -f "$REALM_PID_FILE"
            tail -n 40 "$REALM_LOG_FILE" >&2 || true
            realm_fail "Realm exited while another process answered its readiness probe"
        fi
        realm_say "ready (pid $realm_pid, $REALM_URL)"
        realm_say "log: $REALM_LOG_FILE"
        exit 0
    fi
    sleep 1
    realm_elapsed=$((realm_elapsed + 1))
done

"$REALM_SCRIPT_DIR/down.sh" >/dev/null 2>&1 || true
tail -n 40 "$REALM_LOG_FILE" >&2 || true
realm_fail "Realm did not become ready within ${REALM_START_TIMEOUT}s"
