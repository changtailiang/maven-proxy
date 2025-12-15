// pkg/config/config.go
package config

import (
	"github.com/sirupsen/logrus"
)

// Config 主配置结构体
type Config struct {
	Listen          string        `yaml:"listen" default:"localhost"`
	Port            string        `yaml:"port" default:"8880"`
	Context         string        `yaml:"context" default:"maven"`
	LocalRepository string        `yaml:"localRepository" default:"."`
	User            []*User       `yaml:"user"`
	Repository      []*Repository `yaml:"repository"`
	Logging         *Logging      `yaml:"logging"`
}

// User 用户认证信息
type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

// Repository 仓库配置
type Repository struct {
	Id      string            `yaml:"id"`
	Name    string            `yaml:"name"`
	Target  string            `yaml:"target"`
	Mode    int               `yaml:"mode" default:"4"`
	Cache   bool              `yaml:"cache" default:"false"`
	Mirror  []string          `yaml:"mirror"`
	Type    string            `yaml:"type" default:"hosted"`
	Members []string          `yaml:"members"`
	Routes  map[string]string `yaml:"routes"`
}

// Logging 日志配置
type Logging struct {
	Path  string       `yaml:"path" default:""`
	Level logrus.Level `yaml:"level" default:"debug"`
}
