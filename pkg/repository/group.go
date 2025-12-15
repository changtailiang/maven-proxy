// pkg/repository/group.go
package repository

import (
	"errors"
	"net/http"
	"sort"
	"strings"

	"maven-proxy/pkg/storage"
)

type GroupRepository struct {
	id      string
	mode    int
	members []Repository
	routes  map[string]string
}

func NewGroupRepository(id string, mode int, members []Repository, routes map[string]string) *GroupRepository {
	return &GroupRepository{
		id:      id,
		mode:    mode,
		members: members,
		routes:  routes,
	}
}

func (r *GroupRepository) ID() string {
	return r.id
}

func (r *GroupRepository) Type() string {
	return "group"
}

func (r *GroupRepository) CanRead() bool {
	return r.mode&4 == 4
}

func (r *GroupRepository) CanWrite() bool {
	return r.mode&2 == 2
}

func (r *GroupRepository) Get(path string) ([]byte, int, http.Header, error) {
	// 按优先级遍历成员仓库
	for _, member := range r.members {
		if !member.CanRead() {
			continue
		}

		if data, status, headers, err := member.Get(path); err == nil {
			return data, status, headers, nil
		}
	}

	return nil, http.StatusNotFound, nil, errors.New("artifact not found in any member repository")
}

func (r *GroupRepository) Put(path string, data []byte) error {
	// 根据路由规则选择目标仓库
	targetRepo := r.routeToTarget(path)
	if targetRepo == nil {
		return errors.New("no target repository for path")
	}

	return targetRepo.Put(path, data)
}

func (r *GroupRepository) List(path string) ([]storage.FileInfo, error) {
	// 聚合所有成员仓库的目录列表
	fileMap := make(map[string]storage.FileInfo)

	for _, member := range r.members {
		if !member.CanRead() {
			continue
		}

		entries, err := member.List(path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if _, exists := fileMap[entry.Name]; !exists {
				fileMap[entry.Name] = entry
			}
		}
	}

	result := make([]storage.FileInfo, 0, len(fileMap))
	for _, info := range fileMap {
		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func (r *GroupRepository) routeToTarget(path string) Repository {
	isSnapshot := strings.Contains(strings.ToLower(path), "-snapshot")

	var targetId string
	if isSnapshot {
		targetId = r.routes["snapshot"]
	} else {
		targetId = r.routes["release"]
	}

	for _, member := range r.members {
		if member.ID() == targetId {
			return member
		}
	}

	return nil
}
