// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey_test

import (
	"fmt"
	"sync"

	"go.uber.org/multierr"
)

type Close struct {
	sync.Mutex
	once  sync.Once
	funcs []func() error
}

func (c *Close) add(f ...func() error) {
	c.Lock()
	c.funcs = append(c.funcs, f...)
	c.Unlock()
}

func (c *Close) close() {
	c.once.Do(func() {
		c.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.Unlock()

		var err error

		for _, f := range funcs {
			if e := f(); e != nil {
				err = multierr.Append(err, e)
			}
		}

		if err != nil {
			panic(fmt.Errorf("closer: %w", err))
		}
	})
}
