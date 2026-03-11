# Plugin Distribution Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable Oraculo distribution as a Claude Code plugin with Go binary via npm, while keeping the existing CLI install path working.

**Architecture:** The `claude-kit/` directory becomes a Claude Code plugin root. Skills move from `skills/oraculo/X/` to `skills/X/`. The Go binary is distributed via npm using the `optionalDependencies` pattern (esbuild/turbo). A new `oraculo setup` command handles DB/hooks without skills/MCP.

**Tech Stack:** Go (CLI), Node.js (npm launcher), GitHub Actions (CI), npm registry (@oraculo org)

**Spec:** `docs/superpowers/specs/2026-03-11-plugin-distribution-design.md`

---

## Chunk 1: Restructure `claude-kit/` Directory

Move skills from `skills/oraculo/X/` to `skills/X/` and update all Go code that references the old paths.

### Task 1: Move skill directories

**Files:**
- Move: `claude-kit/skills/oraculo/epic/` → `claude-kit/skills/epic/`
- Move: `claude-kit/skills/oraculo/story/` → `claude-kit/skills/story/`
- Move: `claude-kit/skills/oraculo/plan/` → `claude-kit/skills/plan/`
- Move: `claude-kit/skills/oraculo/execute/` → `claude-kit/skills/execute/`
- Move: `claude-kit/skills/oraculo/validate/` → `claude-kit/skills/validate/`
- Move: `claude-kit/skills/oraculo/tools/` → `claude-kit/skills/tools/`
- Delete: `claude-kit/skills/oraculo/` (empty after moves)
- Keep: `claude-kit/skills/oraculo/cli-workspace/` stays (eval data, not a skill)

- [ ] **Step 1: Move skill directories**

```bash
cd claude-kit/skills
for skill in epic story plan execute validate tools; do
  mv oraculo/$skill ./$skill
done
```

- [ ] **Step 2: Move cli-workspace to project root**

The `cli-workspace/` is eval data, not a skill. Move it out of the skills tree:

```bash
mv claude-kit/skills/oraculo/cli-workspace _sandbox/cli-workspace
```

- [ ] **Step 3: Remove empty oraculo directory**

```bash
rmdir claude-kit/skills/oraculo
```

- [ ] **Step 4: Verify structure**

```bash
ls claude-kit/skills/
# Expected: epic  execute  plan  story  tools  validate
```

- [ ] **Step 5: Commit**

```bash
git add -A claude-kit/skills/ _sandbox/cli-workspace
git commit -m "refactor: move skills from skills/oraculo/X/ to skills/X/

Flatten the directory structure so claude-kit/ can serve as a
Claude Code plugin root. The plugin auto-namespaces skills as
oraculo:X based on the plugin name."
```

### Task 2: Update SKILL.md frontmatter — remove `oraculo:` prefix

**Files:**
- Modify: `claude-kit/skills/epic/SKILL.md`
- Modify: `claude-kit/skills/story/SKILL.md`
- Modify: `claude-kit/skills/plan/SKILL.md`
- Modify: `claude-kit/skills/execute/SKILL.md`
- Modify: `claude-kit/skills/validate/SKILL.md`
- Modify: `claude-kit/skills/tools/SKILL.md`

- [ ] **Step 1: Update each SKILL.md**

In each file, change the `name:` field in the YAML frontmatter:

| File | Before | After |
|------|--------|-------|
| epic/SKILL.md | `name: oraculo:epic` | `name: epic` |
| story/SKILL.md | `name: oraculo:story` | `name: story` |
| plan/SKILL.md | `name: oraculo:plan` | `name: plan` |
| execute/SKILL.md | `name: oraculo:execute` | `name: execute` |
| validate/SKILL.md | `name: oraculo:validate` | `name: validate` |
| tools/SKILL.md | `name: oraculo:tools` | `name: tools` |

- [ ] **Step 2: Update cross-references in SKILL.md content**

Search all SKILL.md and phase files for references like `/oraculo:cli`, `/oraculo:tools` and update them to `/oraculo:tools` → `/tools` (the plugin namespace is automatic).

```bash
grep -r "oraculo:" claude-kit/skills/ --include="*.md" -l
```

For each match, update:
- `oraculo:epic` → `epic` (in name fields and cross-references)
- `oraculo:story` → `story`
- `oraculo:plan` → `plan`
- `oraculo:execute` → `execute`
- `oraculo:validate` → `validate`
- `oraculo:tools` → `tools`
- `/oraculo:cli` → `/tools` (or whatever the actual reference target is)

