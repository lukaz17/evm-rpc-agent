# Project Context

Part of the project memory: human-readable memory of what this project is.

The reading and writing rules for this file and the rest of `.attunic/` are defined in [attunic-protocol.md](../.agents/attunic-protocol.md). The user fills this file in, by hand or by asking an agent to help; change it only on such a request and after the user explicitly confirms that change (see [attunic-protocol.md's Write Rules](../.agents/attunic-protocol.md#write-rules)).

## Abstract

A newcomer should be able to tell what the software does without knowing what it is built with. One paragraph that describes the domain, not the technology, in simple terms:
<!-- Project-specific / Abstract -->

EVM RPC Agent sits between applications and an EVM-compatible blockchain node. It has two main features: it crawls blockchain data from the node and stores it in a database, and it serves RPC requests from that database first, passing a request through to the node when the data is not found or the request is not supported.

## Technical Specifications

A brief overview of the ingredients developers of this project should know.

Write one bullet per ingredient, as `- <Ingredient>: <value>.`, such as primary language, frameworks, architecture style, build system, test runner, linter, databases, environment prerequisites.

Write `None` as the value of an ingredient the project lacks but an agent would look for, such as a build system, test runner, or linter.

<!-- Project-specific / Technical Specifications -->

- Primary language: Go.
- Architecture style: message-passing services routed by `multiplex` from `github.com/tforce-io/tf-golib`.
- Database: MongoDB.
- Build system: go.
- Test runner: go.
- Linter: gofmt.
- Environment prerequisites: a MongoDB server and an EVM JSON-RPC node that exposes `debug_traceBlockByNumber`; the integration tests use MongoDB at `mongodb://localhost:27017` and the node at `http://localhost:8545` (unit tests need neither).

## Layout

Layout of this project; [project-structure.md](../.agents/project-structure.md) holds the standard layouts of its languages and frameworks, kept for reference.

List of important folders, files and their roles in the project:

Write the list as a tree in a fenced block, one entry per line with its role as a `#` comment, mirroring the actual file system.

When the list is not empty and a top-level file/folder is missing from it, prompt the user to update it.

<!-- Project-specific / Layout -->

```
├── .agents/            # project AI agent instructions (attunic-protocol, coding-conventions, development-guideline, license-header, memory-template, project-structure, pull-request-guideline)
├── .attunic/           # project AI agent memory (context, memory, question)
├── .opencode/          # local OpenCode agent files, not tracked by git
├── .vscode/            # local Visual Studio Code settings, not tracked by git
├── config/             # application configuration from defaults, YAML file, environment variables; logging setup
├── core/               # EVM primitive types (address, bytes, integer) with hex, JSON, BSON encoding
├── db/                 # MongoDB persistence of blocks and call traces
├── engine/             # wiring root of the program: cobra commands, app startup, service daemon lifecycle
├── mem/                # process-wide crawl state: current and target heights
├── rpc/                # JSON-RPC client and types for the EVM node
│   └── blobs/          # recorded RPC responses, test fixtures for rpc, db, svc
├── svc/                # crawl pipeline as multiplex services: scheduler, crawler, block and trace download, RPC workers, database writer, notifier
├── xlib/               # zerolog marshalers for slice types
├── .copier-answers.yml # project Copier answers
├── .editorconfig       # EditorConfig configuration
├── .gitignore          # git ignore patterns
├── AGENTS.md           # project AI agent entry point
├── CLAUDE.md           # project Claude Code entry point
├── COPYING             # GPL-3.0 license text
├── README.md           # project overview
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── main.go             # program entry point
```

## Layer Dependencies

List of architecture layers and their relationships:

- Declare the innermost architecture layer first.
- An architecture layer may depend only on itself and on the architecture layers in its `Depends on` column, all declared before it.
- If an architecture layer appears in the table below, the dependency graph must be enforced.

| Architecture layer | Path | Depends on | Must never contain |
|--------------------|------|------------|--------------------|
| <!-- Project-specific / Layer Dependencies --> | | | |

### Deliberate Exceptions

List of documented exceptions that cannot follow the architecture layers above:

- Any place the rules above are knowingly broken, with the reason and the condition for fixing it.
- Only exceptions to the architecture layers above belong here.
- An undocumented exception is a defect; a documented one is allowed, and revisited once its `Revisit-if` condition holds.

| What | Why | Revisit-if |
|------|-----|------------|
| <!-- Project-specific / Deliberate Exceptions --> | | |

## Supporting Tools

All the supporting tools declared here are opinionated for the sake of consistent results across development environments.

List each required tool/library under its own heading, followed by the commands for each use case.

Common use cases include: Build, Test, Lint, Format, Install.

<!-- Project-specific / Supporting Tools -->

### Build

- `go build -o ./build/evm-rpc-agent .` - build the program from the root package, output in `./build` (append `.exe` on Windows).

### Test

- `go test ./...` - run the unit test suite only; hermetic, needs no MongoDB server and no RPC node (integration tests are excluded by the `integration` build tag).
- `go test -tags=integration ./...` - run the whole suite, unit and integration; needs the MongoDB server and RPC node from [Technical Specifications](#technical-specifications), and clears the collections it uses in the `evm-rpc-agent-test` database.
- `go test -race ./...` / `go test -race -tags=integration ./...` - run the unit / whole suite with the race detector; required whenever a test touches a goroutine or shared state, per [development-guideline.md](../.agents/development-guideline.md).
- Narrower: `go test ./core/...`, `go test ./rpc/...`, `go test ./db/...`, `go test ./svc/...` run unit tests only and need neither service; with the tag, `go test -tags=integration ./db/...` needs MongoDB, `go test -tags=integration ./rpc/...` needs the RPC node, `go test -tags=integration ./svc/...` needs both.

### Lint

- `gofmt -l .` - check formatting; must report no files.
- `gofmt -w .` - apply canonical formatting.
- `go vet ./...` - go static analysis; run on every change that touches concurrency.
- `goimports -l .` / `goimports -w .` - check/fix import grouping (stdlib, then third-party, then local).
- `go mod tidy -diff` / `go mod tidy` - check/fix that `go.mod` and `go.sum` match the imports; the check must report no diff.

## Project-specific

<!-- Project-specific / Project-specific -->
