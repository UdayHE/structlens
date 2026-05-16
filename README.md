# StructLens

Convert structured data to a relational schema instantly.

StructLens is a desktop developer tool (macOS) that reads JSON or XML, infers a schema, and maps nested data into relational tables with PostgreSQL-compatible SQL. It ships as a native app built with Tauri + React, backed by a Go engine.

## What It Does

- Opens JSON and XML files via a native file picker
- Streams-parses input — handles files up to 100MB
- Infers field types, optional fields, and arrays
- Maps nested objects and repeated structures into relational tables
- Generates deterministic PostgreSQL-style `CREATE TABLE` statements with foreign keys
- Displays a live schema tree you can explore and click
- Shows a split-pane workspace with the tree on the left and SQL on the right
- Lists sample records grouped by inferred table type

## Desktop App

### Requirements

- macOS (arm64 or x86_64)
- [Rust + Cargo](https://rustup.rs/)
- [Node.js](https://nodejs.org/) (v18+)
- Go 1.24+

### Run in development

```bash
npm install
npm run tauri dev
```

This builds the Go sidecar binary, starts the Vite dev server, and launches the Tauri window.

### Build for production

```bash
npm run tauri build
```

The `.app` bundle is written to `src-tauri/target/release/bundle/macos/`.

### How the sidecar works

The Go engine is compiled to a platform-specific binary under `src-tauri/binaries/` and invoked by Tauri via its sidecar mechanism. The build script handles this automatically:

```bash
./scripts/build-sidecar.sh          # auto-detects host target triple
./scripts/build-sidecar.sh aarch64-apple-darwin
./scripts/build-sidecar.sh x86_64-apple-darwin
```

## CLI

The original CLI is still available for scripting and CI use.

Build:

```bash
go build -o structlens ./cmd/structlens
```

Generate SQL from a file:

```bash
./structlens examples/nested.json
./structlens examples/complex.xml
```

Inspect the inferred schema as a tree:

```bash
./structlens --view tree examples/nested.json
```

Show version:

```bash
./structlens --version
```

### CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--view` | `sql` | Output mode: `sql` or `tree` |
| `--flatten-threshold` | `2` | Max fields in a nested object before it becomes its own table |
| `--array-item-name` | `item` | Name used for unnamed array elements |
| `--version` | — | Print version and exit |
| `--help` | — | Show usage |

## Example

Input `examples/nested.json`:

```json
{
  "order": {
    "id": 1,
    "customer": { "name": "John" },
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

```
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

## How Mapping Works

- Root object becomes the main table
- Arrays become child tables with a foreign key back to the parent
- Nested objects with few fields are flattened into the parent table (configurable via `--flatten-threshold`)
- Every table gets an auto-generated `id BIGINT PRIMARY KEY`
- Output column and table names are converted to `snake_case`
- Type conflicts resolve to `TEXT`; missing fields are marked optional

## Architecture

```
[File picker / CLI input]
        │
        ▼
   Go Engine (sidecar / CLI)
   ┌──────────────────────────────────────────┐
   │  Parser → Inference → Mapper → Exporter  │
   └──────────────────────────────────────────┘
        │
        ▼
   Tauri IPC  ──►  React UI
                   ├── Schema Tree View
                   ├── SQL Viewer
                   └── Records View
```

Main Go packages:

| Package | Responsibility |
|---|---|
| `internal/parser` | Streaming JSON / XML parsing |
| `internal/inference` | Field type and structure inference |
| `internal/mapper` | Relational table mapping |
| `internal/export` | SQL generation |
| `internal/view` | Tree printer |
| `cmd/structlens` | CLI entry point |
| `cmd/structlens-engine` | Sidecar entry point (JSON IPC) |

## Development

Run all Go tests:

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

Validate the engine output directly:

```bash
go run ./cmd/structlens examples/order.json
go run ./cmd/structlens --view tree examples/order.json
```

## Example Files

- [examples/simple.json](examples/simple.json)
- [examples/nested.json](examples/nested.json)
- [examples/complex.xml](examples/complex.xml)
- [examples/books.xml](examples/books.xml)

Expected SQL snapshots:

- [examples/simple.sql](examples/simple.sql)
- [examples/nested.sql](examples/nested.sql)
- [examples/complex.sql](examples/complex.sql)

## Supported Inputs

- JSON
- XML

Not supported yet:

- YAML
- CSV
- Schema diff / migrations

## Roadmap

- Better numeric precision handling (`int` vs `float`)
- YAML support
- Schema diff support
- Windows / Linux builds
