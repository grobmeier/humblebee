const { test } = require("node:test");
const assert = require("node:assert/strict");
const { updateWebsiteRelease } = require("./update-website-release.cjs");

function fixture() {
  const names = [
    "checksums.txt",
    "humblebee_0.7.2_darwin_amd64.tar.gz", "humblebee_0.7.2_darwin_arm64.tar.gz",
    "humblebee_0.7.2_linux_amd64.tar.gz", "humblebee_0.7.2_linux_arm64.tar.gz",
    "humblebee_0.7.2_windows_amd64.zip", "humblebee_0.7.2_windows_arm64.zip",
    "HumbleBee_GUI_v0.7.2_darwin_arm64.zip", "HumbleBee_GUI_v0.7.2_linux_amd64.tar.gz",
    "HumbleBee_GUI_v0.7.2_windows_amd64.zip", "HumbleBee_v0.7.2_sbom.cdx.json"
  ];
  const state = { latest: "v0.7.2", content: "version: v0.7.1\n", writes: [], sleeps: [],
    release: { published_at: "2026-10-16T12:00:00Z", draft: false, prerelease: false,
      assets: names.map((name) => ({ name, state: "uploaded", size: 100 })) } };
  const source = { rest: { repos: {
    getLatestRelease: async () => ({ data: { tag_name: state.latest } }),
    getReleaseByTag: async () => ({ data: state.release })
  } } };
  const website = { rest: { repos: {
    getContent: async (request) => {
      assert.equal(request.repo, "timeandbill-landing");
      assert.equal(request.ref, "develop");
      return { data: { type: "file", encoding: "base64", sha: "file-sha",
        content: Buffer.from(state.content).toString("base64") } };
    },
    createOrUpdateFileContents: async (request) => {
      state.writes.push(request);
      return { data: { commit: { sha: "website-commit" } } };
    }
  } } };
  return { state, source, website, tag: "v0.7.2", sleep: async (ms) => state.sleeps.push(ms) };
}

test("publishes only the release version to the website deploy branch", async () => {
  const f = fixture();
  assert.deepEqual(await updateWebsiteRelease(f), { updated: true, commit: "website-commit" });
  assert.equal(f.state.writes.length, 1);
  const write = f.state.writes[0];
  assert.equal(write.branch, "develop");
  assert.equal(write.path, "_data/humblebee_release.yml");
  assert.equal(write.sha, "file-sha");
  assert.equal(Buffer.from(write.content, "base64").toString(), "version: v0.7.2\n");
});

test("waits for the independently published SBOM", async () => {
  const f = fixture();
  const sbom = f.state.release.assets.pop();
  f.sleep = async (ms) => { f.state.sleeps.push(ms); f.state.release.assets.push(sbom); };
  assert.equal((await updateWebsiteRelease(f)).updated, true);
  assert.deepEqual(f.state.sleeps, [30_000]);
});

test("missing or incomplete uploads fail without changing the website", async () => {
  for (const asset of [undefined, { state: "new" }, { size: 0 }]) {
    const f = fixture();
    if (asset) Object.assign(f.state.release.assets[7], asset);
    else f.state.release.assets.splice(7, 1);
    await assert.rejects(updateWebsiteRelease({ ...f, attempts: 2 }), /assets are missing/);
    assert.equal(f.state.writes.length, 0);
  }
});

test("prerelease, malformed and obsolete tags never update the website", async () => {
  for (const tag of ["v0.7.2-rc.1", "v01.2.3", "main", "v0.7.1"]) {
    const f = fixture();
    assert.equal((await updateWebsiteRelease({ ...f, tag })).updated, false);
    assert.equal(f.state.writes.length, 0);
  }
});

test("draft and prerelease releases fail closed", async () => {
  for (const flag of ["draft", "prerelease"]) {
    const f = fixture();
    f.state.release[flag] = true;
    await assert.rejects(updateWebsiteRelease(f), /not published and stable/);
    assert.equal(f.state.writes.length, 0);
  }
});

test("reruns are idempotent and older releases cannot roll back the website", async () => {
  for (const version of ["v0.7.2", "v0.8.0", "v0.10.0"]) {
    const f = fixture();
    f.state.content = `version: ${version}\n`;
    assert.equal((await updateWebsiteRelease(f)).updated, false);
    assert.equal(f.state.writes.length, 0);
  }
});

test("unexpected website content is preserved", async () => {
  const f = fixture();
  f.state.content += "another_setting: keep\n";
  await assert.rejects(updateWebsiteRelease(f), /Unexpected website version format/);
  assert.equal(f.state.writes.length, 0);
});

test("a newer release published during the check prevents a stale write", async () => {
  const f = fixture();
  const read = f.website.rest.repos.getContent;
  f.website.rest.repos.getContent = async (request) => {
    f.state.latest = "v0.7.3";
    return read(request);
  };
  assert.equal((await updateWebsiteRelease(f)).updated, false);
  assert.equal(f.state.writes.length, 0);
});

test("file conflicts reread state and do not overwrite another release", async () => {
  const f = fixture();
  f.website.rest.repos.createOrUpdateFileContents = async () => {
    f.state.content = "version: v0.8.0\n";
    throw Object.assign(new Error("Conflict"), { status: 409 });
  };
  assert.equal((await updateWebsiteRelease(f)).updated, false);
});

test("permission errors surface instead of pretending publication succeeded", async () => {
  const f = fixture();
  f.website.rest.repos.createOrUpdateFileContents = async () => {
    throw Object.assign(new Error("Forbidden"), { status: 403 });
  };
  await assert.rejects(updateWebsiteRelease(f), /Forbidden/);
});
