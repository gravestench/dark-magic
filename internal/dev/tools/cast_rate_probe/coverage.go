package main

import (
	"fmt"
	"sort"
	"strings"
)

// coverageFor reports required observations without embedding an expected cast-rate formula in the evidence tool.
func coverageFor(cases []probeCase) coverage {
	required := requiredCoverageProfiles()

	for _, observed := range cases {
		key := coverageProfileKey(observed)
		if _, exists := required[key]; exists {
			required[key] = true
		}
	}

	return summarizeCoverage(required)
}

// requiredCoverageProfiles declares measurements around candidate transitions and weapon or sequence boundaries.
func requiredCoverageProfiles() map[string]bool {
	required := map[string]bool{}

	// Paired values on both sides of each candidate transition keep the observation useful even if it disproves that
	// transition. These rates declare required measurements only; they encode no expected delay.
	for _, rate := range []int{0, 8, 9, 19, 20, 36, 37, 62, 63, 104, 105, 199, 200} {
		required[fmt.Sprintf("sc-hth-fcr-%d", rate)] = false
	}

	for _, weapon := range []string{"1HS", "STF"} {
		for _, rate := range []int{0, 105} {
			required[fmt.Sprintf("sc-%s-fcr-%d", strings.ToLower(weapon), rate)] = false
		}
	}

	for _, rate := range []int{0, 105} {
		required[fmt.Sprintf("sq-hth-fcr-%d", rate)] = false
	}

	return required
}

// coverageProfileKey maps validated skills to the established SC or SQ coverage namespace.
func coverageProfileKey(observed probeCase) string {
	prefix := "sc"
	if observed.SkillID == 49 {
		prefix = "sq"
	}

	return fmt.Sprintf(
		"%s-%s-fcr-%d",
		prefix,
		strings.ToLower(observed.WeaponClass),
		observed.RawFasterCastRate,
	)
}

// summarizeCoverage sorts missing keys so incomplete reports remain deterministic across map iteration orders.
func summarizeCoverage(required map[string]bool) coverage {
	result := coverage{Complete: true}

	for key, present := range required {
		if present {
			continue
		}

		result.Complete = false
		result.Missing = append(result.Missing, key)
	}

	sort.Strings(result.Missing)

	return result
}
