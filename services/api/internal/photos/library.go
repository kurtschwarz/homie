package photos

import (
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Library stores photo bytes on disk separately from their SQLite metadata.
type Library struct {
	root string
}

// NewLibrary creates a photo library rooted at path.
func NewLibrary(path string) (*Library, error) {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return nil, err
	}
	return &Library{root: path}, nil
}

// Name returns the stable, filesystem-safe filename for a photo id.
func (l *Library) Name(id uuid.UUID) string {
	return hex.EncodeToString(id[:]) + ".jpg"
}

// Save atomically replaces a photo file with the supplied bytes.
func (l *Library) Save(name string, src io.Reader) error {
	tmp := filepath.Join(l.root, name+".tmp")
	dst := filepath.Join(l.root, name)
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, src); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Remove deletes a photo file. Missing files are already removed.
func (l *Library) Remove(name string) error {
	err := os.Remove(filepath.Join(l.root, name))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// FS exposes the photo directory as a read-only filesystem.
func (l *Library) FS() fs.FS {
	return os.DirFS(l.root)
}
