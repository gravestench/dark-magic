# Contributing

Dark Magic uses explicit construction and lifecycle ownership. New long-lived
native components implement `Start(context.Context)` and
`Stop(context.Context)` and are registered with `internal/host`. Required
dependencies are constructor arguments expressed as narrow consumer-owned
interfaces.

Lua integrations are versioned capability modules in `internal/modruntime`.
Never export the host, runtime manager, renderer backend, or mutable Go objects
directly. Native resources use checked handles and must be attached to the
active script scope so disable, reload, and shutdown release them.

Short-lived screens and overlays belong to `internal/navigation`, not the
application component graph. Renderer and audio mutations cross their
thread-safe command boundaries and are drained by the native owner thread.

Use [internal/service_template](internal/service_template) as the component
starting point. Run `make test` and `make test-race` before submitting changes.
