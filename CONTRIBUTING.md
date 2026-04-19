# Contributing

Thanks for helping improve **git-write-mcp**.

## Conduct and security

- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** applies to issues, discussions, and pull requests.
- Follow **[SECURITY.md](SECURITY.md)** for vulnerability reports—**do not** open public issues for undisclosed security defects.

## Branch and pull-request flow

This repository uses **`develop`** as the integration branch and **`main`** as the release line (see **[docs/workflow.md](docs/workflow.md)**).

1. Sync and branch from **`develop`**:

   ```bash
   git checkout develop && git pull
   git checkout -b feature/your-topic
   ```

2. Open pull requests **into `develop`**. Releases reach **`main`** only via a separate **release PR** managed by maintainers.

3. Prefer **test-driven development** for behavior changes (failing test first). See **[docs/development.md](docs/development.md)**.

## Match CI locally

Before pushing, run the same checks CI uses:

```bash
make ci   # check-fmt, vet, race+shuffle tests, golangci-lint, build
```

Minimum loop while iterating:

```bash
make test
make build
```

If **`make lint`** fails because **golangci-lint** was built with an older Go than **`go.mod`**, upgrade **golangci-lint** (see **[docs/development.md](docs/development.md)**).

## Documentation and changelog

- Update **[docs/](docs/README.md)** when you change tools, configuration, or workflow contracts.
- Add a note under **`[Unreleased]`** in **[CHANGELOG.md](CHANGELOG.md)** when the change is user-visible.

## Pull request hygiene

- Explain **what** changed and **why**; link issues when relevant.
- Keep commits focused; maintainers may squash on merge.

## License

By contributing, you agree your contributions are licensed under the project license (**Apache-2.0** — see **[LICENSE](LICENSE)**).
