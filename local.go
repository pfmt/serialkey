// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey

import (
	"context"
	"sync"
	"sync/atomic"
)

// NewLocal returns the serialkeys keychain based on the memory of the host
// where the module is running.
func NewLocal(opts ...LocalOption) *Local {
	var cfg LocalConfiguration

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Local{
		start: cfg.start,
		table: make(map[string]*int64),
	}
}

// Local is the serialkeys keychain based on the local memory.
type Local struct {
	sync.RWMutex
	start int64
	table map[string]*int64
}

// Next for the passed key name returns an value guaranteed to be greater
// than the value returned for the same key name passed at the time
// of previous call of the next method or the forward method.
// The next method is thread safe.
func (chain *Local) Next(_ context.Context, key string) (int64, error) {
	chain.RLock()

	if value, ok := chain.table[key]; ok {
		i := atomic.AddInt64((*int64)(value), 1)
		chain.RUnlock()
		return i, nil
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if value, ok := chain.table[key]; ok {
		return atomic.AddInt64((*int64)(value), 1), nil
	}

	i := chain.start
	chain.table[key] = &i

	return i, nil
}

// Last for the passed key name returns the value returned for
// the same key name passed at the time of previous call
// of the next method or the forward method.
// The last method is thread safe.
func (chain *Local) Last(_ context.Context, key string) (int64, error) {
	chain.RLock()
	defer chain.RUnlock()

	if value, ok := chain.table[key]; ok {
		return atomic.LoadInt64((*int64)(value)), nil
	}

	return chain.start - 1, nil
}

// Forward for the passed key name returns an value guaranteed
// to be greater or equal to the target value and guaranteed to be greater
// than the value returned for the same key name passed at the time
// of previous call of the forward method or the next method.
// Forward method is thread safe.
func (chain *Local) Forward(_ context.Context, key string, target int64) (int64, error) {
	chain.RLock()

	if value, ok := chain.table[key]; ok {
		if i := atomic.LoadInt64((*int64)(value)); i >= target {
			chain.RUnlock()
			return i, nil
		}
	}

	chain.RUnlock()
	chain.Lock()
	defer chain.Unlock()

	if value, ok := chain.table[key]; ok {
		if i := atomic.LoadInt64((*int64)(value)); i >= target {
			return i, nil
		}
	}

	chain.table[key] = &target

	return target, nil
}

// Close do nothing.
// The close method is thread safe.
func (*Local) Close() error {
	return nil
}

// LocalOption changes configuration.
type LocalOption func(*LocalConfiguration)

// LocalConfiguration holds values changeable by options.
type LocalConfiguration struct {
	start int64
}

// LocalWithStart sets the start number.
func LocalWithStart(start int64) LocalOption {
	return func(cfg *LocalConfiguration) { cfg.start = start }
}
