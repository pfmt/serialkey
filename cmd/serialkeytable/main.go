package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pfmt/serialkey"
)

func main() {
	cmd := kong.Parse(&CLI)

	switch cmd.Command() {
	case "postgresql":

		ctx := context.Background()

		if strings.TrimSpace(CLI.Postgresql.URL) == "" {
			fmt.Fprintf(os.Stderr, "missing pgx URL\n")
			os.Exit(1)
		}

		pool, err := pgxpool.New(ctx, CLI.Postgresql.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pgx connect %s: %s\n", CLI.Postgresql.URL, err)
			os.Exit(1)
		}

		err = serialkey.NewPgxPool(pool).CreateTable(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create PostgreSQL table %s: %s\n", CLI.Postgresql.URL, err)
			os.Exit(1)
		}

	default:
		fmt.Fprint(os.Stderr, cmd.Command())
		os.Exit(1)
	}
}

var CLI struct {
	Postgresql struct {
		URL   string `env:"PGXURL" default:"postgres://postgres@localhost:5432/postgres" help:"Specify a PostgreSQL connection. ${env}=${default}"`
		Table string `env:"TABLE" default:"serialkeys" help:"Specify an alternate table name. ${env}=${default}"`
	} `cmd:"" help:"Create PostgreSQL table."`
}
