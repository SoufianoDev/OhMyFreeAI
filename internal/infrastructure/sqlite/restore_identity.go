package sqlite

// RestoreContentIdentity is measured from an already-open, no-follow file
// handle. It is independent of a pathname after the handle is opened.
type RestoreContentIdentity struct {
	Size   uint64
	SHA256 string
}

func (identity RestoreContentIdentity) matches(sha256 string, size uint64) bool {
	return identity.Size == size && identity.SHA256 == sha256
}
