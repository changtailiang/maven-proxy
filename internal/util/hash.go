// internal/util/hash.go
package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// GenerateHash 为文件生成校验和文件
func GenerateHash(file string) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		dir, err := ioutil.ReadDir(file)
		if err != nil {
			return err
		}
		for _, info := range dir {
			if err = GenerateHash(path.Join(file, info.Name())); err != nil {
				return err
			}
		}
		return nil
	}

	ext := path.Ext(file)
	if ext != ".xml" && ext != ".jar" && ext != ".pom" {
		return nil
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err = CreateHashFile(file, "md5", bytes); err != nil {
		return err
	}
	if err = CreateHashFile(file, "sha1", bytes); err != nil {
		return err
	}
	return nil
}

// CreateHashFile 创建指定类型的校验和文件
func CreateHashFile(file string, hashType string, bytes []byte) error {
	hashFile := fmt.Sprintf("%s.%s", file, hashType)
	if exist, err := FileExists(hashFile); err != nil {
		return err
	} else if !exist {
		if err = ioutil.WriteFile(hashFile, GetHash(bytes, hashType), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// GetHash 计算文件的校验和
func GetHash(file []byte, hashType string) []byte {
	switch hashType {
	case "md5":
		return []byte(fmt.Sprintf("%x", md5.Sum(file)))
	case "sha1":
		return []byte(fmt.Sprintf("%x", sha1.Sum(file)))
	case "sha256":
		return []byte(fmt.Sprintf("%x", sha256.Sum256(file)))
	case "sha512":
		return []byte(fmt.Sprintf("%x", sha512.Sum512(file)))
	default:
		return nil
	}
}

// FileExists 检查文件是否存在
func FileExists(file string) (bool, error) {
	if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}
