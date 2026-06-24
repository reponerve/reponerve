"use strict";

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const { downloadURL, archiveName, binaryNames } = require("./platform");

async function downloadFile(url, dest) {
  const response = await fetch(url, { redirect: "follow" });
  if (!response.ok) {
    throw new Error(`download failed (${response.status}): ${url}`);
  }
  const buffer = Buffer.from(await response.arrayBuffer());
  fs.writeFileSync(dest, buffer);
}

function extractArchive(archivePath, destDir, name) {
  fs.mkdirSync(destDir, { recursive: true });
  if (name.endsWith(".zip")) {
    if (process.platform === "win32") {
      execFileSync(
        "powershell",
        [
          "-NoProfile",
          "-Command",
          `Expand-Archive -Path '${archivePath.replace(/'/g, "''")}' -DestinationPath '${destDir.replace(/'/g, "''")}' -Force`,
        ],
        { stdio: "inherit" },
      );
      return;
    }
    execFileSync("unzip", ["-q", "-o", archivePath, "-d", destDir], {
      stdio: "inherit",
    });
    return;
  }
  execFileSync("tar", ["-xzf", archivePath, "-C", destDir], { stdio: "inherit" });
}

function findBinary(searchDir) {
  for (const name of binaryNames()) {
    const candidate = path.join(searchDir, name);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }
  throw new Error(`binary not found in ${searchDir}`);
}

async function installBinary({ version, vendorDir }) {
  const archive = archiveName(version);
  const url = downloadURL(version);
  const tmpDir = fs.mkdtempSync(path.join(require("os").tmpdir(), "reponerve-"));
  const archivePath = path.join(tmpDir, archive);

  try {
    console.log(`reponerve: downloading ${url}`);
    await downloadFile(url, archivePath);
    extractArchive(archivePath, tmpDir, archive);
    const binary = findBinary(tmpDir);

    fs.mkdirSync(vendorDir, { recursive: true });
    const destName =
      process.platform === "win32" ? "reponerve.exe" : "reponerve";
    const dest = path.join(vendorDir, destName);
    fs.copyFileSync(binary, dest);
    if (process.platform !== "win32") {
      fs.chmodSync(dest, 0o755);
    }
    console.log(`reponerve: installed to ${dest}`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

module.exports = { installBinary };
