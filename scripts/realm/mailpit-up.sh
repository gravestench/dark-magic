#!/bin/sh

set -eu

# mailpit_usage documents the loopback-only development sink and every supported collision override.
mailpit_usage() {
    cat <<'EOF'
Usage: scripts/realm/mailpit-up.sh

Start the pinned, loopback-only Mailpit SMTP sink used for local Realm account
verification and password-reset testing.

Optional overrides:
  DARK_MAGIC_REALM_MAILPIT_IMAGE       default axllent/mailpit:v1.30.0
  DARK_MAGIC_REALM_MAILPIT_NAME        default dark-magic-realm-mailpit
  DARK_MAGIC_REALM_MAILPIT_SMTP_PORT   default 1025
  DARK_MAGIC_REALM_MAILPIT_HTTP_PORT   default 8025
  DARK_MAGIC_REALM_CONTAINER_CLI       default docker (podman is supported)
EOF
}

case "${1:-}" in
    -h|--help) mailpit_usage; exit 0 ;;
    '') ;;
    *) mailpit_usage >&2; exit 2 ;;
esac

mailpit_image=${DARK_MAGIC_REALM_MAILPIT_IMAGE:-axllent/mailpit:v1.30.0}
mailpit_name=${DARK_MAGIC_REALM_MAILPIT_NAME:-dark-magic-realm-mailpit}
mailpit_smtp_port=${DARK_MAGIC_REALM_MAILPIT_SMTP_PORT:-1025}
mailpit_http_port=${DARK_MAGIC_REALM_MAILPIT_HTTP_PORT:-8025}
mailpit_container_cli=${DARK_MAGIC_REALM_CONTAINER_CLI:-docker}

case "$mailpit_container_cli" in
    docker|podman) ;;
    *)
        printf '%s\n' 'realm mailpit: container CLI must be docker or podman' >&2
        exit 1
        ;;
esac
command -v "$mailpit_container_cli" >/dev/null 2>&1 || {
    printf '%s\n' "realm mailpit: $mailpit_container_cli is required" >&2
    exit 1
}
case "$mailpit_name" in
    ''|*[!A-Za-z0-9_.-]*)
        printf '%s\n' 'realm mailpit: invalid container name' >&2
        exit 1
        ;;
esac
for mailpit_port in "$mailpit_smtp_port" "$mailpit_http_port"; do
    case "$mailpit_port" in
        ''|*[!0-9]*)
            printf '%s\n' 'realm mailpit: ports must be numeric' >&2
            exit 1
            ;;
    esac
    [ "$mailpit_port" -ge 1 ] && [ "$mailpit_port" -le 65535 ] || {
        printf '%s\n' 'realm mailpit: port is outside 1-65535' >&2
        exit 1
    }
done

if "$mailpit_container_cli" container inspect "$mailpit_name" >/dev/null 2>&1; then
    # Ownership labels prevent a name collision from making this script start an operator-managed container.
    mailpit_owned=$("$mailpit_container_cli" inspect \
        --format '{{ index .Config.Labels "com.dark-magic.realm-mailpit" }}' \
        "$mailpit_name" 2>/dev/null || true)
    [ "$mailpit_owned" = 1 ] || {
        printf '%s\n' "realm mailpit: refusing container not owned by Dark Magic: $mailpit_name" >&2
        exit 1
    }
    "$mailpit_container_cli" start "$mailpit_name" >/dev/null
else
    # Bind both published ports explicitly to loopback; the development mailbox may contain verification and password
    # recovery links that must not be exposed on the host's external interfaces.
    "$mailpit_container_cli" run --detach --name "$mailpit_name" \
        --label com.dark-magic.realm-mailpit=1 \
        --publish "127.0.0.1:$mailpit_smtp_port:1025" \
        --publish "127.0.0.1:$mailpit_http_port:8025" \
        "$mailpit_image" >/dev/null
fi

mailpit_attempt=0
while [ "$mailpit_attempt" -lt 60 ]; do
    if curl --silent --fail "http://127.0.0.1:$mailpit_http_port/readyz" >/dev/null 2>&1; then
        printf '%s\n' \
            "realm mailpit: ready (SMTP 127.0.0.1:$mailpit_smtp_port, \
inbox http://127.0.0.1:$mailpit_http_port)"
        exit 0
    fi
    mailpit_attempt=$((mailpit_attempt + 1))
    sleep 0.25
done

printf '%s\n' 'realm mailpit: readiness timed out' >&2
exit 1