**Important:** Only change skill references (slash commands). Do NOT change CLI commands like `oraculo tools ...` or `oraculo status` — those are the binary, not skills.

- [ ] **Step 3: Commit**

```bash
git add claude-kit/skills/
git commit -m "refactor: remove oraculo: prefix from SKILL.md names

Plugin auto-namespaces as oraculo:X. CLI install adds the
oraculo- prefix at copy time via directory naming."
```

### Task 3: Update `embed.go` and `install.go`

**Files:**
- Modify: `claude-kit/embed.go`
- Modify: `apps/backend/src/cli/install.go:137-151`

- [ ] **Step 1: Write the failing test**

The existing test `TestInstall_CopiesEmbeddedOraculoSkills` in `apps/backend/src/cli/install_test.go` walks `skills/oraculo/` in the embed FS. After restructuring, it should walk `skills/` directly. Update the test first:

```go
// In install_test.go, update TestInstall_CopiesEmbeddedOraculoSkills:
// Change the walk root from filepath.Join("skills", "oraculo") to "skills"
// Change the rel calculation to be relative to "skills"
// The destination mapping: skills/epic/SKILL.md → .claude/skills/oraculo-epic/SKILL.md

func TestInstall_CopiesEmbeddedOraculoSkills(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	skillsSrcRoot := "skills"
	err = fs.WalkDir(claudekit.SkillsFS, skillsSrcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		// path: skills/epic/SKILL.md
		// rel:  epic/SKILL.md
		rel, err := filepath.Rel(skillsSrcRoot, path)
		if err != nil {
			return err
		}
		// skills/epic/SKILL.md → .claude/skills/oraculo-epic/SKILL.md
		parts := strings.SplitN(rel, string(filepath.Separator), 2)
		var destPath string
		if len(parts) == 2 {
			destPath = filepath.Join(".claude", "skills", "oraculo-"+parts[0], parts[1])
		} else {
			destPath = filepath.Join(".claude", "skills", "oraculo-"+parts[0])
		}
		got, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("expected %s to exist: %v", destPath, err)
			return nil
		}
		want, err := fs.ReadFile(claudekit.SkillsFS, path)
		if err != nil {
			return err
		}
		if string(got) != string(want) {
			t.Errorf("content mismatch for %s", destPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded FS: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -count=1 -run TestInstall_CopiesEmbeddedOraculoSkills ./apps/backend/src/cli/`
Expected: FAIL — embed FS still has old structure but skills have moved

- [ ] **Step 3: Update install.go**

In `apps/backend/src/cli/install.go`, change the skills source root:

```go
// Step 7: Copy embedded skills to .claude/skills/oraculo-<name>/.
// Before:
// skillsSrcRoot := filepath.Join("skills", "oraculo")
// After:
skillsSrcRoot := "skills"
entries, err := fs.ReadDir(claudekit.SkillsFS, skillsSrcRoot)
```

The rest of the code remains the same — it already iterates directories and copies to `oraculo-<name>`.

- [ ] **Step 4: Verify embed.go works**

The `embed.go` has `//go:embed skills` which embeds the entire `skills/` directory. Since we moved files INTO `skills/` (not out of it), no change is needed — the embed directive already covers the new structure.

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -v -count=1 -run TestInstall_CopiesEmbeddedOraculoSkills ./apps/backend/src/cli/`
Expected: PASS

- [ ] **Step 6: Run full test suite**

Run: `go test -v -count=1 ./apps/backend/...`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add apps/backend/src/cli/install.go apps/backend/src/cli/install_test.go
git commit -m "fix: update install to read skills from new flat structure

Skills moved from skills/oraculo/X/ to skills/X/.
Destination unchanged: .claude/skills/oraculo-X/."
```

### Task 4: Update `uninstall.go`

**Files:**
- Modify: `apps/backend/src/cli/uninstall.go:107-113`

- [ ] **Step 1: Check current uninstall logic**

The `removeSkills` function removes `.claude/skills/oraculo/`. But `install.go` copies to `.claude/skills/oraculo-epic/`, `.claude/skills/oraculo-story/`, etc. — individual directories with the `oraculo-` prefix.

