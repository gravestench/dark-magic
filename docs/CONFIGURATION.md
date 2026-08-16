# Process environment files

Dark Magic keeps one private environment file for each composition root:

- `client.env`
- `server.env`
- `realm.env`

On first launch, each executable copies its embedded template into the platform
Dark Magic configuration directory. On macOS that is
`~/Library/Application Support/dark-magic`; on Windows it is the user's roaming
application-data directory; on Linux it is the XDG configuration directory.
Set `DARK_MAGIC_CONFIG_DIR` when a launcher, test, or service manager needs a
different directory.

Configuration precedence is, from strongest to weakest:

1. command-line flags;
2. variables already exported by the parent process;
3. the selected environment file; and
4. built-in command defaults.

All three commands accept `--env-file PATH`. Selecting another file changes the
file in step 3; it does not override an exported variable or a command-line
flag. The default template is still installed so an operator always has a
discoverable reference.

The files use a deliberately small dotenv grammar: blank lines, comments,
optional `export`, and `NAME=VALUE` with optional single or double quotes. They
do not evaluate shell expressions, interpolate variables, or run commands.
Dark Magic creates the directory with mode `0700` and the files with mode
`0600` where the platform supports POSIX modes.

To install all defaults without starting the processes:

```sh
go run ./internal/dev/tools/env_config --role all
```

To update a known template value while preserving comments:

```sh
go run ./internal/dev/tools/env_config --role client \
  --set MPQ_DIRECTORY=/path/to/operator-owned-mpqs
```

Only keys present in the corresponding embedded template may be changed by
this helper. Production secrets should come from a protected secret manager or
service environment; repository templates contain development-only values.

The Realm lifecycle scripts read the same `realm.env`. Set
`DARK_MAGIC_REALM_ENV_FILE` to select another file when invoking a script; the
Realm child process receives that exact file through `--env-file`.
`DARK_MAGIC_REALM_CHECKPOINT_INTERVAL` controls the minimum cadence for
persisting healthy workers' canonical game checkpoints and defaults to `15s`.
`DARK_MAGIC_REALM_OPERATOR_LISTEN` and
`DARK_MAGIC_REALM_OPERATOR_TOKEN_FILE` must be configured together to enable
the independent, loopback-only lifecycle API. The local scripts default it to
`127.0.0.1:6113` and generate the owner-only bearer credential beneath the
Realm runtime directory; neither endpoint nor credential is shared with player
sessions.

`MPQ_DIRECTORY` also defines the external asset-set portion of network session
identity. Dark Magic hashes the mounted files into a path-independent manifest
and caches only owner-local file metadata and digests under the platform cache
directory. Set `DARK_MAGIC_ASSET_SET_CACHE` to relocate that cache or
`DARK_MAGIC_ASSET_SET_REHASH=1` for an explicit full byte revalidation. These
are process-environment controls rather than committed machine-specific
defaults.
