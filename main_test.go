// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/pfmt/serialkey"
)

const timeout = 3 * time.Second

var closer = &Close{}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pgxPool, pgxErr = NewPgxPool(ctx)
	defer closer.close()

	if pgxPool != nil {
		conn, err := pgxPool.Acquire(ctx)
		if err != nil {
			panic(fmt.Errorf("pgx acquire connection: %w", err))
		}
		defer conn.Release()

		_, err = conn.Exec(ctx, "TRUNCATE "+serialkey.Table)
		if err != nil {
			panic(fmt.Errorf("truncate: %w", err))
		}
	}

	os.Exit(m.Run())

	if pgxPool != nil {
		pgxPool.Close()
	}
}

type SerailKeyTest struct {
	test  string
	line  string
	name  string
	count int
	want  int64
	bench bool
	skip  bool
	keep  bool
}

var serailKeyTests = []SerailKeyTest{
	{
		test:  "next one foo",
		line:  testline(),
		name:  "foo",
		count: 1,
		want:  1,
		bench: true,
	}, {
		test:  "next 42 bar",
		line:  testline(),
		name:  "bar",
		count: 42,
		want:  42,
	}, {
		test:  "next one xyz",
		line:  testline(),
		name:  "xyz",
		count: 1,
		want:  1,
	}, {
		test:  "next 42 xyz",
		line:  testline(),
		name:  "xyz",
		count: 42,
		want:  42,
	}, {
		test:  "next 1234 xyz",
		line:  testline(),
		name:  "xyz",
		count: 1234,
		want:  1234,
	},
}

func serailKeyTest(t *testing.T, key serialkey.Chain) {
	t.Parallel()

	var keep, skip []SerailKeyTest
	for _, tt := range serailKeyTests {
		if tt.keep {
			keep = append(keep, tt)
		} else {
			skip = append(skip, tt)
		}
	}

	if len(keep) == 0 {
		keep = serailKeyTests
	} else {
		for _, tt := range skip {
			t.Logf("%s/unkeep: %s", tt.line, tt.test)
		}
	}

	for _, tt := range keep {
		if tt.skip {
			t.Logf("%s/skip: %s", tt.line, tt.test)
			continue
		}

		tt := tt

		t.Run(tt.line+"/"+tt.test, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			var current int64
			var err error

			for i := 0; i < tt.count; i++ {
				current, err = key.Next(ctx, tt.name)
				if err != nil {
					kv := errKV{err: err}
					kv.chain = key
					kv.line = tt.line
					t.Fatal(kv.Sprintf("next 64-bit integer, iteration %d", i))
				}
			}

			if current < tt.want {
				kv := cmpKV{want: tt.want, got: current}
				kv.chain = key
				kv.line = tt.line
				t.Error(kv.Sprint("current 64-bit integer"))
			}

			last, err := key.Last(ctx, tt.name)
			if err != nil {
				kv := errKV{err: err}
				kv.chain = key
				kv.line = tt.line
				t.Fatal(kv.Sprint("last 64-bit integer"))
			}

			if last < current {
				kv := cmpKV{want: current, got: last}
				kv.chain = key
				kv.line = tt.line
				t.Error(kv.Sprint("last 64-bit integer"))
			}
		})
	}
}

func nextSerailKeyBenchmark(b *testing.B, key serialkey.Chain) {
	b.ReportAllocs()

	var keep, skip []SerailKeyTest
	for _, tt := range serailKeyTests {
		if tt.keep {
			keep = append(keep, tt)
		} else {
			skip = append(skip, tt)
		}
	}

	if len(keep) == 0 {
		keep = serailKeyTests
	} else {
		for _, tt := range skip {
			b.Logf("%s/unkeep: %s", tt.line, tt.test)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, tt := range keep {
		if tt.skip {
			b.Logf("%s/skip: %s", tt.line, tt.test)
			continue
		}

		if !tt.bench {
			continue
		}

		b.Run(tt.line, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = key.Next(ctx, tt.name)
			}
		})
	}
}

type cmpKV struct {
	KV
	want int64
	got  int64
}

func (kv cmpKV) Sprint(a ...interface{}) string {
	kv.test = fmt.Sprint(a...)
	return kv.String()
}

func (kv cmpKV) Sprintf(format string, a ...interface{}) string {
	kv.test = format
	return fmt.Sprintf(kv.String(), a...)
}

func (kv cmpKV) String() string {
	s := kv.KV.String()
	s += "\nwant " + kv.test + " more or equal: " + strconv.FormatInt(kv.want, 10)
	s += "\ngot " + kv.test + ": " + strconv.FormatInt(kv.got, 10)
	return s
}

func testline() string {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	return "it was not possible to recover file and line number information about function invocations"
}

type errKV struct {
	KV
	err error
}

func (kv errKV) Sprint(a ...interface{}) string {
	return kv.Sprintf(fmt.Sprint(a...))
}

func (kv errKV) Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format+kv.String(), a...)
}

func (kv errKV) String() string {
	s := kv.KV.String()
	s += fmt.Sprintf("\nwant "+kv.test+" error: %s", kv.err)
	return s
}

type KV struct {
	test  string
	line  string
	chain serialkey.Chain
}

func (kv KV) String() string {
	s := "\nline: " + kv.line
	s += "\nkeychain: " + reflect.TypeOf(kv.chain).String()
	return s
}
