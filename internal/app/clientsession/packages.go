package clientsession

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

const (
	maxSessionExtensions          = 64
	maxSessionPackageBytes  int64 = 256 << 20
	maxSessionDownloadBytes int64 = 2 << 30
)

// PrepareSelfHostedExtensions authenticates the advertised host before asking
// for its recipe and redistributable packages. It deliberately performs no
// profile admission; character/session timers begin only after recomposition.
func PrepareSelfHostedExtensions(
	ctx context.Context,
	assignment SelfHostedAssignment,
	tlsConfig *tls.Config,
	store *modcache.Store,
	localBase simulation.RuntimePackage,
) (simulation.RuntimeRecipe, error) {
	realmAssignment := realm.JoinAssignment{
		GameID:   assignment.GameID,
		Endpoint: assignment.Endpoint,
		Runtime:  assignment.Runtime,
	}

	return PrepareExtensions(ctx, realmAssignment, tlsConfig, store, localBase)
}

// PrepareExtensions authenticates a Realm-issued worker assignment before
// acquiring its exact redistributable extension recipe. The one-use gameplay
// ticket is deliberately not consumed during package preparation.
func PrepareExtensions(
	ctx context.Context,
	assignment realm.JoinAssignment,
	tlsConfig *tls.Config,
	store *modcache.Store,
	localBase simulation.RuntimePackage,
) (simulation.RuntimeRecipe, error) {
	transport, _, err := dialVerified(
		ctx,
		assignment.GameID,
		assignment.Endpoint,
		assignment.Runtime,
		tlsConfig,
	)
	if err != nil {
		return simulation.RuntimeRecipe{}, err
	}

	defer func() { _ = transport.Close() }()

	return AcquireExtensions(ctx, transport, store, localBase)
}

// ExtensionTransport is the authenticated, bounded package-delivery surface
// needed during runtime recomposition. Keeping this boundary smaller than the
// complete gameplay client also lets acceptance tests terminate a transfer at
// an exact byte offset and prove that a later connection can retry safely.
type ExtensionTransport interface {
	Recipe(context.Context) (simulation.RuntimeRecipe, error)
	PackageChunk(context.Context, sessionquic.PackageRequest) (sessionquic.PackageChunk, error)
}

// AcquireExtensions asks the authenticated game transport for its exact recipe
// and installs missing redistributable extension blobs through quarantine.
// Callers must recompose client runtimes from the returned recipe before join.
func AcquireExtensions(
	ctx context.Context,
	transport ExtensionTransport,
	store *modcache.Store,
	localBase simulation.RuntimePackage,
) (simulation.RuntimeRecipe, error) {
	if ctx == nil || transport == nil {
		return simulation.RuntimeRecipe{}, errors.New("client session: package acquisition requires context and transport")
	}

	recipe, err := transport.Recipe(ctx)
	if err != nil {
		return simulation.RuntimeRecipe{}, err
	}

	if err := validateExtensionRecipe(recipe, store, localBase); err != nil {
		return simulation.RuntimeRecipe{}, err
	}

	var downloadBytes int64
	for _, extension := range recipe.Packages.Extensions {
		if err := reserveExtensionBytes(&downloadBytes, extension.Size); err != nil {
			return simulation.RuntimeRecipe{}, err
		}

		if err := acquireExtension(ctx, transport, store, extension); err != nil {
			return simulation.RuntimeRecipe{}, err
		}
	}

	return recipe, nil
}

// validateExtensionRecipe rejects incompatible base content and recipes the bounded cache cannot serve.
func validateExtensionRecipe(
	recipe simulation.RuntimeRecipe,
	store *modcache.Store,
	localBase simulation.RuntimePackage,
) error {
	if recipe.Packages.Base != localBase {
		return errors.New("client session: server requires a different built-in d2legacy distribution")
	}

	if len(recipe.Packages.Extensions) > 0 && store == nil {
		return errors.New("client session: server requires extensions but no mod cache is available")
	}

	if len(recipe.Packages.Extensions) > maxSessionExtensions {
		return errors.New("client session: server recipe contains too many extensions")
	}

	return nil
}

