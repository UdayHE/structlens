package export

import (
	"fmt"
	"strings"

	"structlens/internal/model"
)

const indent = "  "

// GenerateSQL renders PostgreSQL-compatible CREATE TABLE statements.
func GenerateSQL(tables []model.Table) (string, error) {
	if len(tables) == 0 {
		return "", nil
	}

	orderedTables, err := sortTablesByDependencies(tables)
	if err != nil {
		return "", err
	}

	statements := make([]string, 0, len(orderedTables))
	for _, table := range orderedTables {
		table.EnsurePrimaryKey()
		statement, err := buildCreateTableStatement(table)
		if err != nil {
			return "", err
		}
		statements = append(statements, statement)
	}

	return strings.Join(statements, "\n\n"), nil
}

func buildCreateTableStatement(table model.Table) (string, error) {
	if table.Name == "" {
		return "", fmt.Errorf("table name is required")
	}

	lines := make([]string, 0, len(table.Columns)+len(table.ForeignKeys))
	for _, column := range table.Columns {
		lines = append(lines, formatColumnLine(column, table.PrimaryKey))
	}
	for _, foreignKey := range table.ForeignKeys {
		lines = append(lines, formatForeignKeyLine(foreignKey))
	}

	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(quoteIdent(table.Name))
	b.WriteString(" (\n")
	for i, line := range lines {
		b.WriteString(indent)
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(");")

	return b.String(), nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func formatColumnLine(column model.Column, primaryKey string) string {
	columnType := normalizeColumnType(column.Type, column.Name == primaryKey)
	if column.Name == primaryKey {
		return fmt.Sprintf("%s %s PRIMARY KEY", quoteIdent(column.Name), columnType)
	}
	return fmt.Sprintf("%s %s", quoteIdent(column.Name), columnType)
}

func formatForeignKeyLine(foreignKey model.ForeignKey) string {
	return fmt.Sprintf(
		"FOREIGN KEY (%s) REFERENCES %s(%s)",
		quoteIdent(foreignKey.Column),
		quoteIdent(foreignKey.RefTable),
		quoteIdent(foreignKey.RefColumn),
	)
}

func normalizeColumnType(columnType string, isPrimaryKey bool) string {
	normalized := strings.ToUpper(strings.TrimSpace(columnType))
	switch normalized {
	case "", "STRING", "TEXT":
		return "TEXT"
	case "INT", "INTEGER", "BIGINT":
		return "BIGINT"
	case "FLOAT", "DOUBLE", "DOUBLE PRECISION", "NUMBER":
		if isPrimaryKey {
			return "BIGINT"
		}
		return "DOUBLE PRECISION"
	case "BOOL", "BOOLEAN":
		return "BOOLEAN"
	default:
		if isPrimaryKey {
			return "BIGINT"
		}
		return normalized
	}
}

func sortTablesByDependencies(tables []model.Table) ([]model.Table, error) {
	if len(tables) <= 1 {
		return tables, nil
	}

	tablesByName := make(map[string]model.Table, len(tables))
	dependencyCount := make(map[string]int, len(tables))
	dependents := make(map[string][]string, len(tables))
	order := make([]string, 0, len(tables))

	for _, table := range tables {
		tablesByName[table.Name] = table
		dependencyCount[table.Name] = 0
		order = append(order, table.Name)
	}

	for _, table := range tables {
		seenRefs := make(map[string]struct{}, len(table.ForeignKeys))
		for _, foreignKey := range table.ForeignKeys {
			if foreignKey.RefTable == "" || foreignKey.RefTable == table.Name {
				continue
			}
			if _, ok := tablesByName[foreignKey.RefTable]; !ok {
				continue
			}
			if _, seen := seenRefs[foreignKey.RefTable]; seen {
				continue
			}
			seenRefs[foreignKey.RefTable] = struct{}{}
			dependencyCount[table.Name]++
			dependents[foreignKey.RefTable] = append(dependents[foreignKey.RefTable], table.Name)
		}
	}

	queue := make([]string, 0, len(tables))
	queued := make(map[string]struct{}, len(tables))
	for _, name := range order {
		if dependencyCount[name] == 0 {
			queue = append(queue, name)
			queued[name] = struct{}{}
		}
	}

	sorted := make([]model.Table, 0, len(tables))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		sorted = append(sorted, tablesByName[name])

		for _, dependent := range dependents[name] {
			dependencyCount[dependent]--
			if dependencyCount[dependent] == 0 {
				if _, exists := queued[dependent]; exists {
					continue
				}
				queue = append(queue, dependent)
				queued[dependent] = struct{}{}
			}
		}
	}

	if len(sorted) != len(tables) {
		return nil, fmt.Errorf("cyclic table dependency detected")
	}

	return sorted, nil
}
