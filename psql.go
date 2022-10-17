// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serialkey

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed psql_next.sql
var postgreSQLNext []byte

func (db PostgreSQL) next() (string, error) {
	return db.generate(string(postgreSQLNext))
}

//go:embed psql_next_n.sql
var postgreSQLNextN []byte

func (db PostgreSQL) nextN() (string, error) {
	return db.generate(string(postgreSQLNextN))
}

//go:embed psql_last.sql
var postgreSQLLast []byte

func (db PostgreSQL) last() (string, error) {
	return db.generate(string(postgreSQLLast))
}

//go:embed psql_forward.sql
var postgreSQLForward []byte

func (db PostgreSQL) forward() (string, error) {
	return db.generate(string(postgreSQLForward))
}

//go:embed psql_create_table.sql
var PostgreSQLCreateTable []byte

func (db PostgreSQL) createTable() (string, error) {
	return db.generate(string(PostgreSQLCreateTable))
}

func (db PostgreSQL) generate(query string) (string, error) {
	tmpl, err := template.New("postgresql").Parse(query)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer

	err = tmpl.ExecuteTemplate(&buf, "postgresql", db)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

type PostgreSQL struct {
	Table string
}
