# git-write-mcp

stdio **[Model Context Protocol](https://modelcontextprotocol.io/)** server that exposes **controlled, write-capable Git** operations (and optional GitHub REST) for agents on **Linux** and **macOS**. Mutations use **allowlisted `git` argv only**—never arbitrary shell execution.

**Upstream:** https://github.com/ipolyzos/git-write-mcp · **Contact:** git@ipolyzos.com  

Licensed under **Apache-2.0**. Copyright **2026** **Ioannis Polyzos**.

---

## Community & legal

GitHub surfaces these files in the repository **About** / community health UI when present at the repo root:

| Resource | File |
|----------|------|
| **License** | [`LICENSE`](LICENSE) (Apache-2.0 full text) |
| **Code of Conduct** | [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) (Contributor Covenant 2.1) |
| **Security** | [`SECURITY.md`](SECURITY.md) — **report vulnerabilities privately** (do not open public issues for unfixed flaws) |

---

## Documentation

All product, workflow, configuration, and API reference documentation lives under **`docs/`**:

**→ [docs/README.md](docs/README.md)** — documentation hub (overview, safety model, configuration, MCP client setup, tool reference, changelog link).

Quick links: [Overview](docs/overview.md) · [Configuration](docs/configuration.md) · [Workflow](docs/workflow.md) · [MCP tool reference](docs/reference/mcp-tools.md) · [GitHub integration](docs/github-integration.md) · [Changelog](CHANGELOG.md)

---

## Build (from source)

```bash
make build   # outputs bin/git-write-mcp
make test
```

See **[docs/development.md](docs/development.md)** for **`make ci`**, race/shuffle tests, and lint requirements.
