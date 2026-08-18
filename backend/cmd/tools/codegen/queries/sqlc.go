package main

import (
	"github.com/primandproper/platform-go/v11/database/querygen"
)

// The sqlc input types live in platform's querygen, which emits the standard
// queries as values of them. Aliasing rather than converting is what lets a
// table's bespoke queries and its StandardCRUD call go into the same slice.
type (
	// QueryType is the sqlc annotation suffix declaring what a query returns.
	QueryType = querygen.QueryType
	// QueryAnnotation is the `-- name: X :one` line sqlc reads above a query.
	QueryAnnotation = querygen.QueryAnnotation
	// Query is one annotated statement.
	Query = querygen.Query
)

const (
	ExecType     = querygen.ExecType
	ExecRowsType = querygen.ExecRowsType
	ManyType     = querygen.ManyType
	OneType      = querygen.OneType
)
