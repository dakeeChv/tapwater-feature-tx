package main

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	sdk "gitlab.com/jdb.com.la/sdk-go/billing/tapwater"
)

// ErrNoInfo is returned when no info is found.
var ErrNoInfo = errors.New("no info")

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
		if err.Error() != "" {
			return Info{}, ErrNoInfo
		}
		return Info{}, err
	}
	return info, nil
}
