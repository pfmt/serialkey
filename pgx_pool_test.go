// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pfmt/serialkey"
)

var (
	pgxPool *pgxpool.Pool
	pgxErr  error
	pgxOpt  = serialkey.PgxPoolWithStart(1)
)

func TestPgx(t *testing.T) {
	if pgxErr != nil {
		t.Log(pgxErr)
		return
	}

	chain := serialkey.NewPgxPool(pgxPool, pgxOpt)
	serailKeyTest(t, chain)
	closer.add(chain.Close)
}

func BenchmarkPgxNext(b *testing.B) {
	if pgxErr != nil {
		b.Log(pgxErr)
		return
	}

	chain := serialkey.NewPgxPool(pgxPool, pgxOpt)
	nextSerailKeyBenchmark(b, chain)
	closer.add(chain.Close)
}

func NewPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
	url, ok := os.LookupEnv("PGXURL")
	if !ok {
		return nil, errors.New("missing pgx URL")
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("pgx connect %s: %w", url, err)
	}

	return pool, nil
}
