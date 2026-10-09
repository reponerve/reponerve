"use strict";

const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { execFileSync } = require("child_process");
const { downloadURL, checksumsURL, archiveName, binaryNames } = require("./platform");

async function downloadFile(url, dest) {
  const response = await fetch(url, { redirect: "follow" });
  if (!response.ok) {
    throw new Error(`download failed (${response.status}): ${url}`);
  }
  const buffer = Buffer.from(await response.arrayBuffer());
  fs.writeFileSync(dest, buffer);
}

async function verifyChecksum(archivePath, archive, version) {
  const fileBuffer = fs.readFileSync(archivePath);
  const hash = crypto.createHash("sha256").update(fileBuffer).digest("hex");
  const chkUrl = checksumsURL(version);
  try {
    const res = await fetch(chkUrl, { redirect: "follow" });
    if (!res.ok) {
      console.warn(`reponerve: warning: could not fetch checksums from ${chkUrl} (${res.status})`);
      return;
    }
    const text = await res.text();
    const lines = text.split(/\r?\n/);
    let expectedHash = null;
    for (const line of lines) {
      const parts = line.trim().split(/\s+/);
      if (parts.length >= 2 && (parts[1] === archive || parts[1] === `*${archive}`)) {
        expectedHash = parts[0];
        break;
      }
    }
    if (expectedHash && expectedHash.toLowerCase() !== hash.toLowerCase()) {
      throw new Error(`checksum mismatch for ${archive}: expected ${expectedHash}, got ${hash}`);
    }
    if (expectedHash) {
      console.log(`reponerve: checksum verified for ${archive}`);
    }
  } catch (err) {
    if (err.message.includes("checksum mismatch")) {
      throw err;
    }
    console.warn(`reponerve: warning: checksum verification skipped: ${err.message}`);
  }
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
    await verifyChecksum(archivePath, archive, version);
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
