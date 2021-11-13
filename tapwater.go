package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
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
type Info = sdk.Info
type Tx = sdk.Tx

type InfoRequest struct {
	Account  sdk.Account  `json:"account"`
	Customer sdk.Customer `json:"customer"`
}

func (i *InfoRequest) Validate() error {
	AccountField := i.Account.BAN == "" || i.Account.Ccy == ""
	CustomerField := i.Customer.ID == "" || i.Customer.ProvinceID == ""
	if AccountField || CustomerField {
		fmt.Println("request not valid")
		return fmt.Errorf("request not valid")
	}
	return nil
}

type TxRequest struct {
	ID         string          `json:"id"`
	ExternalID string          `json:"externalId"`
	Account    sdk.Account     `json:"account"`
	Customer   sdk.Customer    `json:"customer"`
	Amount     decimal.Decimal `json:"amount"`
	LCYAmount  decimal.Decimal `json:"lcyAmount"`
	Fee        decimal.Decimal `json:"fee"`
	LCYFee     decimal.Decimal `json:"lcyFee"`
	Memo       string          `json:"memo"`
	PhotoURL   string          `json:"photoUrl"`
	Time       time.Time       `json:"time"`
}

func (t *TxRequest) Validate() error {
	AccountField := t.Account.DisplayName == "" || t.Account.BAN == "" || t.Account.Ccy == ""
	CustomerField := t.Customer.ID == "" || t.Customer.ProvinceID == ""
	TxInfoField := t.ExternalID == "" || t.Amount.IsZero() || t.Amount.IsNegative()  || t.LCYAmount.IsZero() || t.LCYAmount.IsNegative() || t.Fee.IsZero() || t.Fee.IsNegative()
	if AccountField || CustomerField || TxInfoField {
		return fmt.Errorf("request not valid")
	}
	return nil
}


func (aq *AquaService) Province(ctx context.Context) ([]Province, error) {
	return aq.aqClient.Provinces(ctx)
}

func (aq *AquaService) Info(ctx context.Context, in InfoRequest) (Info, error) {
	type InfoQuery = sdk.InfoQuery
	In := InfoQuery{
		Account:  in.Account,
		Customer: in.Customer,
	}
	info, err := aq.aqClient.Info(ctx, In)
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

func (aq *AquaService) Tx(ctx context.Context, in Tx) (Tx, error) {
	return Tx{}, nil
}
