# Local and single-machine Realm deployment

## Lifecycle scripts

The implemented local lifecycle entry points run the ordinary `cmd/realm` and
`cmd/server` composition roots; they do not substitute an in-process test
server:

```sh
make realm-up
make realm-down
make realm-fresh-install
make realm-drain-game GAME_ID=<opaque-game-id>
make realm-mailpit-up
make realm-mailpit-down
make realm-test-production
```

`realm-test-production` is a disposable end-to-end account acceptance. It owns
temporary Realm and PostgreSQL directories, initializes the baseline schema,
signs up and auto-verifies a unique account, logs in explicitly, initiates
recovery, submits the real browser reset form, proves the old password is
revoked, restarts both services, and proves the recovered credential persisted.
It removes the temporary profile during teardown and never uses the operator's
configured Realm database.

Supplying an operator-owned MPQ directory extends the same test through the
complete native single-machine journey:

```sh
DARK_MAGIC_REALM_TEST_MPQ_DIRECTORY=/path/to/mpqs make realm-test-production
```

That form builds an ordinary `cmd/server` worker, creates and selects a Realm
character, enters the public channel, creates a named game, joins over pinned
QUIC, submits and observes gameplay input, reconnects, proves a live checkpoint
reached PostgreSQL, kills and restarts Realm while the game remains live,
fences and replaces the surviving worker from that checkpoint, retargets the
client, commits the canonical character, retires the replacement worker, and
restarts Realm again to prove the new character revision persists. Failure artifacts
are retained under the printed temporary directory; successful disposable
profiles are removed. Protected assets never enter the database or test output.

Enable the real SMTP and browser-verification path with:

```sh
DARK_MAGIC_REALM_TEST_MAILPIT=1 make realm-test-production
```

This starts a uniquely named, loopback-only Mailpit container pinned to a
reviewed version, delivers both verification and recovery mail through SMTP,
discovers the links through Mailpit's HTTP integration endpoint, and submits
the same secure-cookie/CSRF-protected forms as a browser. Docker is only
required for this opt-in acceptance (Podman is supported through
`DARK_MAGIC_REALM_CONTAINER_CLI=podman`). `make realm-mailpit-up` and
`make realm-mailpit-down` provide a persistent manual inbox at
`http://127.0.0.1:8025`; point Realm SMTP at `127.0.0.1:1025` and select `smtp`
mail mode.

`realm-up` reads the same private `realm.env` as `cmd/realm`, builds owner-only
binaries under the configured runtime directory, starts Realm without an
interactive stdin shell, records its PID, appends structured output to its log,
and waits for the TLS `/v1/status` endpoint.
Realm supervises a separate game-worker process for each named game. Set
`MPQ_DIRECTORY` before bring-up for portal graphics and production game worlds.

The first invocation copies the embedded Realm template to the platform Dark
Magic configuration directory. Its local-development defaults start an
isolated PostgreSQL cluster on loopback port `55432` and therefore exercise the
production account, character, session, challenge, outbox, audit,
game-directory, and allocation repositories. PostgreSQL data lives below the configured Realm runtime
directory and is stopped by `realm-down`. The scripts require locally installed
PostgreSQL command-line tools and never attach to or modify an unrelated
database cluster.

The local operator API is separately bound to `127.0.0.1:6113`. Its generated
owner-only token lives under the configured Realm runtime directory. Use the
drain target rather than copying that credential into a shell command.

Set `DARK_MAGIC_REALM_CHECKPOINT_INTERVAL` to change the default `15s` minimum
cadence for canonical game checkpoints. Health probes remain independently
configured, and a failed checkpoint retries on the next probe.

Set `DARK_MAGIC_REALM_MANAGE_POSTGRES=0` to use an externally managed database;
`DARK_MAGIC_REALM_POSTGRES_URL` remains required in every deployment. The local template defaults to development
`auto-verify` mail: signup is verified automatically while password-recovery
URLs are printed to the Realm log. Set the mode to `log` to print and manually
consume both kinds of links, or configure SMTP normally.

