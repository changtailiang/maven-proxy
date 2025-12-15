// pkg/repository/proxy.go
package repository

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"maven-proxy/pkg/client"
	"maven-proxy/pkg/storage"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

type ProxyRepository struct {
	id      string
	mode    int
	cache   bool
	mirrors []string
	storage storage.Storage
	client  client.HTTPClient
}

func NewProxyRepository(id string, mode int, cache bool, mirrors []string, storage storage.Storage) *ProxyRepository {
	return &ProxyRepository{
		id:      id,
		mode:    mode,
		cache:   cache,
		mirrors: mirrors,
		storage: storage,
		client:  client.NewDefaultHTTPClient(0),
	}
}

func (r *ProxyRepository) ID() string {
	return r.id
}

func (r *ProxyRepository) Type() string {
	return "proxy"
}

func (r *ProxyRepository) CanRead() bool {
	return r.mode&4 == 4
}

func (r *ProxyRepository) CanWrite() bool {
	return false
}

func (r *ProxyRepository) Get(path string) ([]byte, int, http.Header, error) {
	// 先尝试从本地缓存读取
	if data, status, headers, err := r.storage.Read(path); err == nil {
		return data, status, headers, nil
	}

	// 从远程镜像获取
	for _, mirror := range r.mirrors {
		url := mirror + path
		data, status, headers, err := r.client.Get(url)
		log.Debugf("%d", status, headers, err)
		if err != nil {
			continue
		}

		if status == http.StatusOK {
			// 如果启用缓存且不是 metadata 文件，保存到本地
			if r.cache && !strings.Contains(strings.ToLower(path), "maven-metadata.xml") {
				r.storage.Write(path, data)
			}
			return data, status, headers, nil
		}
	}

	return nil, http.StatusNotFound, nil, fmt.Errorf("artifact not found in any mirror")
}

func (r *ProxyRepository) Put(path string, data []byte) error {
	return fmt.Errorf("proxy repository does not support write operations")
}

func (r *ProxyRepository) List(path string) ([]storage.FileInfo, error) {
	// Proxy 仓库的目录列表来自本地缓存
	return r.storage.List(path)
}
