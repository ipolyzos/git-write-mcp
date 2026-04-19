# Security policy

## Supported versions

Security fixes are applied to the **current integration line** (`develop`) and released from **`main`** per [Semantic Versioning](https://semver.org/). Patch releases may tag older minors when feasible; details appear in [CHANGELOG.md](CHANGELOG.md).

| Version line | Supported |
|--------------|-----------|
| Latest release (`main` / tagged) | Yes |
| `develop` (pre-release integration) | Yes |
| Older minors without maintainer agreement | Best effort |

---

## Reporting a vulnerability

**Please do not** open a public GitHub issue for unfixed security defects—doing so can put users at risk before a fix ships.

Preferred channels:

1. **GitHub Security Advisories** — use **[Report a vulnerability](https://github.com/ipolyzos/git-write-mcp/security/advisories/new)** on this repository (private submission to maintainers).
2. **Email** — send details to **git@ipolyzos.com** with subject prefix **`[SECURITY]`**.

Include:

- A concise description and affected component (`internal/git`, MCP tool name, GitHub integration, etc.).
- Steps to reproduce or proof-of-concept **without** dropping live exploit code into public threads before coordinated disclosure.
- Impacted versions or commits if known.

You should receive an acknowledgment within a few business days. We may coordinate embargo and disclosure timing with you before publishing an advisory or release note.

---

## Product security notes

Operational guidance for approvals, tokens, and repo boundaries is documented under **[docs/security.md](docs/security.md)** (behavior of this MCP server—not the same as vulnerability disclosure).