Fresh install is intentionally guarded. Realm operational state such as its TLS
identity is moved into the configured runtime backup directory rather than deleted. Resetting a
PostgreSQL schema additionally requires
`DARK_MAGIC_REALM_ALLOW_DATABASE_RESET=1`, and the script verifies through
PostgreSQL itself that the server is loopback/local and the database is not
`postgres` or a template database. It refuses remote database resets. Use
`--yes` for non-interactive local automation and `--no-start` to prepare clean
state without bringing Realm back up.

Common local overrides are:

```text
DARK_MAGIC_REALM_RUNTIME_DIR
DARK_MAGIC_REALM_DATA
DARK_MAGIC_REALM_LISTEN
DARK_MAGIC_REALM_URL
DARK_MAGIC_REALM_LOG_LEVEL
DARK_MAGIC_REALM_POSTGRES_URL
DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE
DARK_MAGIC_REALM_SMTP_ADDRESS
DARK_MAGIC_REALM_GAME_HOST
DARK_MAGIC_REALM_DISABLE_WORKERS
```

See [process environment files](../CONFIGURATION.md) for file locations,
precedence, permissions, and `--env-file` behavior. Lifecycle scripts accept an
alternate file through `DARK_MAGIC_REALM_ENV_FILE`.

Run `scripts/realm/up.sh --print-config` to inspect the resolved non-secret
configuration. Database URLs and SMTP passwords are deliberately omitted.

## Status

`cmd/realm` always uses PostgreSQL account/session, verified-account challenge,
transactional mail, character/lease, and audit repositories. The connection is
required through `--postgres-url` (or `DARK_MAGIC_REALM_POSTGRES_URL`). It can supervise ordinary
`cmd/server --realm-worker` processes for public
create/join handoffs. Workers prepare real d2legacy world state and clients
consume the resulting assignment through the native QUIC controller. Native
leave, reconnect-grace expiry, health/capacity probes, dead-worker cleanup, and
empty-game process release now cross the private worker boundary. The full
disposable production acceptance covers Mailpit-backed browser account actions
and safely fenced active-game Realm restart recovery. Live failed-worker
restoration, authenticated client reassignment, and periodic recovery
checkpoints are exercised against real PostgreSQL.
Durable allocation/membership state, atomic departure receipts, and conservative
restart cleanup are already exercised against real PostgreSQL.

The embedded pre-production baseline schema initializes transactionally under a
fixed advisory lock. There is deliberately no incremental migration ledger while
the complete game and Realm storage contracts are still changing. Incompatible
local changes use `make realm-fresh-install`; immutable forward migrations begin
only after production stability. A local developer connection may use
`sslmode=disable` only for a loopback disposable database; operator connections
must authenticate and verify TLS according to their PostgreSQL deployment.

Example repository selection (credentials are operator-local and must not be
committed):

```sh
DARK_MAGIC_REALM_POSTGRES_URL='postgres://realm@127.0.0.1:5432/realm?sslmode=disable' \
DARK_MAGIC_REALM_ACCOUNT_URL='https://accounts.dark-magic.test' \
DARK_MAGIC_REALM_SMTP_ADDRESS='127.0.0.1:1025' \
DARK_MAGIC_REALM_SMTP_FROM='realm@dark-magic.test' \
  go run ./cmd/realm --worker-executable ./server
```

`DARK_MAGIC_REALM_SMTP_USERNAME` and `DARK_MAGIC_REALM_SMTP_PASSWORD` configure
authenticated SMTP. The password has no command-line flag so it does not appear
in process listings. Use `--smtp-require-tls` outside the local Mailpit profile;
Mailpit deliberately exercises the same transactional outbox without pretending
to be a production mail provider.

For a lighter local loop, omit SMTP and bind Realm explicitly to loopback:

```sh
go run ./cmd/realm \
  --listen 127.0.0.1:6112 \
  --postgres-url 'postgres://realm@127.0.0.1:5432/realm?sslmode=disable' \
  --account-mail-mode log
```

