//go:build race

package modruntime

import "time"

// Race instrumentation can make the Lua-heavy authority graph take more than
// one wall-clock second to load even when it is making steady progress. Keep
// the same invocation guard, scaled only for race-instrumented binaries.
const defaultExecutionBudget = 10 * time.Second
