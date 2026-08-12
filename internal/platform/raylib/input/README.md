# Raylib input adapter

This transitional adapter samples Raylib input for the internal application
host. New Lua code consumes immutable snapshots through the versioned
`engine.input/v1` capability in `internal/modruntime`.

The renderer dependency is passed explicitly to `New`; no runtime registry or
dependency polling is involved.
