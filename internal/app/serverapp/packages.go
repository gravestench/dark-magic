package serverapp

import (
	"context"
	"errors"

	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// PackageProvider exposes only extension artifacts declared by the immutable
// runtime recipe, preventing clients from probing arbitrary cache entries.
type PackageProvider struct {
	recipe simulation.RuntimeRecipe
	cache  *modcache.Store
}

// NewPackageProvider validates the advertised recipe before it can be shared
// with clients and requires backing storage whenever extensions are present.
func NewPackageProvider(recipe simulation.RuntimeRecipe, cache *modcache.Store) (*PackageProvider, error) {
	if err := recipe.Validate(); err != nil {
		return nil, err
	}

	if len(recipe.Packages.Extensions) > 0 && cache == nil {
		return nil, errors.New("server: extension recipe requires a mod cache")
	}

	return &PackageProvider{recipe: recipe, cache: cache}, nil
}

// Recipe returns defensive copies of mutable recipe collections so transport
// consumers cannot change the provider's admission allowlist.
func (provider *PackageProvider) Recipe() simulation.RuntimeRecipe {
	recipe := provider.recipe
	recipe.Packages.Extensions = append([]simulation.RuntimePackage(nil), recipe.Packages.Extensions...)
	recipe.CapabilityVersions = cloneStrings(recipe.CapabilityVersions)

	return recipe
}

// ReadChunk serves a verified slice only when its identity and digest match an
// advertised extension. Cancellation wins before any cache access begins.
func (provider *PackageProvider) ReadChunk(
	ctx context.Context,
	request sessionquic.PackageRequest,
) (sessionquic.PackageChunk, error) {
	if err := ctx.Err(); err != nil {
		return sessionquic.PackageChunk{}, err
	}

	descriptor, advertised := provider.advertisedDescriptor(request)
	if !advertised || descriptor.ID == "" || provider.cache == nil {
		return sessionquic.PackageChunk{}, errors.New("server: package is not part of the advertised session recipe")
	}

	data, total, err := provider.cache.ReadVerifiedChunk(descriptor, request.Offset, request.Limit)
	if err != nil {
		return sessionquic.PackageChunk{}, err
	}

	return sessionquic.PackageChunk{
		ID:     request.ID,
		Digest: request.Digest,
		Offset: request.Offset,
		Total:  total,
		Data:   data,
	}, nil
}

// advertisedDescriptor translates the exact advertised package into the cache
// descriptor used for integrity verification; IDs alone are insufficient.
func (provider *PackageProvider) advertisedDescriptor(request sessionquic.PackageRequest) (modcache.Descriptor, bool) {
	for _, extension := range provider.recipe.Packages.Extensions {
		if extension.ID != request.ID || extension.Digest != request.Digest {
			continue
		}

		return modcache.Descriptor{
			ID:              extension.ID,
			Version:         extension.Version,
			Digest:          extension.Digest,
			Size:            extension.Size,
			Redistributable: extension.Redistributable,
		}, true
	}

	return modcache.Descriptor{}, false
}

// cloneStrings preserves nil as a meaningful absence while isolating a
// non-nil capability map from mutation by callers.
func cloneStrings(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}

	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}

	return result
}
