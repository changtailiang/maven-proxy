// pkg/storage/prefixed.go
package storage

import (
	"net/http"
	"path/filepath"
)

// PrefixedStorage 为存储添加路径前缀
type PrefixedStorage struct {
	base   Storage
	prefix string
}

// NewPrefixedStorage 创建带前缀的存储包装器
func NewPrefixedStorage(base Storage, prefix string) *PrefixedStorage {
	return &PrefixedStorage{
		base:   base,
		prefix: prefix,
	}
}

func (s *PrefixedStorage) Read(path string) ([]byte, int, http.Header, error) {
	fullPath := filepath.Join(s.prefix, path)
	return s.base.Read(fullPath)
}

func (s *PrefixedStorage) Write(path string, data []byte) error {
	fullPath := filepath.Join(s.prefix, path)
	return s.base.Write(fullPath, data)
}

func (s *PrefixedStorage) List(path string) ([]FileInfo, error) {
	fullPath := filepath.Join(s.prefix, path)
	return s.base.List(fullPath)
}

func (s *PrefixedStorage) Exists(path string) bool {
	fullPath := filepath.Join(s.prefix, path)
	return s.base.Exists(fullPath)
}
