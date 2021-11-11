package main

import (
	"context"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	sdk "gitlab.com/jdb.com.la/sdk-go/billing/tapwater"
)

// ErrNoInfo is returned when no info is found.
var ErrNoInfo = errors.New("no info")

// ErrNotAllowCcy is return when currency is foreign for fix transaction amount
// Reference: https://litd-team.monday.com/boards/1305312982/pulses/1821535994?asset_id=316617661
var ErrNotAllowCcy = errors.New("not allow foreign currency for this bill type")

// ErrPermissionDenied is returned when the user
// does not have permission to perform the action.
var ErrPermissionDenied = errors.New("permission denied")

type AquaService struct {
	db       *sqlx.DB
	aqClient *sdk.Client
}

type Province = sdk.Province
type InfoQuery = sdk.InfoQuery
type Info = sdk.Info

func (aq *AquaService) Province(ctx context.Context) ([]Province, error) {
	return aq.aqClient.Provinces(ctx)
}

func (aq *AquaService) Info(ctx context.Context, in InfoQuery) (Info, error) {
	info, err := aq.aqClient.Info(ctx, in)
	if err != nil {
		var lowerErr = strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "invalid customer no") || strings.Contains(lowerErr, "invalid province no") {
			return Info{}, ErrNoInfo
		}
		if strings.Contains(lowerErr, "account currency not allow") {
			return Info{}, ErrNotAllowCcy
		}
		return Info{}, err
	}
	return info, nil
}
