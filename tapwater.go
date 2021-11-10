package main

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	sdk "gitlab.com/jdb.com.la/sdk-go/billing/tapwater"
)

// ErrNoCustomer is returned when the customer info is not found
var ErrNoCustomer = errors.New("no customer info")

// ErrPermissionDenied is returned when the user
// does not have permission to perform the action.
var ErrPermissionDenied = errors.New("permission denied")

type AquaService struct {
	db       *sqlx.DB
	aqClient *sdk.Client
}

type Province = sdk.Province

func (aq *AquaService) GetProvince(ctx context.Context) ([]Province, error) {
	return aq.aqClient.Provinces(ctx)
}
