# MCP clients & Cursor setup

git-write-mcp speaks **stdio MCP** only: stdin/stdout **newline-delimited JSON-RPC** (`mcp.StdioTransport` from **`github.com/modelcontextprotocol/go-sdk`**). Any host (Cursor, Claude Desktop, custom glue) launches the binary as a subprocess and forwards stdio.

---

## Cursor (`mcp.json`)

Register a server block similar to:

```json
{
  "mcpServers": {
    "git-write-mcp": {
      "command": "/home/you/.local/bin/git-write-mcp",
      "type": "stdio",
      "args": [],
      "env": {
        "GIT_WRITE_MCP_REPO_ROOT": "/absolute/path/to/your/repo",
        "GIT_WRITE_MCP_AUTO_APPROVE": "0",
        "GITHUB_TOKEN": "",
        "GITHUB_REPOSITORY": "owner/name"
      }
    }
  }
}
```

Notes:

- **`GIT_WRITE_MCP_REPO_ROOT`** must point at **one** clone at a time. Switching workspaces usually means adjusting this path or using a **workspace-local** MCP override if your client supports it.
- **`GIT_WRITE_MCP_AUTO_APPROVE`**: **`1`** lets agents bypass the deny-all approver (still subject to repo-state checks). Use **`0`** for safer defaults except in throwaway clones.
- **`GITHUB_TOKEN`** / **`GITHUB_REPOSITORY`**: required at **process start** for GitHub tools to register **real** handlers. Omitting them yields configuration errors or **stub** `"not implemented"` responses for **`merge_pull_request`** / **`get_pull_request_status`** / **`create_release_from_main`** depending on registration order—verify **`get_capabilities`** after restart.

After editing MCP config, **restart MCP servers** or reload the window.

---

## Isolated sandbox directories

For experiments, **do not point** **`GIT_WRITE_MCP_REPO_ROOT`** at production clones.

1. **`mktemp -d`**, **`git init`**, optional bare **`origin`** for **`push_branch`** demos.
2. Optionally **`git config core.hooksPath`** to **disable hooks** so global commit-msg hooks do not affect automation.
3. Swap **`GIT_WRITE_MCP_REPO_ROOT`** in MCP env to that directory; restart MCP.

---

## Chat agent limitations

Coding agents call MCP **only when the host attaches tool descriptors** to the session. If **`call_mcp_tool`** is unavailable, rely on **`go test ./internal/server/...`** or run the binary manually with env vars.

---

## Capability discovery

Clients should call **`get_capabilities`** first—the JSON lists **`write_tools`**, **`read_tools`**, **`behaviors`**, and **`excluded_read_tools`** for delegation to read Git tooling.
