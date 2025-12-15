// pkg/repository/hosted.go
package repository

import (
	"net/http"

	"maven-proxy/pkg/storage"
)

type HostedRepository struct {
	id      string
	mode    int
	storage storage.Storage
}

func NewHostedRepository(id string, mode int, storage storage.Storage) *HostedRepository {
	return &HostedRepository{
		id:      id,
		mode:    mode,
		storage: storage,
	}
}

func (r *HostedRepository) ID() string {
	return r.id
}

func (r *HostedRepository) Type() string {
	return "hosted"
}

func (r *HostedRepository) CanRead() bool {
	return r.mode&4 == 4
}

func (r *HostedRepository) CanWrite() bool {
	return r.mode&2 == 2
}

func (r *HostedRepository) Get(path string) ([]byte, int, http.Header, error) {
	return r.storage.Read(path)
}

func (r *HostedRepository) Put(path string, data []byte) error {
	return r.storage.Write(path, data)
}

func (r *HostedRepository) List(path string) ([]storage.FileInfo, error) {
	return r.storage.List(path)
}
