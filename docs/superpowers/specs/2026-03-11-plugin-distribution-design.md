# Plugin Distribution for Oraculo

## Context

Oraculo distributes skills by embedding them in the Go binary and copying to `.claude/skills/` during `oraculo install`. This requires manual re-installation for updates and couples skill distribution to binary releases.

Claude Code now supports a plugin system with auto-updates, namespaced skills, and MCP server bundling. This design adds plugin-based distribution as a second channel alongside the existing CLI install.

## Decision: Dual Distribution

Both channels coexist — CLI install (legacy) and Claude Code plugin (new). The user chooses one.

| Channel | Skills | MCP | DB/Hooks |
|---------|--------|-----|----------|
| `oraculo install` | Embedded in binary, copied to `.claude/skills/` | Registered in `.claude/settings.json` | Yes |
| Plugin + `oraculo setup` | Delivered by plugin, auto-updated | Delivered by plugin via npx | `oraculo setup` handles DB/hooks only |

## 1. Restructure `claude-kit/`

### Before

```
claude-kit/
├── embed.go
└── skills/
    └── oraculo/
        ├── epic/SKILL.md
        ├── story/SKILL.md
        └── ...
```

### After

```
claude-kit/
├── .claude-plugin/
│   └── plugin.json
├── .mcp.json
├── embed.go
└── skills/
    ├── epic/SKILL.md + phases/ + references/
    ├── story/SKILL.md + phases/ + references/
    ├── plan/SKILL.md + phases/ + references/
    ├── execute/SKILL.md + phases/ + references/
    ├── validate/SKILL.md + phases/ + references/
    └── tools/SKILL.md
```

The `oraculo/` nesting level is removed. Skills move from `skills/oraculo/X/` to `skills/X/`.

### Adaptations

- `embed.go`: Update embed path from `skills/oraculo` to `skills`
- `install.go`: Update source path; destination stays `.claude/skills/oraculo-X/` (prefix added at copy time)
- `install_test.go`: Update assertions for new paths
- `uninstall.go`: No change (already removes `.claude/skills/oraculo*`)

## 2. SKILL.md Frontmatter

Remove `oraculo:` prefix from `name` field in all SKILL.md files:

```yaml
# Before
name: "oraculo:epic"

# After
name: "epic"
```

- **Plugin path**: Claude Code auto-prefixes with plugin name -> `/oraculo:epic`
- **CLI path**: Directory name `oraculo-epic` provides the namespace

## 3. Plugin Manifest

File: `claude-kit/.claude-plugin/plugin.json`

```json
{
  "name": "oraculo",
  "version": "0.1.0",
  "description": "Socratic guide for quality product development",
  "author": {
    "name": "Lucas"
  },
  "repository": "https://github.com/lucas/oraculo",
  "skills": "./skills/"
}
```

## 4. MCP Configuration

File: `claude-kit/.mcp.json`

```json
{
  "mcpServers": {
    "oraculo": {
      "command": "npx",
      "args": ["-y", "@oraculo/cli@latest", "start"]
    }
  }
}
```

The plugin starts the MCP server via npx. The npm package `@oraculo/cli` resolves to a platform-specific Go binary (see section 5).

## 5. npm Packages

### Package structure in repo

```
npm/
├── cli/
│   ├── package.json        # @oraculo/cli — launcher + optionalDependencies
│   └── bin/cli.js          # Resolves platform binary, forwards args
├── cli-darwin-arm64/
│   └── package.json        # os: ["darwin"], cpu: ["arm64"]
├── cli-darwin-x64/
│   └── package.json        # os: ["darwin"], cpu: ["x64"]
├── cli-linux-x64/
│   └── package.json        # os: ["linux"], cpu: ["x64"]
└── cli-linux-arm64/
    └── package.json        # os: ["linux"], cpu: ["arm64"]
```

### Main package (`@oraculo/cli`)

- `bin/cli.js`: Node script that resolves the platform-specific binary and `execFileSync` with forwarded args
- `optionalDependencies`: Lists all platform packages
- `postinstall`: Fallback download if optionalDependencies didn't install (corporate proxies, `--ignore-optional`)

### Platform packages (`@oraculo/cli-{os}-{arch}`)

Each contains only `package.json` (with `os`/`cpu` fields) and `bin/oraculo` (the compiled Go binary). npm installs only the matching platform.

### Cross-compilation

```bash
GOOS=darwin  GOARCH=arm64 go build -o dist/darwin-arm64/oraculo  ./cmd/oraculo
GOOS=darwin  GOARCH=amd64 go build -o dist/darwin-x64/oraculo    ./cmd/oraculo
GOOS=linux   GOARCH=amd64 go build -o dist/linux-x64/oraculo     ./cmd/oraculo
GOOS=linux   GOARCH=arm64 go build -o dist/linux-arm64/oraculo   ./cmd/oraculo
```

## 6. GitHub Actions

File: `.github/workflows/publish-npm.yml`

Trigger: Push tag `v*`

Steps:
1. Checkout repo
2. Parse version from tag
3. Cross-compile Go binary for 4 platforms
4. Copy binaries into `npm/cli-{platform}/bin/`
5. Set version in all `package.json` files
6. `npm publish` each platform package
7. `npm publish` main `@oraculo/cli` package

Requires: `NPM_TOKEN` secret in GitHub repo settings.

## 7. Marketplace

Separate repo: `lucas/oraculo-marketplace`

File: `.claude-plugin/marketplace.json`

```json
{
  "name": "oraculo-marketplace",
  "owner": { "name": "Lucas" },
  "metadata": {
    "description": "Oraculo plugin marketplace"
  },
  "plugins": [
    {
      "name": "oraculo",
      "source": {
        "source": "git-subdir",
        "url": "https://github.com/lucas/oraculo.git",
        "path": "claude-kit"
      },
      "description": "Socratic guide for quality product development"
    }
  ]
}
```

User installation:
```
/plugin marketplace add lucas/oraculo-marketplace
/plugin install oraculo
```

## 8. New CLI Command: `oraculo setup`

Subset of `oraculo install` — only infrastructure, no skills or MCP:

- Create `.oraculo/` directory
- Write `config.json` with port and language
- Initialize SQLite database
- Register hooks in `.claude/settings.json`

Does NOT:
- Copy skills to `.claude/skills/`
- Register MCP server in `.claude/settings.json`

For users who install the plugin, the flow is:
```
1. /plugin install oraculo          # skills + MCP
2. oraculo setup                     # DB + hooks
```

## 9. `oraculo install` (Legacy, Unchanged Behavior)

Continues to work as before. Internal changes:
- Source path updated from `skills/oraculo/X/` to `skills/X/`
- Destination unchanged: `.claude/skills/oraculo-X/`
- All other behavior (DB, hooks, MCP registration) stays the same

## 10. User Flows

### New user (plugin)
```
/plugin marketplace add lucas/oraculo-marketplace
/plugin install oraculo
oraculo setup
```

### New user (CLI, legacy)
```
oraculo install
```

### Existing user migrating to plugin
```
oraculo uninstall          # removes skills + MCP config
/plugin install oraculo    # plugin provides skills + MCP
oraculo setup              # re-creates DB + hooks (if purged)
```

## Out of Scope

- Windows support (can add later)
- Official Claude Code marketplace submission (future milestone)
- Removing `oraculo install` (kept indefinitely as fallback)
- Plugin hooks (Claude Code plugin hooks, not Oraculo hooks)
