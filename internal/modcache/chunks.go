package modcache

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const MaxChunkBytes = 64 << 10

// Has authenticates the cached blob rather than merely checking its path. A
// true result describes the blob at verification time; later consumers still
// enforce their own immutable-cache or revalidation boundary.
func (store *Store) Has(descriptor Descriptor) (bool, error) {
	if _, err := store.verifyDescriptor(descriptor); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ReadVerifiedChunk serves an exact descriptor selected by an already
// authenticated session recipe. Unlike profile lookup, it supports multiple
// content-addressed versions of one extension ID coexisting in the cache.
func (store *Store) ReadVerifiedChunk(
	descriptor Descriptor,
	offset int64,
	limit int,
) ([]byte, int64, error) {
	if !descriptor.Redistributable || !validID(descriptor.ID) || !validDigest(descriptor.Digest) ||
		descriptor.Size <= 0 || offset < 0 || limit <= 0 || limit > MaxChunkBytes {
		return nil, 0, errors.New("modcache: invalid package chunk request")
	}

	file, err := os.Open(store.blobPath(descriptor.Digest))
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil || info.Size() != descriptor.Size {
		return nil, 0, errors.New("modcache: package blob differs from descriptor")
	}

	if offset > descriptor.Size {
		return nil, 0, errors.New("modcache: package chunk offset exceeds size")
	}

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, 0, fmt.Errorf("modcache: seek package chunk: %w", err)
	}

	remaining := descriptor.Size - offset
	if int64(limit) > remaining {
		limit = int(remaining)
	}

	data := make([]byte, limit)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, 0, fmt.Errorf("modcache: read package chunk: %w", err)
	}

	return data, descriptor.Size, nil
}
