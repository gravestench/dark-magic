# Raylib renderer adapter

This transitional adapter owns Raylib's main-thread loop and drains the
thread-safe `internal/rendercore` command boundary. The internal application
host constructs and starts it explicitly.

Lua code never receives this adapter. It uses `dm.render/v1`, whose retained
nodes and asset handles are scoped and validated by `internal/modruntime`.
