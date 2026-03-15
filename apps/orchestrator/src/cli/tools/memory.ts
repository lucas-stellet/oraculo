import { Command } from "commander";
import { MemoryStore } from "../../db/stores/memory-store.ts";
import { CATEGORIES, CONFIDENCES } from "../../domain/enums.ts";
import type { Category, Confidence } from "../../domain/enums.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function memoryCommands(): Command {
  const cmd = new Command("memory").description(
    "Manage codebase knowledge memory",
  );

  cmd.addCommand(memoryStoreCmd());
  cmd.addCommand(memorySearchCmd());
  cmd.addCommand(memoryDomainsCmd());

  return cmd;
}

function memoryStoreCmd(): Command {
  return new Command("store")
    .description("Store a knowledge finding")
    .requiredOption("--domain <domain>", "Knowledge domain")
    .requiredOption(
      "--category <category>",
      "Finding category: pattern|convention|constraint|dependency|test|architecture",
    )
    .requiredOption("--finding <finding>", "The knowledge finding text")
    .option("--source <source>", "Source file(s) related to the finding", "")
    .option(
      "--confidence <confidence>",
      "Confidence level: high|medium|low",
      "medium",
    )
    .action((opts: any, cmd: Command) => {
      try {
        const category = opts.category as string;
        if (!CATEGORIES.includes(category as Category)) {
          writeError(
            new Error(
              `invalid category "${category}": must be one of ${CATEGORIES.join(", ")}`,
            ),
          );
          return;
        }

        const confidence = opts.confidence as string;
        if (!CONFIDENCES.includes(confidence as Confidence)) {
          writeError(
            new Error(
              `invalid confidence "${confidence}": must be one of ${CONFIDENCES.join(", ")}`,
            ),
          );
          return;
        }

        const db = dbFromCmd(cmd);
        const store = new MemoryStore(db);
        const knowledge = store.Store(
          opts.domain,
          category as Category,
          opts.finding,
          opts.source,
          confidence as Confidence,
        );
        writeJSON(knowledge);
      } catch (err) {
        writeError(err);
      }
    });
}

function memorySearchCmd(): Command {
  return new Command("search")
    .description("Search knowledge findings")
    .argument("<query>", "Search query")
    .option("--domain <domain>", "Filter results by domain")
    .action((query: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new MemoryStore(db);
        const results = store.Search(query, opts.domain);
        writeJSON(results);
      } catch (err) {
        writeError(err);
      }
    });
}

function memoryDomainsCmd(): Command {
  return new Command("domains")
    .description("List all knowledge domains")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new MemoryStore(db);
        writeJSON(store.Domains());
      } catch (err) {
        writeError(err);
      }
    });
}
