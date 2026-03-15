export function cleanEnv(): Record<string, string> {
  const env: Record<string, string> = {};
  for (const [key, value] of Object.entries(process.env)) {
    if (key !== "CLAUDECODE" && value !== undefined) {
      env[key] = value;
    }
  }
  return env;
}