The uninstall currently removes a WRONG path (`.claude/skills/oraculo/` doesn't exist — skills are at `.claude/skills/oraculo-epic/` etc.). Fix this.

- [ ] **Step 2: Write the failing test**

Add to `apps/backend/src/cli/uninstall_test.go` (or the relevant test file):

```go
func TestUninstall_RemovesAllOraculoSkills(t *testing.T) {
	setupInstallDir(t)

	// Install first
	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify skills exist
	matches, _ := filepath.Glob(filepath.Join(".claude", "skills", "oraculo-*"))
	if len(matches) == 0 {
		t.Fatal("expected oraculo-* skill dirs to exist after install")
	}

	// Uninstall
	_, err = uninstallCmd(t)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify skills removed
	matches, _ = filepath.Glob(filepath.Join(".claude", "skills", "oraculo-*"))
	if len(matches) != 0 {
		t.Errorf("expected no oraculo-* skill dirs, found: %v", matches)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test -v -count=1 -run TestUninstall_RemovesAllOraculoSkills ./apps/backend/src/cli/`
Expected: FAIL — uninstall removes `.claude/skills/oraculo/` but skills are at `.claude/skills/oraculo-*/`

- [ ] **Step 4: Fix removeSkills**

```go
func removeSkills(w interface{ Write([]byte) (int, error) }) error {
	// Remove all oraculo-* skill directories
	matches, _ := filepath.Glob(filepath.Join(".claude", "skills", "oraculo-*"))
	for _, m := range matches {
		if err := os.RemoveAll(m); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove skill %s: %w", m, err)
		}
	}
	// Also remove legacy .claude/skills/oraculo/ if it exists
	legacy := filepath.Join(".claude", "skills", "oraculo")
	if err := os.RemoveAll(legacy); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy skills: %w", err)
	}
	return nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -v -count=1 -run TestUninstall_RemovesAllOraculoSkills ./apps/backend/src/cli/`
Expected: PASS

- [ ] **Step 6: Run full test suite**

Run: `go test -v -count=1 ./apps/backend/...`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add apps/backend/src/cli/uninstall.go apps/backend/src/cli/uninstall_test.go
git commit -m "fix: uninstall removes oraculo-* skill directories

Was removing .claude/skills/oraculo/ which didn't match the
actual install destination of .claude/skills/oraculo-*/."
```

---

## Chunk 2: Plugin Manifest and MCP Config

### Task 5: Create plugin manifest

**Files:**
- Create: `claude-kit/.claude-plugin/plugin.json`

- [ ] **Step 1: Create the directory and manifest**

```bash
mkdir -p claude-kit/.claude-plugin
```

Write `claude-kit/.claude-plugin/plugin.json`:

```json
{
  "name": "oraculo",
  "version": "0.1.0",
  "description": "Socratic guide and team orchestrator for quality product development",
  "author": {
    "name": "Lucas"
  },
  "repository": "https://github.com/lucas/oraculo",
  "skills": "./skills/"
}
```

- [ ] **Step 2: Create MCP config**

Write `claude-kit/.mcp.json`:

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

- [ ] **Step 3: Commit**

```bash
git add claude-kit/.claude-plugin/ claude-kit/.mcp.json
git commit -m "feat: add Claude Code plugin manifest and MCP config

Plugin name 'oraculo' auto-namespaces skills as /oraculo:epic, etc.
MCP server starts via npx @oraculo/cli."
```

---

## Chunk 3: New `oraculo setup` Command

### Task 6: Add `setup` command

**Files:**
- Create: `apps/backend/src/cli/setup.go`
- Create: `apps/backend/src/cli/setup_test.go`
- Modify: `apps/backend/src/cli/root.go:23-35`

- [ ] **Step 1: Write the failing test**

Create `apps/backend/src/cli/setup_test.go`:

```go
package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

func setupCmd(t *testing.T, extraArgs ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(append([]string{"setup"}, extraArgs...))
	err := cmd.Execute()
	return buf.String(), err
}

func TestSetup_CreatesOraculoConfig(t *testing.T) {
	setupInstallDir(t)

	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(".oraculo", "config.json"))
	if err != nil {
		t.Fatalf("expected .oraculo/config.json: %v", err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := cfg["port"]; !ok {
		t.Error("missing 'port'")
	}
	if _, ok := cfg["preferred_language"]; !ok {
		t.Error("missing 'preferred_language'")
	}
}

func TestSetup_CreatesDatabase(t *testing.T) {
	setupInstallDir(t)

	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	dbPath := filepath.Join(".oraculo", "oraculo.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db")
	}
}

func TestSetup_CreatesHooksOnly(t *testing.T) {
	setupInstallDir(t)

	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("expected .claude/settings.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Must have hooks
	if _, ok := settings["hooks"]; !ok {
		t.Error("missing 'hooks'")
	}

	// Must NOT have mcpServers
	if _, ok := settings["mcpServers"]; ok {
		t.Error("setup should not configure mcpServers (plugin handles MCP)")
	}
}

func TestSetup_DoesNotCopySkills(t *testing.T) {
	setupInstallDir(t)

	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(".claude", "skills", "oraculo-*"))
	if len(matches) != 0 {
		t.Errorf("setup should not copy skills, found: %v", matches)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v -count=1 -run TestSetup ./apps/backend/src/cli/`
Expected: FAIL — `setup` command doesn't exist yet

- [ ] **Step 3: Implement setup.go**

Create `apps/backend/src/cli/setup.go`:

```go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Initialize Oraculo infrastructure (for plugin users)",
		Long: `Set up .oraculo/ directory, database, and hooks without copying skills
or configuring MCP. Use this when skills and MCP are provided by
the Claude Code plugin.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd)
		},
	}
	cmd.Flags().String("lang", "", "Preferred language for Oraculo sessions (BCP 47, e.g. pt-BR, en-US)")
	return cmd
}

// claudeSettingsHooksOnly is settings.json without mcpServers.
type claudeSettingsHooksOnly struct {
	Hooks map[string][]hookGroup `json:"hooks"`
}

func runSetup(cmd *cobra.Command) error {
	w := cmd.OutOrStdout()

	// Step 1: Create .oraculo/ directory.
	if err := os.MkdirAll(".oraculo", 0o755); err != nil {
		return fmt.Errorf("create .oraculo: %w", err)
	}
	fmt.Fprintln(w, "created .oraculo/")

	// Step 2: Open and close DB to trigger migration.
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	database.Close()
	fmt.Fprintln(w, "created .oraculo/oraculo.db")

	// Step 3: Find an available port.
	port, err := config.FindPort(3100, 3199)
	if err != nil {
		return fmt.Errorf("find port: %w", err)
	}

	// Step 4: Write config.
	lang, _ := cmd.Flags().GetString("lang")
	if lang == "" {
		lang = "pt-BR"
	}
	cfg := &config.Config{Port: port, PreferredLanguage: lang}
	if err := config.Write(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Fprintf(w, "created .oraculo/config.json (port %d)\n", port)

	// Step 5: Create .claude/ directory.
	if err := os.MkdirAll(".claude", 0o755); err != nil {
		return fmt.Errorf("create .claude: %w", err)
	}

	// Step 6: Write .claude/settings.json with hooks ONLY (no MCP, no skills).
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	httpGroup := func(url string) []hookGroup {
		return []hookGroup{{Hooks: []hookDef{{Type: "http", URL: url, Timeout: 5}}}}
	}
	settings := claudeSettingsHooksOnly{
		Hooks: map[string][]hookGroup{
			"SessionStart": {{
				Hooks: []hookDef{{Type: "command", Command: "oraculo hook session-start"}},
			}},
			"SubagentStart":  httpGroup(baseURL + "/hooks/agent-start"),
			"SubagentStop":   httpGroup(baseURL + "/hooks/agent-stop"),
			"TaskCompleted":  httpGroup(baseURL + "/hooks/task-completed"),
			"Stop":           httpGroup(baseURL + "/hooks/stop"),
			"TeammateIdle":   httpGroup(baseURL + "/hooks/teammate-idle"),
			"SessionEnd": {{
				Hooks: []hookDef{{Type: "command", Command: "oraculo hook session-end"}},
			}},
			"PostToolUse": {{
				Matcher: "Bash|Edit|Write|NotebookEdit",
				Hooks:   []hookDef{{Type: "http", URL: baseURL + "/hooks/tool-used", Timeout: 5}},
			}},
		},
	}
	settingsData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(filepath.Join(".claude", "settings.json"), settingsData, 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}
	fmt.Fprintln(w, "created .claude/settings.json (hooks only)")

	fmt.Fprintln(w, "Oraculo setup complete. Install the plugin for skills and MCP.")
	return nil
}
```

- [ ] **Step 4: Register in root.go**

Add `newSetupCmd()` to the `root.AddCommand(...)` call in `apps/backend/src/cli/root.go`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -v -count=1 -run TestSetup ./apps/backend/src/cli/`
Expected: All 4 tests PASS

- [ ] **Step 6: Run full test suite**

Run: `go test -v -count=1 ./apps/backend/...`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add apps/backend/src/cli/setup.go apps/backend/src/cli/setup_test.go apps/backend/src/cli/root.go
git commit -m "feat: add oraculo setup command for plugin users

Creates .oraculo/, DB, and hooks without copying skills or
configuring MCP. For users who get skills+MCP from the plugin."
```

---

## Chunk 4: npm Packages

### Task 7: Create platform package templates

**Files:**
- Create: `npm/cli-darwin-arm64/package.json`
- Create: `npm/cli-darwin-x64/package.json`
- Create: `npm/cli-linux-x64/package.json`
- Create: `npm/cli-linux-arm64/package.json`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p npm/cli-darwin-arm64 npm/cli-darwin-x64 npm/cli-linux-x64 npm/cli-linux-arm64
```

- [ ] **Step 2: Write platform package.json files**

Each file follows the same pattern. Example for `npm/cli-darwin-arm64/package.json`:

```json
{
  "name": "@oraculo/cli-darwin-arm64",
  "version": "0.0.0",
  "description": "Oraculo CLI binary for macOS ARM64",
  "os": ["darwin"],
  "cpu": ["arm64"],
  "files": ["bin/oraculo"],
  "license": "MIT",
  "preferUnplugged": true
}
```

Repeat for each platform, changing `name`, `description`, `os`, and `cpu`:
- `cli-darwin-x64`: `os: ["darwin"]`, `cpu: ["x64"]`
- `cli-linux-x64`: `os: ["linux"]`, `cpu: ["x64"]`
- `cli-linux-arm64`: `os: ["linux"]`, `cpu: ["arm64"]`

Note: `version` is `"0.0.0"` — CI replaces it with the actual version at publish time.

- [ ] **Step 3: Commit**

```bash
git add npm/
git commit -m "feat: add npm platform package templates

Templates for @oraculo/cli-{os}-{arch} packages. CI fills in
the version and binary at publish time."
```

### Task 8: Create main npm launcher package

**Files:**
- Create: `npm/cli/package.json`
- Create: `npm/cli/bin/cli.js`
- Create: `npm/cli/install.js`

- [ ] **Step 1: Write package.json**

Create `npm/cli/package.json`:

```json
{
  "name": "@oraculo/cli",
  "version": "0.0.0",
  "description": "Socratic guide and team orchestrator for quality product development",
  "bin": {
    "oraculo": "bin/cli.js"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "optionalDependencies": {
    "@oraculo/cli-darwin-arm64": "0.0.0",
    "@oraculo/cli-darwin-x64": "0.0.0",
    "@oraculo/cli-linux-x64": "0.0.0",
    "@oraculo/cli-linux-arm64": "0.0.0"
  },
  "license": "MIT",
  "files": ["bin/", "install.js"]
}
```

- [ ] **Step 2: Write the launcher script**

Create `npm/cli/bin/cli.js`:

```javascript
#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const PLATFORMS = {
  "darwin-arm64": "@oraculo/cli-darwin-arm64",
  "darwin-x64": "@oraculo/cli-darwin-x64",
  "linux-x64": "@oraculo/cli-linux-x64",
  "linux-arm64": "@oraculo/cli-linux-arm64",
};

function getBinaryPath() {
  const key = `${process.platform}-${process.arch}`;
  const pkg = PLATFORMS[key];

  if (!pkg) {
    console.error(`Unsupported platform: ${key}`);
    console.error(`Supported: ${Object.keys(PLATFORMS).join(", ")}`);
    process.exit(1);
  }

  // Try optionalDependencies first
  try {
    const binPath = require.resolve(`${pkg}/bin/oraculo`);
    return binPath;
  } catch {}

  // Fallback: postinstall may have downloaded it locally
  const localBin = path.join(__dirname, "..", "bin", "oraculo");
  if (fs.existsSync(localBin)) {
    return localBin;
  }

  console.error(`Could not find oraculo binary for ${key}.`);
  console.error("Try reinstalling: npm install @oraculo/cli");
  process.exit(1);
}

const binary = getBinaryPath();

try {
  execFileSync(binary, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  process.exit(e.status ?? 1);
}
```

- [ ] **Step 3: Write the fallback install script**

Create `npm/cli/install.js`:

```javascript
const { existsSync } = require("fs");
const path = require("path");

const PLATFORMS = {
  "darwin-arm64": "@oraculo/cli-darwin-arm64",
  "darwin-x64": "@oraculo/cli-darwin-x64",
  "linux-x64": "@oraculo/cli-linux-x64",
  "linux-arm64": "@oraculo/cli-linux-arm64",
};

const key = `${process.platform}-${process.arch}`;
const pkg = PLATFORMS[key];

if (!pkg) {
  console.warn(`[oraculo] Unsupported platform: ${key}`);
  process.exit(0);
}

// Check if optionalDependency was installed
try {
  require.resolve(`${pkg}/bin/oraculo`);
  // Binary exists via optionalDependencies — nothing to do
  process.exit(0);
} catch {
  // Not installed — inform the user
  console.warn(
    `[oraculo] Platform package ${pkg} was not installed.`
  );
  console.warn(
    "[oraculo] This can happen with --ignore-optional or strict lockfiles."
  );
  console.warn(
    `[oraculo] Install it manually: npm install ${pkg}`
  );
  process.exit(0);
}
```

- [ ] **Step 4: Make launcher executable**

```bash
chmod +x npm/cli/bin/cli.js
```

- [ ] **Step 5: Commit**

```bash
git add npm/cli/
git commit -m "feat: add @oraculo/cli npm launcher package

JS launcher resolves platform-specific binary from
optionalDependencies. Fallback warns if binary not found."
```

---

## Chunk 5: GitHub Actions CI

### Task 9: Create publish workflow

**Files:**
- Create: `.github/workflows/publish-npm.yml`

- [ ] **Step 1: Create workflows directory**

```bash
mkdir -p .github/workflows
```

- [ ] **Step 2: Write the workflow**

Create `.github/workflows/publish-npm.yml`:

```yaml
name: Publish npm packages

on:
  push:
    tags:
      - "v*"

permissions:
  contents: read

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: actions/setup-node@v4
        with:
          node-version: "20"
          registry-url: "https://registry.npmjs.org"

      # Parse version from tag (strip leading "v")
      - name: Set version
        id: version
        run: echo "VERSION=${GITHUB_REF_NAME#v}" >> "$GITHUB_OUTPUT"

      - uses: oven-sh/setup-bun@v2

      # Build dashboard assets (needed for embed)
      - name: Build dashboard
        run: |
          cd apps/dashboard && bun install && bun run build
          mkdir -p ../backend/src/server/dashboard_assets
          cp -r out/* ../backend/src/server/dashboard_assets/

      # Create bin directories for cross-compiled binaries
      - name: Create bin directories
        run: |
          for pkg in npm/cli-darwin-arm64 npm/cli-darwin-x64 npm/cli-linux-x64 npm/cli-linux-arm64; do
            mkdir -p "$pkg/bin"
          done

      # Cross-compile Go binary
      - name: Cross-compile
        env:
          VERSION: ${{ steps.version.outputs.VERSION }}
          LDFLAGS: "-s -w -X main.version=${{ steps.version.outputs.VERSION }}"
        run: |
          GOOS=darwin  GOARCH=arm64 go build -ldflags "$LDFLAGS" -o npm/cli-darwin-arm64/bin/oraculo  ./apps/backend/cmd/oraculo
          GOOS=darwin  GOARCH=amd64 go build -ldflags "$LDFLAGS" -o npm/cli-darwin-x64/bin/oraculo    ./apps/backend/cmd/oraculo
          GOOS=linux   GOARCH=amd64 go build -ldflags "$LDFLAGS" -o npm/cli-linux-x64/bin/oraculo     ./apps/backend/cmd/oraculo
          GOOS=linux   GOARCH=arm64 go build -ldflags "$LDFLAGS" -o npm/cli-linux-arm64/bin/oraculo   ./apps/backend/cmd/oraculo

      # Set version in all package.json files
      - name: Set package versions
        env:
          VERSION: ${{ steps.version.outputs.VERSION }}
        run: |
          for pkg in npm/cli-darwin-arm64 npm/cli-darwin-x64 npm/cli-linux-x64 npm/cli-linux-arm64; do
            cd "$pkg"
            npm version "$VERSION" --no-git-tag-version --allow-same-version
            cd "$GITHUB_WORKSPACE"
          done
          cd npm/cli
          npm version "$VERSION" --no-git-tag-version --allow-same-version
          # Update optionalDependencies versions
          node -e "
            const pkg = require('./package.json');
            for (const dep of Object.keys(pkg.optionalDependencies || {})) {
              pkg.optionalDependencies[dep] = '$VERSION';
            }
            require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2) + '\n');
          "

      # Publish platform packages first
      - name: Publish platform packages
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: |
          for pkg in npm/cli-darwin-arm64 npm/cli-darwin-x64 npm/cli-linux-x64 npm/cli-linux-arm64; do
            cd "$pkg"
            npm publish --access public
            cd "$GITHUB_WORKSPACE"
          done

      # Publish main package
      - name: Publish main package
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        working-directory: npm/cli
        run: npm publish --access public
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/publish-npm.yml
git commit -m "ci: add npm publish workflow for Go binary distribution

Triggered on v* tags. Cross-compiles Go for 4 platforms,
publishes @oraculo/cli-{platform} and @oraculo/cli packages."
```

---

## Chunk 6: Marketplace Repository

### Task 10: Create marketplace repo

This task is done **outside** the main Oraculo repo — it creates a separate GitHub repository.

- [ ] **Step 1: Create the marketplace repo on GitHub**

Create a new repo: `lucas/oraculo-marketplace` (or your preferred org/name).

- [ ] **Step 2: Add marketplace.json**

Create `.claude-plugin/marketplace.json` in the new repo:

```json
{
  "name": "oraculo-marketplace",
  "owner": {
    "name": "Lucas"
  },
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
      "description": "Socratic guide and team orchestrator for quality product development"
    }
  ]
}
```

- [ ] **Step 3: Push and verify**

```bash
git add .claude-plugin/marketplace.json
git commit -m "init: oraculo marketplace"
git push origin main
```

- [ ] **Step 4: Test installation**

In any project with Claude Code:

```
/plugin marketplace add lucas/oraculo-marketplace
/plugin install oraculo
```

Verify skills appear as `/oraculo:epic`, `/oraculo:story`, etc.

---

## Chunk 7: Update Makefile and Documentation

### Task 11: Add cross-compile target to Makefile

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add cross-compile target**

Add `cross-compile` to the `.PHONY` declaration and add the target to `Makefile`:

```makefile
# Cross-compile for npm distribution (local testing)
cross-compile:
	@mkdir -p npm/cli-darwin-arm64/bin npm/cli-darwin-x64/bin npm/cli-linux-x64/bin npm/cli-linux-arm64/bin
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-arm64/bin/oraculo  $(BUILD)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-x64/bin/oraculo    $(BUILD)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-x64/bin/oraculo     $(BUILD)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-arm64/bin/oraculo   $(BUILD)
```

- [ ] **Step 2: Add npm binaries to .gitignore**

Append to `.gitignore`:

```
# npm platform binaries (built by CI)
npm/cli-*/bin/
```

- [ ] **Step 3: Update the `clean` target**

```makefile
clean:
	rm -f $(BINARY)
	rm -rf apps/backend/src/server/dashboard_assets
	rm -rf npm/cli-*/bin/
```

- [ ] **Step 4: Commit**

```bash
git add Makefile .gitignore
git commit -m "build: add cross-compile target and gitignore npm binaries"
```

### Task 12: Update CLAUDE.md project structure

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update the project structure section**

Update the `claude-kit/` section in CLAUDE.md to reflect the new flat structure and add the `npm/` directory:

```markdown
### `claude-kit/`

Claude Code plugin and distributable skill kit. This directory serves dual purpose:
- **Plugin mode:** Claude Code plugin root (`.claude-plugin/plugin.json`)
- **CLI mode:** Embedded in the Go binary via `embed.FS`, copied during `oraculo install`

### `npm/`

npm package templates for distributing the Go binary. CI cross-compiles and publishes:
- `@oraculo/cli` — JS launcher with optionalDependencies
- `@oraculo/cli-{darwin,linux}-{arm64,x64}` — platform-specific Go binaries
```

Also update the skill paths in the tree diagram from `skills/oraculo/epic/` to `skills/epic/`.

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for plugin structure and npm packages"
```
