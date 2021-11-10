package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"gitlab.com/jdb.com.la/sdk-go"
)

// GetEnv returns the environment variable or default value.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	PORT = fmt.Sprintf("%s", "3333")
	ADDR = fmt.Sprintf(":%s", PORT)
)

var (
	BANK_BASE_URL   = os.Getenv("BANK_BASE_URL")
	BANK_API_KEY    = os.Getenv("BANK_API_KEY")
	BANK_SECRET_KEY = os.Getenv("BANK_SECRET_KEY")
	BANK_HMACKEY    = os.Getenv("BANK_HMACKEY")
)

var (
	PGHOST     = GetEnv("PGHOST", "127.0.0.1")
	PGUSER     = os.Getenv("PGUSER")
	PGPASSWORD = os.Getenv("PGPASSWORD")
	PGDATABASE = os.Getenv("PGDATABASE")
	PGPORT     = GetEnv("PGPORT", "5432")

	// For the connection string syntax, see
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING.
	// Set the statement_timeout config parameter for this session.
	// See https://www.postgresql.org/docs/current/runtime-config-client.html.
	timeoutOption = fmt.Sprintf("-c statement_timeout=%d", 10*time.Minute/time.Millisecond)
	PGDSN         = fmt.Sprintf("user='%s' password='%s' host='%s' port=%s dbname='%s' sslmode=disable options='%s'",
		PGUSER, PGPASSWORD, PGHOST, PGPORT, PGDATABASE, timeoutOption)
)


func main() {
	if err := excute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func excute() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	sdk, err := sdk.New(ctx, &sdk.Config{
		BaseURL:   BANK_BASE_URL,
		APIKey:    BANK_API_KEY,
		SecretKey: BANK_SECRET_KEY,
		HMACKey:   []byte(BANK_HMACKEY),
	})
	if err != nil {
		return fmt.Errorf("create sdk client failure: %v", err)
	}
	defer sdk.Close()

	aqClient, err := sdk.TapWater(ctx)
	if err != nil {
		return fmt.Errorf("create jdb tapwater client failure: %v", err)
	}

	db, err := sqlx.Open("postgres", PGDSN)
	if err != nil && db.Ping() != nil {
		return fmt.Errorf("open database failure: %v", err)
	}

	aqService := &AquaService{
		db:       db,
		aqClient: aqClient,
	}

	h := &HTTPTransport{
		aqService: aqService,
	}

	e := newEchoServer()
	h.install(e)

	errCh := make(chan error, 1)
	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go func() {
		errCh <- e.Start(ADDR)
	}()

	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server failure: %v", err)
		}
		fmt.Println("server shutdown")
	case err := <-errCh:
		return fmt.Errorf("start server error: %v", err)
	}

	return nil
}

func newEchoServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	return e
}

