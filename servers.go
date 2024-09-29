package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/types"
)

type Server struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type ServersInfo struct {
	DefaultLogin string   `json:"default_login"`
	Logins       []string `json:"logins"`
	Servers      []Server `json:"servers"`
}

func FetchServersInfo(cr client.Credentials) (*ServersInfo, error) {
	ctx := context.Background()

	clt, err := client.New(ctx, client.Config{
		Credentials: []client.Credentials{
			cr,
		},
	})
	if err != nil {
		return nil, err
	}
	defer clt.Close()

	roles, err := clt.GetCurrentUserRoles(context.Background())
	if err != nil {
		return nil, err
	}

	logins := make([]string, 0)
	for _, role := range roles {
		for _, login := range role.GetLogins(types.Allow) {
			if !slices.Contains(logins, login) {
				logins = append(logins, login)
			}
		}
	}

	servers := make([]Server, 0)

	req := proto.ListResourcesRequest{
		ResourceType: types.KindNode,
		Limit:        500,
	}

	for {
		res, err := clt.ListResources(context.Background(), req)
		if err != nil {
			return nil, err
		}

		for _, node := range res.Resources {
			name := types.FriendlyName(node)
			labels := node.GetStaticLabels()

			servers = append(servers, Server{Name: name, Labels: labels})
		}

		if res.NextKey == "" {
			break
		}

		req.StartKey = res.NextKey
	}

	return &ServersInfo{
		Logins:  logins,
		Servers: servers,
	}, nil
}

func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	folder := filepath.Join(cacheDir, "tssh")
	err = os.MkdirAll(folder, os.ModeDir|0755)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(folder, "servers.json")

	return filePath, nil
}

func GetServersInfoFromCache() (*ServersInfo, error) {
	filepath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	serversInfo := ServersInfo{}
	err = json.Unmarshal(file, &serversInfo)
	if err != nil {
		return nil, err
	}

	return &serversInfo, nil
}

func StoreServersInfo(info *ServersInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	filepath, err := getCachePath()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}
