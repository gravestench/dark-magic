#!/bin/sh

# Disposable PostgreSQL-backed account acceptance. This deliberately crosses
# the ordinary cmd/realm HTTPS composition root and browser recovery form.

set -eu
umask 077

realm_test_usage() {
    cat <<'EOF'
Usage: scripts/realm/test-production.sh

Start a disposable PostgreSQL-backed Realm, exercise schema initialization, signup,
explicit login, browser password recovery, restart persistence, and teardown.

Optional collision overrides:
  DARK_MAGIC_REALM_TEST_LISTEN         default 127.0.0.1:16114
  DARK_MAGIC_REALM_TEST_OPERATOR_LISTEN default 127.0.0.1:16115
  DARK_MAGIC_REALM_TEST_POSTGRES_PORT  default 55439
  DARK_MAGIC_REALM_TEST_MAILPIT_SMTP_PORT default 16116
  DARK_MAGIC_REALM_TEST_MAILPIT_HTTP_PORT default 16117

Set DARK_MAGIC_REALM_TEST_MPQ_DIRECTORY to an operator-owned MPQ directory to
also run the complete character/channel/worker/QUIC/checkpoint/reconnect/commit
acceptance. Without it, the redistributable account and PostgreSQL checks run.

Set DARK_MAGIC_REALM_TEST_MAILPIT=1 to deliver verification and recovery mail
through the pinned local Mailpit container and consume both browser forms.
EOF
}

case "${1:-}" in
    -h|--help) realm_test_usage; exit 0 ;;
    '') ;;
    *) realm_test_usage >&2; exit 2 ;;
esac

realm_test_script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
realm_test_repository=$(CDPATH= cd -- "$realm_test_script_dir/../.." && pwd -P)
realm_test_root=$(mktemp -d "${TMPDIR:-/tmp}/dark-magic-realm-production.XXXXXX")
realm_test_config="$realm_test_root/config"
realm_test_runtime="$realm_test_root/runtime"
realm_test_listen=${DARK_MAGIC_REALM_TEST_LISTEN:-127.0.0.1:16114}
realm_test_operator_listen=${DARK_MAGIC_REALM_TEST_OPERATOR_LISTEN:-127.0.0.1:16115}
realm_test_port=${DARK_MAGIC_REALM_TEST_POSTGRES_PORT:-55439}
realm_test_mpq=${DARK_MAGIC_REALM_TEST_MPQ_DIRECTORY:-}
realm_test_mailpit=${DARK_MAGIC_REALM_TEST_MAILPIT:-0}
realm_test_mailpit_smtp_port=${DARK_MAGIC_REALM_TEST_MAILPIT_SMTP_PORT:-16116}
realm_test_mailpit_http_port=${DARK_MAGIC_REALM_TEST_MAILPIT_HTTP_PORT:-16117}
realm_test_mailpit_name="dark-magic-realm-mailpit-$$"
realm_test_container_cli=${DARK_MAGIC_REALM_CONTAINER_CLI:-docker}
realm_test_mailpit_started=0
realm_test_disable_workers=1
[ -z "$realm_test_mpq" ] || realm_test_disable_workers=0
realm_test_origin="https://$realm_test_listen"
realm_test_operator_origin="https://$realm_test_operator_listen"
realm_test_database_url="postgres://realm:dark-magic-production-test@127.0.0.1:$realm_test_port/realm?sslmode=disable"
realm_test_passed=0

realm_test_cleanup() {
    DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/down.sh" >/dev/null 2>&1 || true
    if [ "$realm_test_mailpit_started" -eq 1 ]; then
        DARK_MAGIC_REALM_MAILPIT_NAME="$realm_test_mailpit_name" \
        DARK_MAGIC_REALM_CONTAINER_CLI="$realm_test_container_cli" \
            "$realm_test_script_dir/mailpit-down.sh" >/dev/null 2>&1 || true
    fi
    if [ "$realm_test_passed" -eq 1 ]; then
        case "$realm_test_root" in
            "${TMPDIR:-/tmp}"/dark-magic-realm-production.*) rm -rf "$realm_test_root" ;;
        esac
    else
        printf '%s\n' "realm production test: retained failure diagnostics at $realm_test_root" >&2
    fi
}
trap realm_test_cleanup EXIT HUP INT TERM

