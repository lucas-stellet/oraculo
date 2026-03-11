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

try {
  require.resolve(`${pkg}/bin/oraculo`);
  process.exit(0);
} catch {
  console.warn(`[oraculo] Platform package ${pkg} was not installed.`);
  console.warn("[oraculo] This can happen with --ignore-optional or strict lockfiles.");
  console.warn(`[oraculo] Install it manually: npm install ${pkg}`);
  process.exit(0);
}
