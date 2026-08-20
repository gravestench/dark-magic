#!/bin/sh

set -eu
. "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)/common.sh"
. "$REALM_SCRIPT_DIR/postgres.sh"

# realm_usage states the ownership source so operators understand that shutdown never discovers processes by port.
realm_usage() {
    cat <<'EOF'
Usage: scripts/realm/down.sh

Gracefully stop the Realm recorded in the local lifecycle PID file. Realm then
drains and terminates all game workers it supervises.
EOF
}

case "${1:-}" in
    -h|--help) realm_usage; exit 0 ;;
    '') ;;
    *) realm_usage >&2; realm_fail "unknown argument: $1" ;;
esac

if ! realm_pid=$(realm_read_pid); then
    rm -f "$REALM_PID_FILE"
    realm_say "not running"
    realm_postgres_down
    exit 0
fi

# A numeric PID is not sufficient ownership proof because operating systems eventually reuse process IDs. Match the
# recorded process command before sending either the graceful or forced signal.
if ! realm_process_matches "$realm_pid"; then
    if kill -0 "$realm_pid" 2>/dev/null; then
        realm_fail "PID $realm_pid does not belong to $REALM_BIN; refusing to signal it"
    fi
    rm -f "$REALM_PID_FILE"
    realm_say "removed stale PID file"
    realm_postgres_down
    exit 0
fi

realm_say "stopping pid $realm_pid"
kill -TERM "$realm_pid"
realm_attempt=0

# Give Realm time to drain workers and commit their state. SIGKILL is reserved for a process that outlives the full
# grace period, since it bypasses those shutdown guarantees.
while kill -0 "$realm_pid" 2>/dev/null && [ "$realm_attempt" -lt 30 ]; do
    sleep 1
    realm_attempt=$((realm_attempt + 1))
done
if kill -0 "$realm_pid" 2>/dev/null; then
    realm_say "grace period expired; forcing pid $realm_pid to exit"
    kill -KILL "$realm_pid"
fi
rm -f "$REALM_PID_FILE"
realm_say "stopped"
realm_postgres_down