// reserveExtensionBytes enforces per-package and aggregate download budgets without integer overflow.
func reserveExtensionBytes(total *int64, size int64) error {
	if size > maxSessionPackageBytes || *total > maxSessionDownloadBytes-size {
		return errors.New("client session: server extension recipe exceeds download limits")
	}

	*total += size

	return nil
}

// acquireExtension skips verified cache hits and installs missing redistributable bytes through quarantine.
func acquireExtension(
	ctx context.Context,
	transport ExtensionTransport,
	store *modcache.Store,
	extension simulation.RuntimePackage,
) error {
	descriptor := extensionDescriptor(extension)

	installed, err := store.Has(descriptor)
	if err != nil || installed {
		return err
	}

	if !extension.Redistributable {
		return fmt.Errorf(
			"client session: required extension %q is missing and not redistributable",
			extension.ID,
		)
	}

	reader, writer := io.Pipe()
	downloadDone := make(chan error, 1)

	go func() {
		downloadDone <- downloadExtension(ctx, transport, writer, extension)
	}()

	_, installErr := store.InstallVerified(ctx, reader, descriptor)
	_ = reader.Close()
	downloadErr := <-downloadDone

	if err := errors.Join(installErr, downloadErr); err != nil {
		return fmt.Errorf("client session: acquire extension %q: %w", extension.ID, err)
	}

	return nil
}

// extensionDescriptor converts the authenticated recipe entry into the cache's verification contract.
func extensionDescriptor(extension simulation.RuntimePackage) modcache.Descriptor {
	return modcache.Descriptor{
		ID:              extension.ID,
		Version:         extension.Version,
		Digest:          extension.Digest,
		Size:            extension.Size,
		Redistributable: extension.Redistributable,
	}
}

// downloadExtension streams bounded chunks in order and retries only explicit server rate limits.
func downloadExtension(
	ctx context.Context,
	transport ExtensionTransport,
	destination *io.PipeWriter,
	pkg simulation.RuntimePackage,
) error {
	defer func() { _ = destination.Close() }()

	retryDelay := 10 * time.Millisecond

	for offset := int64(0); offset < pkg.Size; {
		chunk, err := transport.PackageChunk(ctx, extensionChunkRequest(pkg, offset))
		if err != nil {
			if !isPackageRateLimit(err) {
				return err
			}

			if !waitPackageRetry(ctx, retryDelay) {
				return ctx.Err()
			}

			retryDelay = min(250*time.Millisecond, retryDelay*2)

			continue
		}

		retryDelay = 10 * time.Millisecond

		if chunk.Total != pkg.Size || len(chunk.Data) == 0 || int64(len(chunk.Data)) > pkg.Size-offset {
			return sessionquic.ErrWire
		}

		if _, err := destination.Write(chunk.Data); err != nil {
			return err
		}

		offset += int64(len(chunk.Data))
	}

	return nil
}

// extensionChunkRequest caps each request to the protocol limit and remaining advertised bytes.
func extensionChunkRequest(pkg simulation.RuntimePackage, offset int64) sessionquic.PackageRequest {
	limit := sessionquic.MaxPackageChunkBytes
	if remaining := pkg.Size - offset; int64(limit) > remaining {
		limit = int(remaining)
	}

	return sessionquic.PackageRequest{
		ID:     pkg.ID,
		Digest: pkg.Digest,
		Offset: offset,
		Limit:  limit,
	}
}

// isPackageRateLimit permits retry only for the authenticated protocol's explicit backpressure signal.
func isPackageRateLimit(err error) bool {
	var remote *sessionquic.RemoteError

	return errors.As(err, &remote) && remote.Message == sessionquic.PackageRateLimitMessage
}

// waitPackageRetry makes exponential backoff promptly cancelable and avoids leaking timers.
func waitPackageRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
