// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxPool returns the serialkeys keychain based on the pgx pool.
func NewPgxPool(pool *pgxpool.Pool, opts ...PgxPoolOption) *PgxPool {
	cfg := PgxPoolConfiguration{table: Table}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &PgxPool{
		start: cfg.start,
		table: cfg.table,
		pool:  pool,
	}
}

// PgxPool is the serialkeys keychain based on the pgx pool.
type PgxPool struct {
	sync.RWMutex
	start        int64
	table        string
	pool         *pgxpool.Pool
	nextQuery    string
	nextNQuery   string
	lastQuery    string
	forwardQuery string
	created      bool
	closed       bool
}

// Next for the passed key name returns an value guaranteed to be greater
// than the value returned for the same key name passed at the time
// of previous call of the next method or the forward method.
// The next method is thread safe.
func (chain *PgxPool) Next(ctx context.Context, key string) (int64, error) {
	chain.RLock()

	if chain.nextQuery != "" {
		value, err := chain.next(ctx, key)
		chain.RUnlock()
		return value, err
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if chain.nextQuery == "" {
		q, err := PostgreSQL{Table: chain.table}.next()
		if err != nil {
			return 0, fmt.Errorf("generate the next value fetching query: %w", err)
		}
		chain.nextQuery = q
	}

	return chain.next(ctx, key)
}

func (chain *PgxPool) next(ctx context.Context, key string) (int64, error) {
	conn, err := chain.conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var value int64

	err = conn.QueryRow(ctx, chain.nextQuery, key, chain.start).Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("fetch next value %s: %w", key, err)
	}

	return value, nil
}

// NextN for the passed key name returns an value guaranteed
// to be greater than the returned value for the same key name
// passed at the time of the previous method call.
// The next method is thread safe.
//
// TODO: Replace NextN method by CopyNext method.
func (chain *PgxPool) NextN(ctx context.Context, key string, count int64) (int64, error) {
	chain.RLock()

	if chain.nextNQuery != "" {
		value, err := chain.nextN(ctx, key, count)
		chain.RUnlock()
		return value, err
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if chain.nextNQuery == "" {
		q, err := PostgreSQL{Table: chain.table}.nextN()
		if err != nil {
			return 0, fmt.Errorf("generate the next N values fetching query: %w", err)
		}
		chain.nextNQuery = q
	}

	return chain.nextN(ctx, key, count)
}

func (chain *PgxPool) nextN(ctx context.Context, key string, count int64) (int64, error) {
	conn, err := chain.conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var value int64

	err = conn.QueryRow(ctx, chain.nextNQuery, key, count).Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("fetch next values %s: %w", key, err)
	}

	return value, nil
}

// Last for the passed key name returns the value returned for
// the same key name passed at the time of previous call
// of the next method or the forward method.
// Last method must be thread safe.
func (chain *PgxPool) Last(ctx context.Context, key string) (int64, error) {
	chain.RLock()

	if chain.lastQuery != "" {
		value, err := chain.last(ctx, key)
		chain.RUnlock()
		return value, err
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if chain.lastQuery == "" {
		q, err := PostgreSQL{Table: chain.table}.last()
		if err != nil {
			return 0, fmt.Errorf("generate the last value fetching query: %w", err)
		}
		chain.lastQuery = q
	}

	return chain.last(ctx, key)
}

func (chain *PgxPool) last(ctx context.Context, key string) (int64, error) {
	conn, err := chain.conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var value int64

	err = conn.QueryRow(ctx, chain.lastQuery, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return chain.start - 1, nil

	} else if err != nil {
		return 0, fmt.Errorf("fetch last value %s: %w", key, err)
	}

	return value, nil
}

// Forward for the passed key name returns an value guaranteed
// to be greater or equal to the target value and guaranteed to be greater
// than the value returned for the same key name passed at the time
// of previous call of the forward method or the next method.
// Forward method is thread safe.
func (chain *PgxPool) Forward(ctx context.Context, key string, target int64) (int64, error) {
	chain.RLock()

	if chain.forwardQuery != "" {
		value, err := chain.forward(ctx, key, target)
		chain.RUnlock()
		return value, err
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if chain.forwardQuery == "" {
		q, err := PostgreSQL{Table: chain.table}.forward()
		if err != nil {
			return 0, fmt.Errorf("generate the forward value query: %w", err)
		}
		chain.forwardQuery = q
	}

	return chain.forward(ctx, key, target)
}

func (chain *PgxPool) forward(ctx context.Context, key string, target int64) (int64, error) {
	conn, err := chain.conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var value int64

	err = conn.QueryRow(ctx, chain.forwardQuery, key, target).Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("forward value %s to %d: %w", key, target, err)
	}

	return value, nil
}

// CreateTable creates the PostgreSQL table if not exists.
// The create table method is thread safe.
func (chain *PgxPool) CreateTable(ctx context.Context) error {
	chain.RLock()

	if chain.created {
		chain.RUnlock()
		return nil
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if !chain.created {
		q, err := PostgreSQL{Table: chain.table}.createTable()
		if err != nil {
			return fmt.Errorf("generate the table creation query: %w", err)
		}

		conn, err := chain.conn(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		_, err = conn.Exec(ctx, q)
		if err != nil {
			return fmt.Errorf("execute the table creation query: %w", err)
		}

		chain.created = true
	}

	return nil
}

func (chain *PgxPool) conn(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := chain.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	return conn, nil
}

// Close closes pgx pool.
// The close method is thread safe.
func (chain *PgxPool) Close() error {
	chain.RLock()

	if chain.closed {
		chain.RUnlock()
		return nil
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if chain.closed {
		return nil
	}

	chain.pool.Close()
	chain.closed = true

	return nil
}

// PgxPoolOption changes configuration.
type PgxPoolOption func(*PgxPoolConfiguration)

// PgxPoolConfiguration holds values changeable by options.
type PgxPoolConfiguration struct {
	start int64
	table string
}

// PgxPoolWithStart sets the start number.
func PgxPoolWithStart(start int64) PgxPoolOption {
	return func(cfg *PgxPoolConfiguration) { cfg.start = start }
}

// PgxPoolWithTable sets the table name.
func PgxPoolWithTable(table string) PgxPoolOption {
	return func(cfg *PgxPoolConfiguration) { cfg.table = table }
}
