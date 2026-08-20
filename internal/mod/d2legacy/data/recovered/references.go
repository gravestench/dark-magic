package recovered

import (
	"sort"
	"strings"
)

// ValidateReferences checks joins from recovered rows into mounted Sounds.txt
// and localization tables. Missing references remain diagnostics because game
// editions and language packs legitimately contain different subsets.
func ValidateReferences(
	snapshot Snapshot,
	soundNames map[string]struct{},
	text func(string) (string, error),
) []ReferenceIssue {
	issues := validateSpeechReferences(snapshot.Speech, soundNames, text)
	issues = append(issues, validateQuestReferences(snapshot.Quests, text)...)

	// Stable kind/identifier ordering makes diagnostics reproducible across the
	// independently ordered speech and quest inputs.
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Kind == issues[j].Kind {
			return issues[i].Identifier < issues[j].Identifier
		}

		return issues[i].Kind < issues[j].Kind
	})

	return issues
}

// validateSpeechReferences checks both the case-insensitive sound join and its
// localized caption while preserving source speech order in intermediate data.
func validateSpeechReferences(
	speechRows []Speech,
	soundNames map[string]struct{},
	text func(string) (string, error),
) []ReferenceIssue {
	var issues []ReferenceIssue

	for _, speech := range speechRows {
		if _, found := soundNames[strings.ToLower(speech.Sound)]; !found {
			issues = append(issues, ReferenceIssue{
				Kind: "sound", Identifier: speech.Sound, Detail: "not present in Sounds.txt",
			})
		}

		if text == nil {
			continue
		}

		if _, err := text(speech.StringKey); err != nil {
			issues = append(issues, ReferenceIssue{
				Kind: "string", Identifier: speech.StringKey, Detail: err.Error(),
			})
		}
	}

	return issues
}

// validateQuestReferences flattens title, stage, and alternate localization
// keys without changing their source order before final diagnostic sorting.
func validateQuestReferences(
	quests []Quest,
	text func(string) (string, error),
) []ReferenceIssue {
	if text == nil {
		return nil
	}

	var issues []ReferenceIssue

	for _, quest := range quests {
		keys := []string{quest.TitleStringKey}
		for _, stage := range quest.Stages {
			keys = append(keys, stage.StringKey)
			keys = append(keys, stage.Alternates...)
		}

		for _, key := range keys {
			if key == "" {
				continue
			}

			if _, err := text(key); err != nil {
				issues = append(issues, ReferenceIssue{
					Kind: "string", Identifier: key, Detail: err.Error(),
				})
			}
		}
	}

	return issues
}