`log` prints the verification or recovery action URL. `auto-verify` consumes
signup verification automatically and still prints password-recovery URLs.
These modes expose bearer links and are rejected unless `--listen` names a
loopback address; neither is a production default. `smtp` is selected
automatically when `DARK_MAGIC_REALM_SMTP_ADDRESS` is present, otherwise mail
delivery defaults to `disabled`. The mode may also be set through
`DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE`.

## Required topology

```text
Dark Magic clients
  |-- HTTPS --> Realm API and account portal
  |-- QUIC  --> allocated cmd/server process

Realm
  |-- PostgreSQL
  |-- SMTP
  |-- private authenticated worker control

Workers
  |-- read-only legally supplied game assets
```

One machine may run every process. It remains a real Realm: accounts and
characters are server-owned, and clients use the same authentication,
admission, and commit paths as a cluster.

## M23 development profile

The native development profile uses:

- local Realm and `cmd/server` binaries;
- PostgreSQL for production repository behavior;
- Mailpit for SMTP and browser-visible verification/reset messages;
- a developer-local certificate authority for trusted HTTPS;
- process allocation with a configured local UDP port range; and
- `MPQ_DIRECTORY` pointing to a legally supplied, read-only asset directory.

Focused unit tests continue using in-memory repositories and an in-process
worker adapter. They do not replace the process/PostgreSQL acceptance.

## Local names and HTTPS

Browser integration may use reserved `.test` names:

```text
accounts.dark-magic.test
realm.dark-magic.test
mail.dark-magic.test
game.dark-magic.test
```

A developer may map them to loopback in `/etc/hosts`:

```text
127.0.0.1 accounts.dark-magic.test realm.dark-magic.test mail.dark-magic.test game.dark-magic.test
```

Install `mkcert` or an equivalent developer-owned CA when testing with those
names so the system browser exercises trusted HTTPS. The automated loopback
acceptance instead trusts its isolated ephemeral certificate explicitly. No CA
private key, generated leaf key, account credential, or host-specific asset
path is committed.

Mailpit receives the ordinary transactional outbox mail through SMTP. A local
signup still requires opening and consuming its verification link; local mode
does not silently activate the account.

## Single-host operator profile

After the native path is proven, an optional Compose/system-service profile may
package Realm, PostgreSQL, mail worker, and worker supervision for operators.
It must expose ordinary configuration for:

- public HTTPS address and certificates;
- PostgreSQL connection and backup destination;
- SMTP provider;
- worker address and UDP port range;
- asset directory and validated `AssetSetID`;
- Realm and worker workload keys;
- resource and game-capacity limits; and
- structured logs and health endpoints.

It must not require Kubernetes or Cloudflare.

## M23 acceptance

The local Realm is complete only when an automated acceptance can:

1. start clean PostgreSQL, Mailpit, Realm, and worker supervision;
2. create and verify an account through HTTPS;
3. create and select a Realm character;
4. enter the public Realm channel;
5. create a uniquely named game;
6. allocate a separate `cmd/server` process;
7. admit the selected character and play through the real QUIC path;
8. checkpoint and commit canonical character changes;
9. reconnect during the allowed window;
10. restart Realm without losing durable state; and
11. terminate cleanly without an orphaned worker or lease.

Synthetic redistributable records remain available for CI. Real-asset acceptance
is optional and consumes only an operator-supplied directory.

The ordinary Go suite supplies the destructive/adversarial half of this gate:
it checks complete-identity mismatches at every compatibility boundary,
tampered packages, checkpoints, and replay evidence, interrupted extension
delivery followed by a clean retry, duplicate live character admission,
explicit sequential replay-schema migration, failed-worker replacement, and
fail-closed startup recovery. These cases use the same control-plane,
repository, package-cache, and worker contracts as the production harness; they
do not need protected assets or a long-running local service. The production
harness complements them by proving the successful PostgreSQL/process/QUIC
restart and recovery path.

Kubernetes is a separate M44 integration profile described in
[cloud deployment](CLOUD_DEPLOYMENT.md).
