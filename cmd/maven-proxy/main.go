// cmd/maven-proxy/main.go
package main

import (
	"log"

	"maven-proxy/internal/server"
	"maven-proxy/pkg/auth"
	"maven-proxy/pkg/config"
	"maven-proxy/pkg/repository"
	"maven-proxy/pkg/storage"
)

func main() {
	// 加载配置
	loader := config.NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 初始化存储层
	baseStorage := storage.NewFileSystemStorage(cfg.LocalRepository)

	// 创建认证器
	authenticator := auth.NewBasicAuthenticator(cfg.User)

	// 创建服务器
	srv := server.NewServer(cfg, authenticator)

	// 初始化仓库
	repoStore := make(map[string]repository.Repository)

	// 第一遍：创建所有 hosted 和 proxy 仓库
	for _, repoCfg := range cfg.Repository {
		if repoCfg.Mode == 0 {
			continue
		}

		switch repoCfg.Type {
		case "hosted", "":
			// 创建 hosted 仓库
			repoStorage := storage.NewPrefixedStorage(baseStorage, repoCfg.Target)
			repo := repository.NewHostedRepository(
				repoCfg.Id,
				repoCfg.Mode,
				repoStorage,
			)
			repoStore[repoCfg.Id] = repo
			log.Printf("initialized hosted repository: %s", repoCfg.Id)

		case "proxy":
			// 创建 proxy 仓库
			repoStorage := storage.NewPrefixedStorage(baseStorage, repoCfg.Target)
			repo := repository.NewProxyRepository(
				repoCfg.Id,
				repoCfg.Mode,
				repoCfg.Cache,
				repoCfg.Mirror,
				repoStorage,
			)
			repoStore[repoCfg.Id] = repo
			log.Printf("initialized proxy repository: %s", repoCfg.Id)
		}
	}

	// 第二遍：创建 group 仓库（依赖前面创建的仓库）
	for _, repoCfg := range cfg.Repository {
		if repoCfg.Mode == 0 || repoCfg.Type != "group" {
			continue
		}

		// 收集成员仓库
		members := []repository.Repository{}
		for _, memberId := range repoCfg.Members {
			if memberRepo, exists := repoStore[memberId]; exists {
				members = append(members, memberRepo)
			} else {
				log.Printf("warning: group repository '%s' references non-existent member '%s'",
					repoCfg.Id, memberId)
			}
		}

		if len(members) == 0 {
			log.Printf("warning: group repository '%s' has no valid members, skipping", repoCfg.Id)
			continue
		}

		// 创建 group 仓库
		repo := repository.NewGroupRepository(
			repoCfg.Id,
			repoCfg.Mode,
			members,
			repoCfg.Routes,
		)
		repoStore[repoCfg.Id] = repo
		log.Printf("initialized group repository: %s with %d members", repoCfg.Id, len(members))
	}

	// 注册所有仓库到服务器
	for id, repo := range repoStore {
		srv.RegisterRepository(id, repo)
	}

	// 启动服务器
	addr := cfg.Listen + ":" + cfg.Port
	log.Printf("maven-proxy server starting on %s", addr)
	log.Printf("Context path: /%s", cfg.Context)

	if err := srv.Run(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
