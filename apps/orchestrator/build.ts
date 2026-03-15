const targets = [
  "bun-darwin-arm64",
  "bun-darwin-x64",
  "bun-linux-x64",
  "bun-linux-arm64",
] as const;

for (const target of targets) {
  console.log(`Building for ${target}...`);
  const proc = Bun.spawn([
    "bun", "build", "--compile",
    `--target=${target}`,
    "--outfile", `./dist/${target}/oraculo`,
    "./src/index.ts",
  ], { stdout: "inherit", stderr: "inherit" });
  const exitCode = await proc.exited;
  if (exitCode !== 0) {
    console.error(`Build failed for ${target}`);
    process.exit(1);
  }
}

console.log("All builds complete.");
