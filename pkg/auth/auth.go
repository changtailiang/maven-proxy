// pkg/auth/auth.go
package auth

import (
	"encoding/base64"
	"fmt"
	"strings"

	"maven-proxy/pkg/config"
)

// Authenticator 认证器接口
type Authenticator interface {
	// Authenticate 验证请求是否包含有效的认证信息
	Authenticate(authorization string) bool
}

// BasicAuthenticator Basic 认证实现
type BasicAuthenticator struct {
	authStore map[string]bool // 存储有效的 Base64 编码认证信息
}

// NewBasicAuthenticator 创建 Basic 认证器
func NewBasicAuthenticator(users []*config.User) *BasicAuthenticator {
	auth := &BasicAuthenticator{
		authStore: make(map[string]bool),
	}

	// 预处理用户认证信息
	for _, user := range users {
		base := fmt.Sprintf("%s:%s", user.Name, user.Password)
		encoded := base64.StdEncoding.EncodeToString([]byte(base))
		auth.authStore[encoded] = true
	}

	return auth
}

// Authenticate 验证 Authorization 头
func (a *BasicAuthenticator) Authenticate(authorization string) bool {
	// 检查是否为 Basic Auth
	if !strings.HasPrefix(authorization, "Basic ") {
		return false
	}

	// 提取 Base64 编码的凭证
	encoded := strings.TrimSpace(authorization[6:])

	// 验证凭证是否存在于认证存储中
	return a.authStore[encoded]
}