for realm_test_tool in go curl psql; do
    command -v "$realm_test_tool" >/dev/null 2>&1 || {
        printf '%s\n' "realm production test: required command is unavailable: $realm_test_tool" >&2
        exit 1
    }
done
if [ "$realm_test_mailpit" = 1 ]; then
    command -v "$realm_test_container_cli" >/dev/null 2>&1 || {
        printf '%s\n' "realm production test: Mailpit mode requires $realm_test_container_cli" >&2
        exit 1
    }
elif [ "$realm_test_mailpit" != 0 ]; then
    printf '%s\n' 'realm production test: DARK_MAGIC_REALM_TEST_MAILPIT must be 0 or 1' >&2
    exit 1
fi

cd "$realm_test_repository"
[ -z "$realm_test_mpq" ] || [ -d "$realm_test_mpq" ] || {
    printf '%s\n' "realm production test: MPQ directory is unavailable: $realm_test_mpq" >&2
    exit 1
}
DARK_MAGIC_CONFIG_DIR="$realm_test_config" go run ./internal/dev/tools/env_config --role realm \
    --set MPQ_DIRECTORY="$realm_test_mpq" \
    --set DARK_MAGIC_REALM_LISTEN="$realm_test_listen" \
	--set DARK_MAGIC_REALM_OPERATOR_LISTEN="$realm_test_operator_listen" \
    --set DARK_MAGIC_REALM_ACCOUNT_URL="$realm_test_origin" \
    --set DARK_MAGIC_REALM_RUNTIME_DIR="$realm_test_runtime" \
    --set DARK_MAGIC_REALM_DATA="$realm_test_root/data" \
    --set DARK_MAGIC_REALM_POSTGRES_PORT="$realm_test_port" \
    --set DARK_MAGIC_REALM_POSTGRES_PASSWORD=dark-magic-production-test \
    --set DARK_MAGIC_REALM_POSTGRES_URL="$realm_test_database_url" \
    --set DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE=disabled \
    --set DARK_MAGIC_REALM_CHECKPOINT_INTERVAL=250ms \
    --set DARK_MAGIC_REALM_DISABLE_WORKERS="$realm_test_disable_workers" >/dev/null

DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/up.sh"

realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' "$realm_test_origin/v1/status")
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: status endpoint returned $realm_test_status" >&2
    exit 1
}

# The privileged lifecycle API must not exist on the player-facing listener.
realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' --data '{"game_id":"missing"}' \
    "$realm_test_origin/v1/operator/games/drain")
[ "$realm_test_status" = 404 ] || {
    printf '%s\n' "realm production test: operator route leaked onto public listener (status $realm_test_status)" >&2
    exit 1
}
realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' -H 'Authorization: Bearer invalid' \
    --data '{"game_id":"missing"}' "$realm_test_operator_origin/v1/operator/games/drain")
[ "$realm_test_status" = 401 ] || {
    printf '%s\n' "realm production test: operator listener accepted an invalid credential (status $realm_test_status)" >&2
    exit 1
}
realm_test_operator_header="$realm_test_root/operator-header"
printf 'Authorization: Bearer %s\n' "$(tr -d '\r\n' < "$realm_test_runtime/operator-token")" > "$realm_test_operator_header"
realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' -H "@$realm_test_operator_header" \
    --data '{"game_id":"missing"}' "$realm_test_operator_origin/v1/operator/games/drain")
realm_test_operator_missing_status=404
[ "$realm_test_disable_workers" -eq 0 ] || realm_test_operator_missing_status=503
[ "$realm_test_status" = "$realm_test_operator_missing_status" ] || {
    printf '%s\n' "realm production test: authenticated operator request returned $realm_test_status" >&2
    exit 1
}

