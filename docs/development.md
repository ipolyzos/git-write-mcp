# Development

---

## Test-driven workflow

For **code** changes (behavior or contracts):

1. Define scope.
2. Add **failing tests** first (compile-fail or assertion fail).
3. Implement the **minimum** to pass.
4. Run **`make ci`** or the same steps CI uses (`check-fmt`, `vet`, **`make test`** with race + shuffle, `lint`, **`make build`**).
5. Update documentation when behavior or contracts change—see **`docs/`** index.

---

## Build & test targets

```bash
make build    # produces bin/git-write-mcp
make test     # go test -race -count=1 -shuffle=on ./...
make ci       # check-fmt, vet, test, golangci-lint run, build
```

**`make test`** / **`make ci`** align with GitHub Actions: **`go test -race -count=1 -shuffle=on ./...`** (race detector + shuffled order).

---

## Lint toolchain

**`make lint`** requires **`golangci-lint`** compatible with the **Go toolchain** declared in **`go.mod`**. If lint fails with *“Go language version used to build golangci-lint is lower than the targeted Go version”*, upgrade **`golangci-lint`** (or install via the **[install page](https://golangci-lint.run/welcome/install/)**) so it matches your local Go minor version.

---

## MCP integration test

**`internal/server`** includes **`TestMCP_AllRegisteredTools_AutoApprove`**, which exercises every **`ListTools`** entry over an isolated temporary Git repository with **`GIT_WRITE_MCP_AUTO_APPROVE=1`** (dry-run where appropriate). Run:

```bash
go test -count=1 -v ./internal/server/... -run TestMCP_AllRegisteredTools_AutoApprove
```
