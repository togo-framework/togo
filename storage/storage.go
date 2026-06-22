// Package storage defines the blob storage contract. Implementations live in
// their own repos (e.g. github.com/togo-framework/storage) and register a provider.
package storage

// Storage is the blob storage contract.
type Storage interface {
	Put(path string, data []byte) error
	Get(path string) ([]byte, error)
	Delete(path string) error
	Path(path string) string
}
