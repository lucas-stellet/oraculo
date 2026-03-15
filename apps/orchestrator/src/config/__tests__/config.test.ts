import { describe, it, expect, beforeEach, afterEach } from "bun:test";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, basename } from "node:path";
import { readConfig, writeConfig, getProjectName } from "../config.ts";

describe("config", () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await mkdtemp(join(tmpdir(), "oraculo-config-test-"));
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  describe("readConfig", () => {
    it("returns defaults when config file does not exist", async () => {
      const cfg = await readConfig(tempDir);
      expect(cfg).toEqual({
        port: 0,
        name: "",
        preferred_language: "",
        skills: { design_agent: [], code_agent: [] },
      });
    });

    it("reads existing config file", async () => {
      const configDir = join(tempDir, ".oraculo");
      await Bun.write(
        join(configDir, "config.json"),
        JSON.stringify({ port: 8080, name: "my-project", preferred_language: "pt-BR" }),
      );

      const cfg = await readConfig(tempDir);
      expect(cfg.port).toBe(8080);
      expect(cfg.name).toBe("my-project");
      expect(cfg.preferred_language).toBe("pt-BR");
    });

    it("returns error for malformed JSON", async () => {
      const configDir = join(tempDir, ".oraculo");
      await Bun.write(join(configDir, "config.json"), "not json{{{");

      expect(readConfig(tempDir)).rejects.toThrow();
    });
  });

  describe("writeConfig", () => {
    it("creates .oraculo directory and writes config", async () => {
      const cfg = { port: 3000, name: "test-proj", preferred_language: "en-US", skills: { design_agent: [], code_agent: [] } };
      await writeConfig(tempDir, cfg);

      const file = Bun.file(join(tempDir, ".oraculo", "config.json"));
      const data = await file.json();
      expect(data.port).toBe(3000);
      expect(data.name).toBe("test-proj");
    });

    it("overwrites existing config", async () => {
      const cfg1 = { port: 3000, name: "v1", preferred_language: "", skills: { design_agent: [], code_agent: [] } };
      await writeConfig(tempDir, cfg1);

      const cfg2 = { port: 4000, name: "v2", preferred_language: "pt-BR", skills: { design_agent: [], code_agent: [] } };
      await writeConfig(tempDir, cfg2);

      const file = Bun.file(join(tempDir, ".oraculo", "config.json"));
      const data = await file.json();
      expect(data.port).toBe(4000);
      expect(data.name).toBe("v2");
    });

    it("writes valid JSON with trailing newline", async () => {
      const cfg = { port: 0, name: "", preferred_language: "", skills: { design_agent: [], code_agent: [] } };
      await writeConfig(tempDir, cfg);

      const file = Bun.file(join(tempDir, ".oraculo", "config.json"));
      const text = await file.text();
      expect(text.endsWith("\n")).toBe(true);
      expect(() => JSON.parse(text)).not.toThrow();
    });
  });

  describe("getProjectName", () => {
    it("returns configured name when set", async () => {
      const cfg = { port: 0, name: "custom-name", preferred_language: "", skills: { design_agent: [], code_agent: [] } };
      await writeConfig(tempDir, cfg);

      const name = await getProjectName(tempDir);
      expect(name).toBe("custom-name");
    });

    it("falls back to directory basename when name is empty", async () => {
      const name = await getProjectName(tempDir);
      expect(name).toBe(basename(tempDir));
    });
  });
});
