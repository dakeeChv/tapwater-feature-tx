package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/random"
	"github.com/shopspring/decimal"
	sdk "gitlab.com/jdb.com.la/sdk-go/billing/tapwater"
)

// ErrNoInfo is returned when no info is found.
var ErrNoInfo = errors.New("no info")

// ErrNotAllowCcy is return when currency is foreign for fix transaction amount.
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
		fmt.Println("validator")
		return fmt.Errorf("request not valid")
	}
	return nil
}

type TxRequest struct {
	ExternalID string          `json:"externalId"`
	Account    sdk.Account     `json:"account"`
	Customer   sdk.Customer    `json:"customer"`
	Amount     decimal.Decimal `json:"amount"`
	LCYAmount  decimal.Decimal `json:"lcyAmount"`
	Fee        decimal.Decimal `json:"fee"`
	Memo       string          `json:"memo"`
	PhotoURL   string          `json:"photoUrl"`
}

func (t *TxRequest) Validate() error {
	AccountField := t.Account.DisplayName == "" || t.Account.BAN == "" || t.Account.Ccy == ""
	CustomerField := t.Customer.ID == "" || t.Customer.ProvinceID == ""
	TxInfoField := t.ExternalID == "" || t.Amount.IsZero() || t.Amount.IsNegative() || t.LCYAmount.IsZero() || t.LCYAmount.IsNegative() || t.Fee.IsNegative()
	fmt.Println("Validator: ", TxInfoField)
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

func (aq *AquaService) DoTx(ctx context.Context, in TxRequest) (Tx, error) {
	TxQuery := Tx{
		ExternalID: in.ExternalID,
		Account:    in.Account,
		Customer:   in.Customer,
		Amount:     in.Amount,
		LCYAmount:  in.LCYAmount,
		Fee:        in.Fee,
		Memo:       in.Memo,
		PhotoURL:   in.PhotoURL,
	}
	TxRespone, err := aq.aqClient.DoTx(ctx, TxQuery)

	tx := aq.db.MustBeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

	doneCh := make(chan error, 1)
	errCh := make(chan error, 1)
	go func() {
		doneCh <- saveTx(ctx, tx, TxQuery)
	}()
	
	for {
		select {
		case <-doneCh:
			if err != nil {
				errCh <- err
				break
			}
			updateTx := `UPDATE transaction SET reference_number = $1, lcy_fee = $2, success = $3 WHERE external_number = $4`
			updateCustomer := `UPDATE customer SET display_name = $1 WHERE id = $2`
			tx.MustExecContext(ctx, updateTx, TxRespone.ID, TxRespone.LCYFee, "true", TxRespone.ExternalID)
			tx.MustExecContext(ctx, updateCustomer, TxRespone.Customer.DisplayName, TxRespone.Customer.ID)
			tx.Commit()
			return TxRespone, nil
		case err := <-errCh:
			tx.Commit()
			return Tx{}, err
		}
	}
}

func saveTx(ctx context.Context, tx *sqlx.Tx, in Tx) error {
	queryAccount :=
		`INSERT INTO account (
			cif, ban, display_name,
			type, currency) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (ban) DO NOTHING`

	queryCustomer :=
		`INSERT INTO customer (
			id, province_id)
		VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING`

	queryTx :=
		`INSERT INTO transaction (
			id, external_number, customer, account,
			monthly, amount, lcy_amount, fee,
			memo, photo_url, user_id, device_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	tx.MustExecContext(ctx, queryAccount, in.Account.CIF, in.Account.BAN, in.Account.DisplayName, in.Account.Type, in.Account.Ccy)
	tx.MustExecContext(ctx, queryCustomer, in.Customer.ID, in.Customer.ProvinceID)
	tx.MustExecContext(ctx, queryTx, random.New().String(32), in.ExternalID, in.Customer.ID, in.Account.BAN, in.Customer.Time, in.Amount, in.LCYAmount, in.Fee, in.Memo, in.PhotoURL, "daky", "101")
	return nil
}
