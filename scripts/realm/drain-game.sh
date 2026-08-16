#!/bin/sh

set -eu
. "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)/common.sh"

realm_usage() {
    cat <<'EOF'
Usage: scripts/realm/drain-game.sh GAME_ID

Hide a Realm game from discovery/admission, canonically commit every active
character, and retire its worker. The owner-only operator token is transmitted
only to the loopback operator API and is never placed in the curl command line.
EOF
}

case "${1:-}" in
    -h|--help) realm_usage; exit 0 ;;
    '') realm_usage >&2; realm_fail "game ID is required" ;;
esac
[ "$#" -eq 1 ] || { realm_usage >&2; realm_fail "exactly one game ID is required"; }
realm_game_id=$1
case "$realm_game_id" in *[!A-Za-z0-9_-]*) realm_fail "game ID contains unsupported characters" ;; esac

realm_require_command curl
[ -f "$REALM_OPERATOR_TOKEN_FILE" ] || realm_fail "operator token is unavailable: $REALM_OPERATOR_TOKEN_FILE"
realm_operator_token=$(tr -d '\r\n' < "$REALM_OPERATOR_TOKEN_FILE")
[ "${#realm_operator_token}" -ge 32 ] || realm_fail "operator token is invalid"
realm_header=$(mktemp "$REALM_RUNTIME_DIR/.operator-header.XXXXXX")
trap 'rm -f "$realm_header"' EXIT HUP INT TERM
chmod 600 "$realm_header"
printf 'Authorization: Bearer %s\n' "$realm_operator_token" > "$realm_header"
unset realm_operator_token

curl --silent --show-error --fail --insecure \
    --header "@$realm_header" \
    --header 'Content-Type: application/json' \
    --data "{\"game_id\":\"$realm_game_id\"}" \
    "https://$REALM_OPERATOR_LISTEN/v1/operator/games/drain"
printf '\n'
