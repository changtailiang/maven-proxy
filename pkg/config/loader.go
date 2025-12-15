// pkg/config/loader.go
package config

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/creasty/defaults"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var log = logrus.New()

// Loader 配置加载器
type Loader struct {
	configPath string
}

// NewLoader 创建配置加载器
func NewLoader() *Loader {
	loader := &Loader{}
	flag.StringVar(&loader.configPath, "c", "config.yaml", "配置文件路径")
	flag.Parse()
	return loader
}

// Load 加载配置文件
func (l *Loader) Load() (*Config, error) {
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	log.Infof("configure file: %s", l.configPath)

	// 读取配置文件
	data, err := ioutil.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	// 解析 YAML
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 设置默认值
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("set defaults failed: %w", err)
	}

	// 验证和预处理仓库配置
	if err := l.validateRepositories(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateRepositories 验证和预处理仓库配置
func (l *Loader) validateRepositories(cfg *Config) error {
	repoStore := make(map[string]*Repository)

	// 第一遍：基本验证和索引
	for _, repo := range cfg.Repository {
		// 跳过未启用的仓库
		if repo.Mode == 0 {
			log.Warnf("repository '%s' is disabled (mode=0), skipping", repo.Id)
			continue
		}

		// 设置默认 target
		if repo.Target == "" {
			repo.Target = repo.Id
		}

		// 验证 group 类型仓库
		if repo.Type == "group" {
			if len(repo.Members) == 0 {
				log.Warnf("group repository '%s' has no members, skipping", repo.Id)
				continue
			}
		}

		repoStore[repo.Id] = repo
		log.Infof("repository: http://%s:%s/%s/%s local dirname: %s",
			cfg.Listen, cfg.Port, cfg.Context, repo.Id, repo.Target)
	}

	// 第二遍：验证 group 仓库的成员引用
	for _, repo := range cfg.Repository {
		if repo.Type == "group" && len(repo.Members) > 0 {
			validMembers := []string{}
			for _, memberId := range repo.Members {
				if _, exists := repoStore[memberId]; exists {
					validMembers = append(validMembers, memberId)
				} else {
					log.Warnf("group repository '%s' references non-existent member '%s'",
						repo.Id, memberId)
				}
			}
			repo.Members = validMembers

			if len(validMembers) == 0 {
				log.Warnf("group repository '%s' has no valid members after validation", repo.Id)
			}
		}
	}

	return nil
}
