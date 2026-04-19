#!/usr/bin/env node

const { spawnSync } = require("node:child_process")
const fs = require("node:fs")
const os = require("node:os")
const path = require("node:path")

const command = process.argv[2] || "run"
const repoRoot = path.resolve(__dirname, "..")
const cypressBin = path.join(repoRoot, "node_modules", ".bin", "cypress")

const browserCandidates = [
  { name: "chrome", check: () => hasMacApp("Google Chrome.app") || hasMacApp("Google Chrome Canary.app") },
  { name: "chromium", check: () => hasMacApp("Chromium.app") },
  { name: "edge", check: () => hasMacApp("Microsoft Edge.app") },
  { name: "firefox", check: () => hasMacApp("Firefox.app") },
  { name: "webkit", check: () => hasPlaywrightWebKit() },
]

function hasMacApp(appName) {
  if (os.platform() !== "darwin") {
    return false
  }

  return fs.existsSync(path.join("/Applications", appName))
}

function hasPlaywrightWebKit() {
  try {
    require.resolve("playwright-webkit/package.json", { paths: [repoRoot] })
    return true
  } catch {
    return false
  }
}

function resolveBrowser() {
  const explicitBrowser = process.env.CYPRESS_BROWSER
  if (explicitBrowser) {
    return explicitBrowser
  }

  const matched = browserCandidates.find((candidate) => candidate.check())
  if (matched) {
    return matched.name
  }

  return os.platform() === "darwin" ? "webkit" : "electron"
}

const browser = resolveBrowser()
const args = [command, "--browser", browser, ...process.argv.slice(3)]

const result = spawnSync(cypressBin, args, {
  cwd: repoRoot,
  stdio: "inherit",
  shell: false,
})

if (result.error) {
  console.error(result.error.message)
  process.exit(1)
}

process.exit(result.status ?? 1)
