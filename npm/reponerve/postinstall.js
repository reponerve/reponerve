"use strict";

const fs = require("fs");
const path = require("path");
const { installBinary } = require("./lib/install");

async function main() {
  if (process.env.REPONERVE_SKIP_POSTINSTALL === "1") {
    console.log("reponerve: skipping postinstall (REPONERVE_SKIP_POSTINSTALL=1)");
    return;
  }

  const root = __dirname;
  const vendorDir = path.join(root, "vendor");
  const pkg = require(path.join(root, "package.json"));
  const version = process.env.REPONERVE_VERSION || pkg.reponerve?.version || pkg.version;

  const existing = fs.existsSync(vendorDir) && fs.readdirSync(vendorDir).length > 0;
  if (existing && process.env.REPONERVE_FORCE_POSTINSTALL !== "1") {
    console.log("reponerve: vendor binary present, skipping download");
    return;
  }

  await installBinary({ version, vendorDir });
}

main().catch((err) => {
  console.error(`reponerve postinstall failed: ${err.message}`);
  process.exit(1);
});
