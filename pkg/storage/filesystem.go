// pkg/storage/filesystem.go
package storage

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type FileSystemStorage struct {
	basePath string
}

func NewFileSystemStorage(basePath string) *FileSystemStorage {
	return &FileSystemStorage{basePath: basePath}
}

func (s *FileSystemStorage) Read(path string) ([]byte, int, http.Header, error) {
	fullPath := filepath.Join(s.basePath, path)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, 404, nil, err
	}

	headers := http.Header{
		"Content-Type": []string{getContentType(path)},
	}

	return data, 200, headers, nil
}

func (s *FileSystemStorage) Write(path string, data []byte) error {
	fullPath := filepath.Join(s.basePath, path)

	// 创建父目录
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	return ioutil.WriteFile(fullPath, data, 0o644)
}

func (s *FileSystemStorage) List(path string) ([]FileInfo, error) {
	fullPath := filepath.Join(s.basePath, path)

	// 读取目录内容
	entries, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	// 转换为 FileInfo 列表
	fileInfos := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		fileInfos = append(fileInfos, FileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			ModTime: entry.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	return fileInfos, nil
}

func (s *FileSystemStorage) Exists(path string) bool {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// getContentType 根据文件扩展名返回 MIME 类型
func getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jar":
		return "application/java-archive"
	case ".pom":
		return "application/xml"
	case ".xml":
		return "application/xml"
	case ".sha1", ".md5", ".sha256", ".sha512":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
