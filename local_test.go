// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey_test

import (
	"testing"

	"github.com/pfmt/serialkey"
)

var localOpt = serialkey.LocalWithStart(1)

func TestLocal(t *testing.T) {
	chain := serialkey.NewLocal(localOpt)
	serailKeyTest(t, chain)
	closer.add(chain.Close)
}

func BenchmarkLocalNext(b *testing.B) {
	chain := serialkey.NewLocal(localOpt)
	nextSerailKeyBenchmark(b, chain)
	closer.add(chain.Close)
}
