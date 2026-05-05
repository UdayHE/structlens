# StructLens

Convert structured data to a relational schema instantly.

StructLens is a Go CLI that reads JSON or XML, infers a schema, maps nested data into relational tables, and generates PostgreSQL-compatible SQL. It can also print the inferred structure as a readable tree so you can inspect the shape before generating tables.

## What It Does

- Parses JSON and XML with streaming parsers
- Infers field types, optional fields, and arrays
- Maps nested objects and repeated structures into relational tables
- Generates deterministic PostgreSQL-style `CREATE TABLE` statements
- Prints a tree view of the inferred schema for quick inspection

## Quick Start

Build the CLI:

```bash
go build -o structlens ./cmd/structlens
```

Generate SQL from JSON:

```bash
./structlens examples/nested.json
```

Inspect the inferred structure instead of SQL:

```bash
./structlens --view tree examples/nested.json
```

Show version:

```bash
./structlens --version
```

## Example

Input:

```json
{
  "order": {
    "id": 1,
    "customer": {
      "name": "John"
    },
    "items": [
      { "product": "A", "qty": 2 },
      { "product": "B", "qty": 1 }
    ]
  }
}
```

SQL output:

```sql
CREATE TABLE orders (
  id BIGINT PRIMARY KEY,
  customer_name TEXT
);

CREATE TABLE items (
  id BIGINT PRIMARY KEY,
  order_id BIGINT,
  product TEXT,
  qty DOUBLE PRECISION,
  FOREIGN KEY (order_id) REFERENCES orders(id)
);
```

Tree output:

```text
Schema Summary:
- Total fields: 7
- Arrays: 1
- Optional fields: 0

Root: order (3 fields)
├── id (int)
├── customer (1 fields)
│   └── name (string)
└── items[] (array, 2 fields)
    ├── product (string)
    └── qty (int)
```

## Installation

### From source

```bash
go build -o structlens ./cmd/structlens
```

### Run without building

```bash
go run ./cmd/structlens examples/simple.json
```

## Usage

```bash
structlens [flags] <input.json|input.xml>
```

### Flags

- `--view sql|tree`
  Default: `sql`
- `--flatten-threshold <n>`
  Default: `2`
- `--array-item-name <name>`
  Default: `item`
- `--version`
- `--help`

### Common commands

```bash
structlens examples/simple.json
structlens --flatten-threshold 1 examples/nested.json
structlens --view tree examples/nested.json
structlens --array-item-name entry examples/complex.xml
```

## How Mapping Works

- Root object becomes the main table
- Arrays become child tables
- Nested objects may be flattened into the parent table
- Every table gets an `id BIGINT PRIMARY KEY`
- Foreign keys are created for child tables
- Output column and table names are converted to `snake_case`

## Supported Inputs

- JSON
- XML

Unsupported today:

- YAML
- CSV
- Schema diff / migrations

## Example Files

Bundled examples:

- [examples/simple.json](/Users/udayhegde/GoProjects/structlens/examples/simple.json)
- [examples/nested.json](/Users/udayhegde/GoProjects/structlens/examples/nested.json)
- [examples/complex.xml](/Users/udayhegde/GoProjects/structlens/examples/complex.xml)

Expected SQL snapshots:

- [examples/simple.sql](/Users/udayhegde/GoProjects/structlens/examples/simple.sql)
- [examples/nested.sql](/Users/udayhegde/GoProjects/structlens/examples/nested.sql)
- [examples/complex.sql](/Users/udayhegde/GoProjects/structlens/examples/complex.sql)

## Architecture

The current pipeline is:

```text
Parser -> Inference -> Mapper -> SQL Generator
                   -> Tree Printer
```

Main packages:

- `internal/parser`
- `internal/inference`
- `internal/mapper`
- `internal/export`
- `internal/view`
- `cmd/structlens`

## Development

Run tests:

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

Run the sample fixture used by CLI tests:

```bash
go run ./cmd/structlens examples/order.json
go run ./cmd/structlens --view tree examples/order.json
```

## Limits And Notes

- Intended to handle files up to 100MB
- Parsers are streaming-oriented; the tool avoids loading raw input with ad hoc file parsing logic
- SQL output targets PostgreSQL-style DDL
- Tree output is for inspection only and does not change inference or mapping behavior
- Current numeric inference displays `int` in the tree and emits `DOUBLE PRECISION` for numeric SQL columns unless the column is a primary or foreign key

## Roadmap

- Better numeric precision handling (`int` vs `float`)
- YAML support
- Schema diff support
- Desktop UI with Tauri + React
