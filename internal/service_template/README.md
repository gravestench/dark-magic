# Internal component template

Construct required dependencies explicitly with `New`, implement
`Start(context.Context)` and `Stop(context.Context)`, and register the result
with `internal/host`. Keep dependency interfaces narrow and owned by the
consumer.

Lua access is not added to a component directly. Expose a versioned capability
module through `internal/modruntime`, returning checked handles for native
resources. A script component owns all callbacks, handles, subscriptions, and
render nodes through its runtime scope.
