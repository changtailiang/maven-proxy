// internal/util/file.go
package util

import (
	"os"
	"path"
)

// CreateParentIfNotExist 创建文件的父目录（如果不存在）
func CreateParentIfNotExist(file string) error {
	dirPath := path.Dir(file)

	if stat, err := os.Stat(dirPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dirPath, 0o755); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return os.ErrNotExist
	}
	return nil
}

// CreateFileIfNotExist 创建文件（如果不存在）
func CreateFileIfNotExist(file string) error {
	if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		if err = CreateParentIfNotExist(file); err != nil {
			return err
		}
		if _, err := os.Create(file); err != nil {
			return err
		}
	}
	return nil
}
