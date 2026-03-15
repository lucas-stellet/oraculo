import { describe, test, expect } from "bun:test";
import { createProgram } from "../root.ts";

describe("createProgram", () => {
  test("returns a Commander program named oraculo", () => {
    const program = createProgram();
    expect(program.name()).toBe("oraculo");
  });

  test("has correct version", () => {
    const program = createProgram();
    expect(program.version()).toBe("0.1.0");
  });

  test("registers all expected commands", () => {
    const program = createProgram();
    const names = program.commands.map((c) => c.name());
    expect(names).toContain("start");
    expect(names).toContain("kill");
    expect(names).toContain("restart");
    expect(names).toContain("status");
    expect(names).toContain("logs");
    expect(names).toContain("version");
    expect(names).toContain("install");
    expect(names).toContain("setup");
    expect(names).toContain("uninstall");
  });

  test("start command has http and mcp subcommands", () => {
    const program = createProgram();
    const start = program.commands.find((c) => c.name() === "start")!;
    const subNames = start.commands.map((c) => c.name());
    expect(subNames).toContain("http");
    expect(subNames).toContain("mcp");
  });

  test("install command has --global, --local, and --lang options", () => {
    const program = createProgram();
    const install = program.commands.find((c) => c.name() === "install")!;
    const optNames = install.options.map((o) => o.long);
    expect(optNames).toContain("--global");
    expect(optNames).toContain("--local");
    expect(optNames).toContain("--lang");
  });

  test("uninstall command has --purge option", () => {
    const program = createProgram();
    const uninstall = program.commands.find((c) => c.name() === "uninstall")!;
    const optNames = uninstall.options.map((o) => o.long);
    expect(optNames).toContain("--purge");
  });

  test("setup command has --lang option", () => {
    const program = createProgram();
    const setup = program.commands.find((c) => c.name() === "setup")!;
    const optNames = setup.options.map((o) => o.long);
    expect(optNames).toContain("--lang");
  });
});
