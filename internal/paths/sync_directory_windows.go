package paths

// SyncDirectory is a no-op on Windows. Go cannot open a directory with the
// write-compatible handle required by FlushFileBuffers, and os.File.Sync on a
// directory consequently returns ERROR_ACCESS_DENIED. The temporary file is
// still flushed before the atomic rename.
func SyncDirectory(string) error {
	return nil
}