realm_test_schema_tables=$(psql "$realm_test_database_url" -X -A -t -v ON_ERROR_STOP=1 \
	-c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('realm_accounts', 'realm_characters', 'realm_games', 'realm_game_checkpoints', 'realm_mail_outbox')")
[ "$realm_test_schema_tables" = 5 ] || {
    printf '%s\n' "realm production test: PostgreSQL schema was not initialized" >&2
    exit 1
}

DARK_MAGIC_TEST_POSTGRES="$realm_test_database_url" GOCACHE="${GOCACHE:-/tmp/dark-magic-realm-go-build}" \
    go test ./internal/app/realm -run '^TestPostgres' -count=1

# Repository tests intentionally claim and complete outbox jobs themselves.
# They also use durable game fixtures. Reset this disposable pre-production
# schema before the end-to-end journey so fixture state cannot masquerade as a
# player session. Enable the mail worker only after the reset so it cannot race
# a repository test for the same outbox row.
psql "$realm_test_database_url" -X -v ON_ERROR_STOP=1 -c 'DROP SCHEMA public CASCADE' >/dev/null
psql "$realm_test_database_url" -X -v ON_ERROR_STOP=1 -c 'CREATE SCHEMA public' >/dev/null
DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/down.sh"
if [ "$realm_test_mailpit" -eq 1 ]; then
    DARK_MAGIC_REALM_MAILPIT_NAME="$realm_test_mailpit_name" \
    DARK_MAGIC_REALM_CONTAINER_CLI="$realm_test_container_cli" \
    DARK_MAGIC_REALM_MAILPIT_SMTP_PORT="$realm_test_mailpit_smtp_port" \
    DARK_MAGIC_REALM_MAILPIT_HTTP_PORT="$realm_test_mailpit_http_port" \
        "$realm_test_script_dir/mailpit-up.sh"
    realm_test_mailpit_started=1
    DARK_MAGIC_CONFIG_DIR="$realm_test_config" go run ./internal/dev/tools/env_config --role realm \
        --set DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE=smtp \
        --set DARK_MAGIC_REALM_SMTP_ADDRESS="127.0.0.1:$realm_test_mailpit_smtp_port" \
        --set DARK_MAGIC_REALM_SMTP_FROM=realm@example.test >/dev/null
else
    DARK_MAGIC_CONFIG_DIR="$realm_test_config" go run ./internal/dev/tools/env_config --role realm \
        --set DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE=auto-verify >/dev/null
fi
DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/up.sh" --no-build

realm_test_suffix="$(date -u '+%s')_$$"
realm_test_name="Smoke_$realm_test_suffix"
realm_test_email="smoke_$realm_test_suffix@example.test"
realm_test_old_password=old-production-password
realm_test_new_password=new-production-password

realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' \
    --data "{\"name\":\"$realm_test_name\",\"email\":\"$realm_test_email\",\"password\":\"$realm_test_old_password\"}" \
    "$realm_test_origin/v1/accounts")
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: signup returned $realm_test_status" >&2
    exit 1
}

