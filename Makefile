LOCAL_BIN:=$(CURDIR)/bin
PGXURL ?= postgres://postgres@localhost:5432/postgres

test: postgresql
	go test -count=1 -race ./...

bench: postgresql
	go test -count=1 -race -bench ./...

postgresql: deps
	PATH=$(LOCAL_BIN):$(PATH) PGXURL=$(PGXURL) serialkeytable postgresql

deps:
	GOBIN=$(LOCAL_BIN) go install github.com/pfmt/serialkey/cmd/serialkeytable@latest
