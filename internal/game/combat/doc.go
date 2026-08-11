// Package combat contains deterministic, renderer-neutral combat vocabulary.
//
// Diablo II commonly keeps combat and life-like values with eight fractional
// bits. In plain language, it cuts one whole point into 256 tiny pieces. This
// package makes that scale visible so a caller cannot accidentally mix a whole
// display value with a raw simulation value.
//
// The package does not implement hit chance, resistance, absorb, PvP scaling,
// or other formulas whose exact ordering still needs verification. Those rules
// will consume these types after executable evidence establishes their policy.
package combat
