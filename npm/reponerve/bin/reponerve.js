#!/usr/bin/env node
"use strict";

const path = require("path");
const { spawnSync } = require("child_process");
const fs = require("fs");
const { binaryNames } = require("../lib/platform");

const root = path.join(__dirname, "..");
const vendorDir = path.join(root, "vendor");

function resolveBinary() {
  for (const name of binaryNames()) {
    const candidate = path.join(vendorDir, name);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }
  console.error(
    "reponerve binary not found. Re-run: npm install reponerve (or npm rebuild reponerve)",
  );
  process.exit(1);
}

const binary = resolveBinary();
const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  env: process.env,
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
