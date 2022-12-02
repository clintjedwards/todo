package api

import (
	"context"

	proto "github.com/clintjedwards/todo/proto"
)

// GetSystemInfo returns system information and health
func (api *API) GetSystemInfo(context context.Context, request *proto.GetSystemInfoRequest) (*proto.GetSystemInfoResponse, error) {
	version, commit := parseVersion(appVersion)

	return &proto.GetSystemInfoResponse{
		Commit:         commit,
		DevModeEnabled: api.config.DevMode,
		Semver:         version,
	}, nil
}
