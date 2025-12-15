// pkg/storage/storage.go
package storage

import (
	"net/http"
	"time"
)

// Storage 存储接口定义
type Storage interface {
	// Read 读取文件，返回数据、状态码、响应头和错误
	Read(path string) ([]byte, int, http.Header, error)

	// Write 写入文件
	Write(path string, data []byte) error

	// List 列出目录内容
	List(path string) ([]FileInfo, error)

	// Exists 检查文件或目录是否存在
	Exists(path string) bool
}

// FileInfo 文件或目录的元信息
type FileInfo struct {
	Name    string    // 文件或目录名称
	Size    int64     // 文件大小（字节），目录为 0
	ModTime time.Time // 最后修改时间
	IsDir   bool      // 是否为目录
}
