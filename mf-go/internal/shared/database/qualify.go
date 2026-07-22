package database

import (
	"regexp"
	"strings"
)

var postgresSchemaName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// QualifyTable returns schema.table when schema is set, otherwise table.
func QualifyTable(schema, table string) string {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return table
	}
	if !postgresSchemaName.MatchString(schema) {
		return table
	}
	return schema + "." + table
}
