"use strict";

const REPO = process.env.REPONERVE_REPO || "reponerve/reponerve";

function goos() {
  switch (process.platform) {
    case "darwin":
      return "darwin";
    case "linux":
      return "linux";
    case "win32":
      return "windows";
    default:
      throw new Error(
        `unsupported platform: ${process.platform} (see docs/install.md)`,
      );
  }
}

function goarch() {
  switch (process.arch) {
    case "x64":
      return "amd64";
    case "arm64":
      return "arm64";
    default:
      throw new Error(`unsupported architecture: ${process.arch}`);
  }
}

function normalizeVersion(version) {
  if (!version) {
    throw new Error("version is required");
  }
  return version.startsWith("v") ? version : `v${version}`;
}

function releaseVersion(version) {
  return normalizeVersion(version).slice(1);
}

function archiveName(version) {
  const ver = releaseVersion(version);
  const os = goos();
  const arch = goarch();
  const base = `reponerve_${ver}_${os}_${arch}`;
  if (os === "windows") {
    return `${base}.zip`;
  }
  return `${base}.tar.gz`;
}

function downloadURL(version) {
  const tag = normalizeVersion(version);
  const file = archiveName(version);
  return `https://github.com/${REPO}/releases/download/${tag}/${file}`;
}

function binaryNames() {
  if (process.platform === "win32") {
    return ["reponerve.exe", "reponerve"];
  }
  return ["reponerve"];
}

module.exports = {
  REPO,
  goos,
  goarch,
  normalizeVersion,
  releaseVersion,
  archiveName,
  downloadURL,
  binaryNames,
};
