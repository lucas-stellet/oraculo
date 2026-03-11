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
    return require.resolve(`${pkg}/bin/oraculo`);
  } catch {}

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
