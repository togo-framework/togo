// Package storage provides a file storage abstraction. The default is the local
// filesystem; S3/Supabase Storage ship as plugins that implement Storage.
package storage

import (
	"os"
	"path/filepath"
)

// Storage is the blob storage contract.
type Storage interface {
	Put(path string, data []byte) error
	Get(path string) ([]byte, error)
	Delete(path string) error
	Path(path string) string
}

type fsStore struct{ root string }

// NewFS returns a filesystem-backed store rooted at root.
func NewFS(root string) Storage { return &fsStore{root: root} }

func (s *fsStore) full(p string) string { return filepath.Join(s.root, filepath.Clean("/"+p)) }

func (s *fsStore) Put(path string, data []byte) error {
	full := s.full(path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func (s *fsStore) Get(path string) ([]byte, error) { return os.ReadFile(s.full(path)) }

func (s *fsStore) Delete(path string) error { return os.Remove(s.full(path)) }

func (s *fsStore) Path(path string) string { return s.full(path) }
