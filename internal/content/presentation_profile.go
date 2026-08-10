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
		first, ok := profiles[0].(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("content: first presentation profile is malformed")
		}
		requested, ok = first["id"].(string)
		if !ok || strings.TrimSpace(requested) == "" {
			return nil, nil, fmt.Errorf("content: first presentation profile has no ID")
		}
	}
	var profile map[string]any
	for _, value := range profiles {
		candidate, ok := value.(map[string]any)
		if ok && candidate["id"] == requested {
			profile = candidate
			break
		}
	}
	if profile == nil {
		return nil, nil, fmt.Errorf("content: unsupported presentation profile %q", requested)
	}
	if _, ok := profile["resolution"].(map[string]any); !ok {
		return nil, nil, fmt.Errorf("content: presentation profile %q has no resolution", requested)
	}
	selected := cloneJSONObject(document)
	selected["resolution"] = cloneJSONValue(profile["resolution"])
	selected["active_profile"] = requested
	if overrides, ok := profile["overrides"].(map[string]any); ok {
		mergeJSONObject(selected, overrides)
	}
	return selected, profile, nil
}

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

func cloneJSONObject(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = cloneJSONValue(value)
	}
	return result
}

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

func mergeJSONObject(target, overrides map[string]any) {
	for key, value := range overrides {
		if object, ok := value.(map[string]any); ok {
			base, ok := target[key].(map[string]any)
			if !ok {
				target[key] = cloneJSONObject(object)
			} else {
				mergeJSONObject(base, object)
			}
			continue
		}
		target[key] = cloneJSONValue(value)
	}
}