if [ "$realm_test_mailpit" -eq 1 ]; then
    realm_test_verification_url=
    realm_test_attempt=0
    while [ "$realm_test_attempt" -lt 60 ] && [ -z "$realm_test_verification_url" ]; do
        realm_test_message=$(curl --silent --get \
            --data-urlencode "query=to:$realm_test_email" \
            "http://127.0.0.1:$realm_test_mailpit_http_port/view/latest.txt" || true)
        realm_test_verification_url=$(printf '%s\n' "$realm_test_message" | \
            sed -n 's#.*\(https://[^[:space:]]*/verify?token=[^[:space:]]*\).*#\1#p' | head -n 1)
        [ -n "$realm_test_verification_url" ] || sleep 0.25
        realm_test_attempt=$((realm_test_attempt + 1))
    done
    [ -n "$realm_test_verification_url" ] || {
        printf '%s\n' 'realm production test: Mailpit did not receive account verification mail' >&2
        exit 1
    }
    realm_test_verification_token=${realm_test_verification_url#*token=}
    realm_test_cookie_jar="$realm_test_root/browser-cookies"
    realm_test_verification_page="$realm_test_root/verification.html"
    realm_test_status=$(curl --silent --insecure --cookie-jar "$realm_test_cookie_jar" \
        --output "$realm_test_verification_page" --write-out '%{http_code}' "$realm_test_verification_url")
    [ "$realm_test_status" = 200 ] || {
        printf '%s\n' "realm production test: browser verification page returned $realm_test_status" >&2
        exit 1
    }
    realm_test_csrf=$(sed -n 's/.*name="csrf_token" value="\([a-f0-9][a-f0-9]*\)".*/\1/p' "$realm_test_verification_page" | head -n 1)
    [ "${#realm_test_csrf}" -eq 64 ] || {
        printf '%s\n' 'realm production test: browser verification page omitted its CSRF token' >&2
        exit 1
    }
    realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
        --cookie "$realm_test_cookie_jar" \
        --data-urlencode "token=$realm_test_verification_token" \
        --data-urlencode "csrf_token=$realm_test_csrf" \
        "$realm_test_origin/verify")
    [ "$realm_test_status" = 200 ] || {
        printf '%s\n' "realm production test: browser account verification returned $realm_test_status" >&2
        exit 1
    }
fi

realm_test_attempt=0
while [ "$realm_test_attempt" -lt 30 ]; do
    realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
        -H 'Content-Type: application/json' \
        --data "{\"name\":\"$realm_test_name\",\"password\":\"$realm_test_old_password\"}" \
        "$realm_test_origin/v1/sessions")
    [ "$realm_test_status" = 200 ] && break
    realm_test_attempt=$((realm_test_attempt + 1))
    sleep 1
done
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: verified account could not log in" >&2
    exit 1
}

realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' --data "{\"email\":\"$realm_test_email\"}" \
    "$realm_test_origin/v1/accounts/recovery")
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: recovery request returned $realm_test_status" >&2
    exit 1
}

realm_test_recovery_url=
realm_test_attempt=0
while [ "$realm_test_attempt" -lt 30 ] && [ -z "$realm_test_recovery_url" ]; do
    if [ "$realm_test_mailpit" -eq 1 ]; then
        realm_test_message=$(curl --silent --get \
            --data-urlencode "query=to:$realm_test_email subject:Reset" \
            "http://127.0.0.1:$realm_test_mailpit_http_port/view/latest.txt" || true)
        realm_test_recovery_url=$(printf '%s\n' "$realm_test_message" | \
            sed -n 's#.*\(https://[^[:space:]]*/recover?token=[^[:space:]]*\).*#\1#p' | head -n 1)
    else
        realm_test_recovery_url=$(psql "$realm_test_database_url" -X -A -t -v ON_ERROR_STOP=1 \
            -c "SELECT payload->>'recovery_url' FROM realm_mail_outbox WHERE recipient = '$realm_test_email' AND kind = 'reset_password' ORDER BY available_at DESC LIMIT 1")
    fi
    [ -n "$realm_test_recovery_url" ] || sleep 1
    realm_test_attempt=$((realm_test_attempt + 1))
done
[ -n "$realm_test_recovery_url" ] || {
    printf '%s\n' "realm production test: recovery URL was not delivered" >&2
    exit 1
}
realm_test_token=${realm_test_recovery_url#*token=}

realm_test_cookie_jar="$realm_test_root/browser-cookies"
realm_test_recovery_page="$realm_test_root/recovery.html"
realm_test_status=$(curl --silent --insecure --cookie-jar "$realm_test_cookie_jar" \
    --output "$realm_test_recovery_page" --write-out '%{http_code}' "$realm_test_recovery_url")
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: browser recovery page returned $realm_test_status" >&2
    exit 1
}
realm_test_csrf=$(sed -n 's/.*name="csrf_token" value="\([a-f0-9][a-f0-9]*\)".*/\1/p' "$realm_test_recovery_page" | head -n 1)
[ "${#realm_test_csrf}" -eq 64 ] || {
    printf '%s\n' "realm production test: browser recovery page omitted its CSRF token" >&2
    exit 1
}
realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    --cookie "$realm_test_cookie_jar" \
    --data-urlencode "token=$realm_test_token" \
    --data-urlencode "csrf_token=$realm_test_csrf" \
    --data-urlencode "password=$realm_test_new_password" \
    --data-urlencode "confirm_password=$realm_test_new_password" \
    "$realm_test_origin/recover")
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: browser password replacement returned $realm_test_status" >&2
    exit 1
}

realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' \
    --data "{\"name\":\"$realm_test_name\",\"password\":\"$realm_test_old_password\"}" \
    "$realm_test_origin/v1/sessions")
[ "$realm_test_status" = 401 ] || {
    printf '%s\n' "realm production test: old password was not revoked (status $realm_test_status)" >&2
    exit 1
}

DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/down.sh"
DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/up.sh" --no-build
realm_test_status=$(curl --silent --insecure --output /dev/null --write-out '%{http_code}' \
    -H 'Content-Type: application/json' \
    --data "{\"name\":\"$realm_test_name\",\"password\":\"$realm_test_new_password\"}" \
    "$realm_test_origin/v1/sessions")
[ "$realm_test_status" = 200 ] || {
    printf '%s\n' "realm production test: recovered account did not survive restart" >&2
    exit 1
}

if [ "$realm_test_disable_workers" -eq 0 ]; then
    realm_test_password_file="$realm_test_root/account-password"
    realm_test_ready_file="$realm_test_root/session-ready"
    realm_test_continue_file="$realm_test_root/session-continue"
    realm_test_session_output="$realm_test_root/session.json"
    realm_test_session_error="$realm_test_root/session.error"
    printf '%s\n' "$realm_test_new_password" > "$realm_test_password_file"
    DARK_MAGIC_REALM_ACCEPTANCE_URL="$realm_test_origin" \
    DARK_MAGIC_REALM_ACCEPTANCE_ACCOUNT="$realm_test_name" \
    DARK_MAGIC_REALM_ACCEPTANCE_PASSWORD_FILE="$realm_test_password_file" \
    DARK_MAGIC_REALM_ACCEPTANCE_CHARACTER=Realmhero \
    DARK_MAGIC_REALM_ACCEPTANCE_READY_FILE="$realm_test_ready_file" \
    DARK_MAGIC_REALM_ACCEPTANCE_CONTINUE_FILE="$realm_test_continue_file" \
    DARK_MAGIC_REALM_ACCEPTANCE_REALM_RESTART=1 \
        go run ./internal/dev/tools/realm_acceptance >"$realm_test_session_output" 2>"$realm_test_session_error" &
    realm_test_session_pid=$!
    realm_test_attempt=0
    while [ "$realm_test_attempt" -lt 120 ] && [ ! -f "$realm_test_ready_file" ]; do
        if ! kill -0 "$realm_test_session_pid" 2>/dev/null; then
            wait "$realm_test_session_pid" || true
            sed -n '1,120p' "$realm_test_session_error" >&2
            printf '%s\n' "realm production test: native session failed before checkpoint" >&2
            exit 1
        fi
        realm_test_attempt=$((realm_test_attempt + 1))
        sleep 0.25
    done
    [ -f "$realm_test_ready_file" ] || {
        kill "$realm_test_session_pid" 2>/dev/null || true
        wait "$realm_test_session_pid" || true
        printf '%s\n' "realm production test: native session did not reach reconnect barrier" >&2
        exit 1
    }
    realm_test_checkpoint_count=0
    realm_test_attempt=0
    while [ "$realm_test_attempt" -lt 120 ] && [ "$realm_test_checkpoint_count" -eq 0 ]; do
        realm_test_checkpoint_count=$(psql "$realm_test_database_url" -X -A -t -v ON_ERROR_STOP=1 \
            -c "SELECT count(*) FROM realm_game_checkpoints")
        [ "$realm_test_checkpoint_count" -gt 0 ] || sleep 0.25
        realm_test_attempt=$((realm_test_attempt + 1))
    done
    [ "$realm_test_checkpoint_count" -gt 0 ] || {
        kill "$realm_test_session_pid" 2>/dev/null || true
        wait "$realm_test_session_pid" || true
        printf '%s\n' "realm production test: live worker checkpoint was not persisted" >&2
        exit 1
    }

    # Model abrupt Realm authority loss, not an operator shutdown: SIGKILL
    # deliberately skips ProcessAllocator.Close so the exact worker generation
    # survives for startup fencing and checkpoint restoration.
    IFS= read -r realm_test_realm_pid < "$realm_test_runtime/realm.pid"
    kill -KILL "$realm_test_realm_pid"
    realm_test_attempt=0
    while [ "$realm_test_attempt" -lt 120 ] && kill -0 "$realm_test_realm_pid" 2>/dev/null; do
        realm_test_attempt=$((realm_test_attempt + 1))
        sleep 0.1
    done
    if kill -0 "$realm_test_realm_pid" 2>/dev/null; then
        kill "$realm_test_session_pid" 2>/dev/null || true
        wait "$realm_test_session_pid" || true
        printf '%s\n' "realm production test: terminated Realm process did not exit" >&2
        exit 1
    fi
    DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/up.sh" --no-build
    touch "$realm_test_continue_file"
    if ! wait "$realm_test_session_pid"; then
        sed -n '1,120p' "$realm_test_session_error" >&2
        printf '%s\n' "realm production test: native Realm session failed" >&2
        exit 1
    fi
    realm_test_active_leases=$(psql "$realm_test_database_url" -X -A -t -v ON_ERROR_STOP=1 \
        -c "SELECT count(*) FROM realm_character_leases")
    realm_test_active_allocations=$(psql "$realm_test_database_url" -X -A -t -v ON_ERROR_STOP=1 \
        -c "SELECT count(*) FROM realm_allocations WHERE state <> 'completed'")
    [ "$realm_test_active_leases" -eq 0 ] && [ "$realm_test_active_allocations" -eq 0 ] || {
        printf '%s\n' "realm production test: completed session retained a lease or allocation" >&2
        exit 1
    }
    DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/down.sh"
    DARK_MAGIC_CONFIG_DIR="$realm_test_config" "$realm_test_script_dir/up.sh" --no-build
    DARK_MAGIC_REALM_ACCEPTANCE_URL="$realm_test_origin" \
    DARK_MAGIC_REALM_ACCEPTANCE_ACCOUNT="$realm_test_name" \
    DARK_MAGIC_REALM_ACCEPTANCE_PASSWORD_FILE="$realm_test_password_file" \
    DARK_MAGIC_REALM_ACCEPTANCE_CHARACTER=Realmhero \
    DARK_MAGIC_REALM_ACCEPTANCE_MODE=verify \
    DARK_MAGIC_REALM_ACCEPTANCE_MIN_REVISION=2 \
        go run ./internal/dev/tools/realm_acceptance >/dev/null
    realm_test_passed=1
    if [ "$realm_test_mailpit" -eq 1 ]; then
        printf '%s\n' "realm production test: PASS (PostgreSQL, Mailpit browser actions, worker, QUIC, reconnect, checkpoint, commit, and restart verified)"
    else
        printf '%s\n' "realm production test: PASS (full PostgreSQL, worker, QUIC, reconnect, checkpoint, commit, and restart journey verified)"
    fi
    exit 0
fi

realm_test_passed=1
if [ "$realm_test_mailpit" -eq 1 ]; then
    printf '%s\n' "realm production test: PASS (PostgreSQL, SMTP verification, explicit login, and browser recovery verified)"
else
    printf '%s\n' "realm production test: PASS (pre-production schema, explicit login and browser recovery verified)"
fi
