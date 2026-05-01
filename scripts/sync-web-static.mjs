#!/usr/bin/env node

import { cpSync, existsSync, mkdirSync, rmSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const sourceDir = path.join(repoRoot, "web", "out");
const targetDir = path.join(repoRoot, "static", "out");
const readmePath = path.join(targetDir, "README.md");
const readmeContent = "# Octopus Frontend Build Output";

if (!existsSync(sourceDir) || !statSync(sourceDir).isDirectory()) {
  console.error(`Missing frontend export directory: ${sourceDir}`);
  process.exit(1);
}

mkdirSync(path.join(repoRoot, "static"), { recursive: true });
rmSync(targetDir, { recursive: true, force: true });
cpSync(sourceDir, targetDir, { recursive: true });

writeFileSync(readmePath, readmeContent, "utf8");

console.log(`Synced frontend export from ${sourceDir} to ${targetDir}`);
