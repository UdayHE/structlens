package model

// Table represents a relational table definition.
type Table struct {
	Name        string
	PrimaryKey  string
	Columns     []Column
	ForeignKeys []ForeignKey
	columnIndex map[string]struct{}
}

// Column represents a single table column.
type Column struct {
	Name     string
	Type     string
	Nullable bool
}

// ForeignKey represents a table relationship.
type ForeignKey struct {
	Column    string
	RefTable  string
	RefColumn string
}

func (t *Table) ensureColumnIndex() {
	if t.columnIndex != nil {
		return
	}
	t.columnIndex = make(map[string]struct{}, len(t.Columns))
	for _, column := range t.Columns {
		t.columnIndex[column.Name] = struct{}{}
	}
}

// AddColumn appends a column to the table.
func (t *Table) AddColumn(name, columnType string, nullable bool) {
	t.ensureColumnIndex()
	if _, exists := t.columnIndex[name]; exists {
		return
	}
	t.Columns = append(t.Columns, Column{
		Name:     name,
		Type:     columnType,
		Nullable: nullable,
	})
	t.columnIndex[name] = struct{}{}
}

// EnsurePrimaryKey adds a default primary key when the table does not have one.
func (t *Table) EnsurePrimaryKey() {
	if t.PrimaryKey == "" {
		t.PrimaryKey = "id"
	}
	t.AddColumn(t.PrimaryKey, "BIGINT", false)
}

// AddForeignKey appends a foreign key definition to the table.
func (t *Table) AddForeignKey(column, refTable, refColumn string) {
	t.ForeignKeys = append(t.ForeignKeys, ForeignKey{
		Column:    column,
		RefTable:  refTable,
		RefColumn: refColumn,
	})
}
