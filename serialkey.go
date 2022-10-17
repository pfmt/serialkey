// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey

import "context"

const Table = "serialkeys"

// Chain is the persistence interface for the serialkey sequences.
type Chain interface {
	// Next for the passed key name returns an value guaranteed to be greater
	// than the value returned for the same key name passed at the time
	// of previous call of the next method or the forward method.
	// Next method must be thread safe.
	Next(ctx context.Context, key string) (value int64, err error)

	// Last for the passed key name returns the value returned for
	// the same key name passed at the time of previous call
	// of the next method or the forward method.
	// Last method must be thread safe.
	Last(ctx context.Context, key string) (value int64, err error)

	// Forward for the passed key name returns an value guaranteed
	// to be greater or equal to the target value and guaranteed to be greater
	// than the value returned for the same key name passed at the time
	// of previous call of the forward method or the next method.
	// Forward method must be thread safe.
	Forward(ctx context.Context, key string, target int64) (result int64, err error)

	// Close closes key chain.
	//
	// Close method must be thread safe.
	// Specific implementations may document their own behavior.
	Close() (err error)
}
