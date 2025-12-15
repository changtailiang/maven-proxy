// pkg/repository/repository.go
package repository

import (
	"net/http"

	"maven-proxy/pkg/storage"
)

// Repository 仓库接口定义
type Repository interface {
	// ID 返回仓库唯一标识
	ID() string

	// Type 返回仓库类型 (hosted, proxy, group)
	Type() string

	// CanRead 检查是否有读权限
	CanRead() bool

	// CanWrite 检查是否有写权限
	CanWrite() bool

	// Get 获取文件，返回数据、状态码、响应头和错误
	Get(path string) ([]byte, int, http.Header, error)

	// Put 上传文件
	Put(path string, data []byte) error

	// List 列出目录内容
	List(path string) ([]storage.FileInfo, error)
}
