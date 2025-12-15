// pkg/client/client.go
package client

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPClient HTTP 客户端接口
type HTTPClient interface {
	// Get 发起 GET 请求，返回响应数据、状态码、响应头和错误
	Get(url string) ([]byte, int, http.Header, error)

	// Download 下载文件到指定路径，支持断点续传
	Download(url string, destPath string) (int, http.Header, error)
}

// DefaultHTTPClient 默认 HTTP 客户端实现
type DefaultHTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// NewDefaultHTTPClient 创建默认 HTTP 客户端
func NewDefaultHTTPClient(timeout time.Duration) *DefaultHTTPClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Get 发起 GET 请求
func (c *DefaultHTTPClient) Get(url string) ([]byte, int, http.Header, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, fmt.Errorf("read response body failed: %w", err)
	}

	return data, resp.StatusCode, resp.Header, nil
}

// Download 下载文件到指定路径，支持断点续传
func (c *DefaultHTTPClient) Download(url string, destPath string) (int, http.Header, error) {
	// 检查本地文件是否存在，获取已下载的大小
	var offset int64 = 0
	if stat, err := os.Stat(destPath); err == nil {
		offset = stat.Size()
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("create request failed: %w", err)
	}

	// 如果本地文件存在，添加 Range 头实现断点续传
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	// 发起请求
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	// 200: 完整下载, 206: 部分内容（断点续传）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return resp.StatusCode, resp.Header, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 打开或创建目标文件
	var file *os.File
	if offset > 0 && resp.StatusCode == http.StatusPartialContent {
		// 断点续传，追加模式打开
		file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_APPEND, 0o644)
	} else {
		// 完整下载，创建新文件
		file, err = os.Create(destPath)
		offset = 0
	}

	if err != nil {
		return resp.StatusCode, resp.Header, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, fmt.Errorf("write file failed: %w", err)
	}

	// 设置文件时间戳为远程服务器的时间
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if t, err := time.Parse(time.RFC1123, lastModified); err == nil {
			os.Chtimes(destPath, t, t)
		}
	}

	return resp.StatusCode, resp.Header, nil
}
