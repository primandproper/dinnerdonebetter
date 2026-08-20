package main

import (
	"github.com/primandproper/platform-go/v12/database/dialect"
	"github.com/primandproper/platform-go/v12/database/querygen"
)

// pgGen is the Postgres query generator. platform-go v12 moved the fragment and
// StandardCRUD builders onto a dialect-bound Generator; every builder in this
// package generates for Postgres only (each switches on `database` and returns
// nil for anything else), so one generator serves them all.
var pgGen = querygen.For(dialect.Postgres)
