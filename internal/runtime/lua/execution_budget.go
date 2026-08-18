//go:build !race

package modruntime

import "time"

// Production invocations keep a tight runaway-script boundary.
const defaultExecutionBudget = time.Second
