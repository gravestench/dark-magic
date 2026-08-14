package serverapp

import (
	"context"
	"errors"

	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

type PackageProvider struct {
	recipe simulation.RuntimeRecipe
	cache  *modcache.Store
}

func NewPackageProvider(recipe simulation.RuntimeRecipe, cache *modcache.Store) (*PackageProvider, error) {
	if err := recipe.Validate(); err != nil {
		return nil, err
	}
	if len(recipe.Packages.Extensions) > 0 && cache == nil {
		return nil, errors.New("server: extension recipe requires a mod cache")
	}
	return &PackageProvider{recipe: recipe, cache: cache}, nil
}

func (provider *PackageProvider) Recipe() simulation.RuntimeRecipe {
	recipe := provider.recipe
	recipe.Packages.Extensions = append([]simulation.RuntimePackage(nil), recipe.Packages.Extensions...)
	recipe.CapabilityVersions = cloneStrings(recipe.CapabilityVersions)
	return recipe
}

func (provider *PackageProvider) ReadChunk(ctx context.Context, request sessionquic.PackageRequest) (sessionquic.PackageChunk, error) {
	if err := ctx.Err(); err != nil {
		return sessionquic.PackageChunk{}, err
	}
	var descriptor modcache.Descriptor
	for _, extension := range provider.recipe.Packages.Extensions {
		if extension.ID == request.ID && extension.Digest == request.Digest {
			descriptor = modcache.Descriptor{ID: extension.ID, Version: extension.Version, Digest: extension.Digest,
				Size: extension.Size, Redistributable: extension.Redistributable}
			break
		}
	}
	if descriptor.ID == "" || provider.cache == nil {
		return sessionquic.PackageChunk{}, errors.New("server: package is not part of the advertised session recipe")
	}
	data, total, err := provider.cache.ReadVerifiedChunk(descriptor, request.Offset, request.Limit)
	if err != nil {
		return sessionquic.PackageChunk{}, err
	}
	return sessionquic.PackageChunk{ID: request.ID, Digest: request.Digest, Offset: request.Offset, Total: total, Data: data}, nil
}

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
