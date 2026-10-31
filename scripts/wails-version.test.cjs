const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");

const root = path.resolve(__dirname, "..");
const read = (file) => fs.readFileSync(path.join(root, file), "utf8");
const version = read("go.mod").match(/github\.com\/wailsapp\/wails\/v2\s+(v\d+\.\d+\.\d+)/)[1];

for (const file of [".github/workflows/test.yml", ".github/workflows/release-gui.yml", "GUI.md", "CONTRIBUTE.md"]) {
  test(`${file} uses the Wails module version for CLI commands`, () => {
    const commands = [...read(file).matchAll(/github\.com\/wailsapp\/wails\/v2\/cmd\/wails@(v\d+\.\d+\.\d+)/g)];
    assert.ok(commands.length > 0, `No versioned Wails CLI command found in ${file}`);
    for (const command of commands) assert.equal(command[1], version);
  });
}

test("macOS packaging uses and documents the Wails module version by default", () => {
  const script = read("scripts/release-macos-app.sh");
  assert.equal(script.match(/WAILS_VERSION\s+Wails CLI version\. Default: (v\d+\.\d+\.\d+)/)[1], version);
  assert.equal(script.match(/wails_version="\$\{WAILS_VERSION:-(v\d+\.\d+\.\d+)\}"/)[1], version);
});
