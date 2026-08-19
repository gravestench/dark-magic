package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

// PresentationProfile identifies one manifest-owned presentation variant.
type PresentationProfile struct {
	ID        string
	Width     int
	Height    int
	ScreenIDs []string
}

// ResolvePresentationProfile resolves a requested profile, or the first
// declared profile when requested is empty, before native renderer startup.
func ResolvePresentationProfile(source fs.FS, requested string) (PresentationProfile, error) {
	data, err := fs.ReadFile(source, presentationManifest)
	if err != nil {
		return PresentationProfile{}, fmt.Errorf("content: read presentation manifest: %w", err)
	}

	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		return PresentationProfile{}, fmt.Errorf("content: decode presentation manifest: %w", err)
	}

	selected, profile, err := ApplyPresentationProfile(document, requested)
	if err != nil {
		return PresentationProfile{}, err
	}

	resolution, ok := selected["resolution"].(map[string]any)
	if !ok {
		return PresentationProfile{}, fmt.Errorf("content: profile %q has no resolution", requested)
	}

	width, widthOK := jsonInteger(resolution["width"])
	height, heightOK := jsonInteger(resolution["height"])

	id, idOK := profile["id"].(string)
	if !idOK || !widthOK || !heightOK || width <= 0 || height <= 0 {
		return PresentationProfile{}, fmt.Errorf("content: selected presentation profile is malformed")
	}

	result := PresentationProfile{ID: id, Width: width, Height: height}

	if screens, ok := profile["screens"].([]any); ok {
		for _, value := range screens {
			id, ok := value.(string)
			if !ok || strings.TrimSpace(id) == "" {
				return PresentationProfile{}, fmt.Errorf("content: profile %q has invalid screen scope", result.ID)
			}

			result.ScreenIDs = append(result.ScreenIDs, id)
		}
	}

	return result, nil
}

// ApplyPresentationProfile returns a deep-cloned manifest with the selected
// profile's resolution and sparse overrides merged into the base facts.
func ApplyPresentationProfile(document map[string]any, requested string) (map[string]any, map[string]any, error) {
	profiles, ok := document["supported_profiles"].([]any)
	if !ok || len(profiles) == 0 {
		return nil, nil, fmt.Errorf("content: presentation manifest has no supported profiles")
	}

	if strings.TrimSpace(requested) == "" {
		var err error

		requested, err = firstPresentationProfileID(profiles)
		if err != nil {
			return nil, nil, err
		}
	}

	profile := findPresentationProfile(profiles, requested)
	if profile == nil {
		return nil, nil, fmt.Errorf("content: unsupported presentation profile %q", requested)
	}

	if _, ok := profile["resolution"].(map[string]any); !ok {
		return nil, nil, fmt.Errorf("content: presentation profile %q has no resolution", requested)
	}

	// Clone before merging so profile selection cannot mutate the shared decoded manifest or a sibling profile.
	selected := cloneJSONObject(document)
	selected["resolution"] = cloneJSONValue(profile["resolution"])

	selected["active_profile"] = requested
	if overrides, ok := profile["overrides"].(map[string]any); ok {
		mergeJSONObject(selected, overrides)
	}

	return selected, profile, nil
}

// firstPresentationProfileID resolves the manifest-defined default without accepting a malformed first entry.
func firstPresentationProfileID(profiles []any) (string, error) {
	first, ok := profiles[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("content: first presentation profile is malformed")
	}

	id, ok := first["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("content: first presentation profile has no ID")
	}

	return id, nil
}

// findPresentationProfile returns the requested manifest object without copying so callers retain its declared scope.
func findPresentationProfile(profiles []any, requested string) map[string]any {
	for _, value := range profiles {
		candidate, ok := value.(map[string]any)
		if ok && candidate["id"] == requested {
			return candidate
		}
	}

	return nil
}

// jsonInteger accepts only lossless JSON integers so renderer dimensions cannot silently truncate fractional values.
func jsonInteger(value any) (int, bool) {
	switch number := value.(type) {
	case float64:
		integer := int(number)
		return integer, float64(integer) == number
	case json.Number:
		integer, err := number.Int64()
		return int(integer), err == nil
	default:
		return 0, false
	}
}

// cloneJSONObject deep-clones JSON-compatible objects so profile overrides cannot mutate the decoded source document.
func cloneJSONObject(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = cloneJSONValue(value)
	}

	return result
}

// cloneJSONValue recursively copies mutable JSON arrays and objects while reusing immutable scalar values.
func cloneJSONValue(value any) any {
	switch current := value.(type) {
	case map[string]any:
		return cloneJSONObject(current)
	case []any:
		result := make([]any, len(current))
		for index, item := range current {
			result[index] = cloneJSONValue(item)
		}

		return result
	default:
		return value
	}
}

// mergeJSONObject recursively overlays sparse profile objects while replacing scalar and array values by deep copy.
func mergeJSONObject(target, overrides map[string]any) {
	for key, value := range overrides {
		object, isObject := value.(map[string]any)
		if !isObject {
			target[key] = cloneJSONValue(value)
			continue
		}

		base, hasObjectBase := target[key].(map[string]any)
		if !hasObjectBase {
			target[key] = cloneJSONObject(object)
			continue
		}

		mergeJSONObject(base, object)
	}
}
