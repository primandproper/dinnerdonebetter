package main

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

var (
	allTablesHat sync.Mutex
	allTables    = map[string]bool{}
)

func registerTableName(table string) {
	allTablesHat.Lock()
	defer allTablesHat.Unlock()
	allTables[table] = true
}

func getAllTables() []string {
	allTablesHat.Lock()
	defer allTablesHat.Unlock()

	tables := make([]string, 0, len(allTables))
	for t := range allTables {
		tables = append(tables, t)
	}

	slices.Sort(tables)

	return tables
}

func buildMaintenanceQueries(database string) []*Query {
	switch database {
	case postgres:
		queries := []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "DestroyAllData",
					Type: ExecType,
				},
				Content: fmt.Sprintf(`TRUNCATE %s CASCADE;`, strings.Join(getAllTables(), ", ")),
			},
		}
		return append(queries, buildQueueTestMessagesQueries(database)...)
	default:
		return nil
	}
}
