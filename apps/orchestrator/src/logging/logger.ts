import pino from "pino";

export function createLogger(name: string): pino.Logger {
  const isDev = process.env.NODE_ENV !== "production";

  if (isDev) {
    return pino({
      name,
      level: "debug",
      transport: {
        target: "pino-pretty",
        options: { colorize: true },
      },
    });
  }

  // Production with pino-roll
  try {
    return pino({
      name,
      level: "info",
      transport: {
        target: "pino-roll",
        options: {
          file: `${process.env.HOME}/.oraculo/logs/${name}`,
          frequency: "daily",
          size: "50m",
          limit: { count: 7 },
          symlink: true,
        },
      },
    });
  } catch {
    // Fallback to sync stdout
    return pino({ name, level: "info" });
  }
}
