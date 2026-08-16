# Realm accounts and security

## Current implementation

The PostgreSQL control plane now supports explicit signup, normalized unique
email ownership, inactive accounts, digest-only single-use verification and
recovery challenges, recovery-wide session revocation, and a transactional
leased mail outbox. A provider-neutral worker renders the two bounded account
templates and delivers them over SMTP with optional mandatory STARTTLS. The
older direct account creation path remains only for deterministic in-memory
fixtures.

The native client performs endpoint trust and compatibility checks, then always
shows explicit account-name and password entry. It never creates an account,
opens a browser authorization flow, refreshes a remembered Realm credential, or
logs in merely because an email challenge was completed.

## M23 account lifecycle

Production account creation is explicit:

1. The user supplies an account name, email address, and credential through the
   HTTPS account portal.
2. The Realm normalizes the email for comparison and creates an unverified
   account plus a single-use, expiring verification challenge.
3. A transactional outbox row commits with the account change.
4. The mail worker sends a verification link.
5. Verification consumes the stored challenge digest and activates login.
6. The player returns to the native client and logs in with the verified
   account credentials.

Password reset follows the same generic-response, digest-only, single-use,
bounded-expiry pattern. Passkeys may be added without changing Realm character
ownership.

The implemented JSON endpoints are:

- `POST /v1/accounts` for signup;
- `POST /v1/accounts/verify` for verification;
- `POST /v1/accounts/recovery` for the non-enumerating recovery request; and
- `POST /v1/accounts/recovery/complete` for password replacement and session
  revocation.

Challenge rows retain only SHA-256 digests. The raw action link exists in the
transactional mail payload until delivery because it is the content that must
reach the mailbox; it is never written to audit or ordinary logs.

The database enforces one account per canonical verified email. This limits one
account per mailbox; it cannot establish that two different mailboxes belong to
the same human.

## Credential rules

- Passwords exist transiently in the focused native-client Lua form and cross
  the Realm capability as one call argument. They are cleared after submission
  and never appear in copied status, retained renderer text, workers, logs, or
  audit fields.
- Password hashes use a deliberately expensive, versioned password-hashing
  policy that can be upgraded at successful login.
- Verification and reset tokens are random, short-lived, single-use, and stored
  only by digest.
- Browser sessions use secure, HTTP-only, same-site cookies and CSRF defenses.
- Native access tokens are short-lived and memory-only in the game client. A
  fresh process and each fresh Realm connection require credentials again.
- Login, signup, verification, reset, and authorization endpoints have bounded
  request sizes and independent rate limits.
- Public recovery responses do not reveal whether an email or account exists.

## Local email and HTTPS

The M23 local environment uses Mailpit as the SMTP sink and browser inbox. It
must exercise the same outbox and verification links as a production mail
provider; local testing does not silently mark accounts verified.

Developer-only `log` and `auto-verify` outbox senders provide a faster native
loop when Mailpit is unnecessary. They require an explicit loopback Realm bind
and emit prominent warnings. `log` exposes both action links;
`auto-verify` consumes signup verification and exposes only recovery links.
Their behavior is intentionally impossible on wildcard or non-loopback binds.

Local browser endpoints use trusted HTTPS names under `.test`, with certificates
from a developer-local CA. Production trust and edge configuration are deferred
to M44, but the application-level HTTPS and redirect rules are M23 behavior.

## Audit

Audit records include stable operation names, outcome, account and character
identifiers where authorized, peer/workload identity, error classification,
and correlation IDs. They exclude:

- passwords and password hashes;
- bearer, refresh, reset, or verification tokens;
- raw email contents;
- chat message contents;
- character save payloads; and
- protected game assets.

Account creation, verification, login, logout, recovery, credential changes,
session revocation, character lifecycle, channel moderation, game lifecycle,
worker allocation, and durable commits are auditable.

## Deferred to M44

Cloud secret managers, workload identity, edge WAF policy, managed email,
centralized audit export, and production key-rotation automation are deployment
work. They do not change the account lifecycle above.
