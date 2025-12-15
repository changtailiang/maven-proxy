// internal/server/handlers.go
package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"maven-proxy/internal/util"
	"maven-proxy/pkg/repository"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleGet(c *gin.Context) {
	repoId := c.Param("repoId")
	filePath := c.Param("path")

	repo, exists := s.repositories[repoId]
	if !exists {
		c.String(http.StatusNotFound, "repository not found")
		return
	}

	if !repo.CanRead() {
		c.String(http.StatusForbidden, "repository not support read")
		return
	}

	// 处理目录浏览
	ext := path.Ext(filePath)
	if ext == "" {
		if !strings.HasSuffix(filePath, "/") {
			c.Redirect(http.StatusMovedPermanently, c.Request.RequestURI+"/")
			return
		}

		// 渲染目录列表
		if html, err := s.renderDirectoryListing(repo, filePath); err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
			return
		}
	}

	// 获取文件内容
	data, status, headers, err := repo.Get(filePath)
	if err != nil {
		c.String(status, err.Error())
		return
	}

	// 处理哈希生成
	if generate := c.Query("generate_md5_sha1"); strings.EqualFold(generate, "true") {
		// 对于 hosted 仓库，生成哈希文件
		if repo.Type() == "hosted" {
			localPath := path.Join(s.config.LocalRepository, repo.ID(), filePath)
			if err := util.GenerateHash(localPath); err != nil {
				c.String(http.StatusInternalServerError, "generate hash failed")
				return
			}
		}
	}

	c.Data(status, headers.Get("Content-Type"), data)
}

func (s *Server) handlePut(c *gin.Context) {
	// 认证已经在中间件中完成

	repoId := c.Param("repoId")
	filePath := c.Param("path")

	repo, exists := s.repositories[repoId]
	if !exists {
		c.String(http.StatusNotFound, "repository not found")
		return
	}

	if !repo.CanWrite() {
		c.String(http.StatusForbidden, "repository not support write")
		return
	}

	// 读取请求体
	length, err1 := strconv.Atoi(c.GetHeader("Content-Length"))
	data, err2 := ioutil.ReadAll(c.Request.Body)
	if err1 != nil || err2 != nil || length <= 0 || length != len(data) {
		c.String(http.StatusInternalServerError, "data read failed")
		return
	}

	// 上传文件
	if err := repo.Put(filePath, data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// 处理哈希生成
	if generate := c.Query("generate_md5_sha1"); strings.EqualFold(generate, "true") {
		if repo.Type() == "hosted" {
			localPath := path.Join(s.config.LocalRepository, repo.ID(), filePath)
			if err := util.GenerateHash(localPath); err != nil {
				c.String(http.StatusInternalServerError, "generate hash failed")
				return
			}
		}
	}

	c.String(http.StatusOK, "OK")
}

func (s *Server) renderDirectoryListing(repo repository.Repository, filePath string) (string, error) {
	entries, err := repo.List(filePath)
	if err != nil {
		return "", err
	}

	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html>\n<head>\n")
	html.WriteString("<meta charset=\"utf-8\">\n")
	html.WriteString(fmt.Sprintf("<title>Index of %s - %s</title>\n", filePath, repo.ID()))
	html.WriteString("<style>\n")
	html.WriteString("body { font-family: monospace; margin: 20px; }\n")
	html.WriteString("h1 { border-bottom: 1px solid #ccc; padding-bottom: 10px; }\n")
	html.WriteString("table { border-collapse: collapse; width: 100%; }\n")
	html.WriteString("th { text-align: left; padding: 8px; background-color: #f0f0f0; }\n")
	html.WriteString("td { padding: 8px; border-bottom: 1px solid #eee; }\n")
	html.WriteString("a { text-decoration: none; color: #0066cc; }\n")
	html.WriteString("a:hover { text-decoration: underline; }\n")
	html.WriteString(".dir { font-weight: bold; }\n")
	html.WriteString(".size { text-align: right; }\n")
	html.WriteString(".date { color: #666; }\n")
	html.WriteString("</style>\n")
	html.WriteString("</head>\n<body>\n")

	html.WriteString(fmt.Sprintf("<h1>Index of %s - %s</h1>\n", filePath, repo.ID()))
	html.WriteString("<table>\n")
	html.WriteString("<tr><th>Name</th><th>Size</th><th>Last Modified</th></tr>\n")

	// 添加父目录链接
	if filePath != "/" {
		parentPath := path.Dir(filePath)
		if !strings.HasSuffix(parentPath, "/") {
			parentPath += "/"
		}
		fullParentPath := fmt.Sprintf("/%s/%s%s", s.config.Context, repo.ID(), parentPath)
		html.WriteString(fmt.Sprintf("<tr><td class=\"dir\"><a href=\"%s\">../</a></td><td>-</td><td>-</td></tr>\n", path.Dir(filepath.Clean(fullParentPath))))
	}

	// 列出目录内容
	for _, entry := range entries {
		name := entry.Name
		size := "-"
		if !entry.IsDir {
			size = s.formatSize(entry.Size)
		} else {
			name += "/"
		}

		modTime := entry.ModTime.Format("2006-01-02 15:04:05")
		linkPath := path.Join(filePath, entry.Name)
		if entry.IsDir {
			linkPath += "/"
		}
		fullLinkPath := fmt.Sprintf("/%s/%s%s", s.config.Context, repo.ID(), linkPath)

		class := ""
		if entry.IsDir {
			class = " class=\"dir\""
		}

		html.WriteString(fmt.Sprintf("<tr><td%s><a href=\"%s\">%s</a></td><td class=\"size\">%s</td><td class=\"date\">%s</td></tr>\n",
			class, fullLinkPath, name, size, modTime))
	}

	html.WriteString("</table>\n")
	html.WriteString("</body>\n</html>")

	return html.String(), nil
}

func (s *Server) formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
