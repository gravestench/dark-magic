#!/bin/sh

set -eu

mailpit_usage() {
    cat <<'EOF'
Usage: scripts/realm/mailpit-down.sh

Remove the local Mailpit container created by mailpit-up.sh.

Optional override:
  DARK_MAGIC_REALM_MAILPIT_NAME        default dark-magic-realm-mailpit
  DARK_MAGIC_REALM_CONTAINER_CLI       default docker (podman is supported)
EOF
}

case "${1:-}" in
    -h|--help) mailpit_usage; exit 0 ;;
    '') ;;
    *) mailpit_usage >&2; exit 2 ;;
esac

mailpit_name=${DARK_MAGIC_REALM_MAILPIT_NAME:-dark-magic-realm-mailpit}
mailpit_container_cli=${DARK_MAGIC_REALM_CONTAINER_CLI:-docker}
case "$mailpit_container_cli" in docker|podman) ;; *) printf '%s\n' 'realm mailpit: container CLI must be docker or podman' >&2; exit 1 ;; esac
command -v "$mailpit_container_cli" >/dev/null 2>&1 || {
    printf '%s\n' "realm mailpit: $mailpit_container_cli is required" >&2
    exit 1
}
if ! "$mailpit_container_cli" container inspect "$mailpit_name" >/dev/null 2>&1; then
    printf '%s\n' 'realm mailpit: already stopped'
    exit 0
fi
mailpit_owned=$("$mailpit_container_cli" inspect --format '{{ index .Config.Labels "com.dark-magic.realm-mailpit" }}' "$mailpit_name" 2>/dev/null || true)
[ "$mailpit_owned" = 1 ] || {
    printf '%s\n' "realm mailpit: refusing container not owned by Dark Magic: $mailpit_name" >&2
    exit 1
}
"$mailpit_container_cli" rm --force "$mailpit_name" >/dev/null
printf '%s\n' 'realm mailpit: stopped'
