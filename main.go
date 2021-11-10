package main

import (
	"fmt"
	"os"
	"time"

	"github.com/labstack/echo/v4"
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


func main() {}

func newEchoServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	return e
}