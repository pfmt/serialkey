# serialkey

[![Build Status](https://cloud.drone.io/api/badges/pfmt/serialkey/status.svg)](https://cloud.drone.io/pfmt/serialkey)
[![Go Reference](https://pkg.go.dev/badge/github.com/pfmt/serialkey.svg)](https://pkg.go.dev/github.com/pfmt/serialkey)

Dynamically named sequences for Go.  
Source files are distributed under the BSD-style license.

## About

The software is considered to be at a alpha level of readiness,
its extremely slow and allocates a lots of memory.

## PostgreSQL

Create table `make postgresql` or
`cat psql_create_table.sql | sed --expression='s/{{\.Table}}/serialkeys/g'`

```sql
CREATE TABLE IF NOT EXISTS serialkeys (
    key text primary key,
    value bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone
);
```

## Benchmark

```sh
$ go test -count=1 -race -bench ./...
goos: linux
goarch: amd64
pkg: github.com/pfmt/serialkey
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkLocalNext/main_test.go:66-8         	 6259942	       200.5 ns/op
BenchmarkPgxNext/main_test.go:66-8           	    2852	    362782 ns/op
PASS
ok  	github.com/pfmt/serialkey	3.031s
```
