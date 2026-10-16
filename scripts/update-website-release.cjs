const stableVersion = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
const sourceRepo = { owner: "grobmeier", repo: "humblebee" };
const targetFile = {
  owner: "grobmeier", repo: "timeandbill-landing",
  path: "_data/humblebee_release.yml", ref: "develop"
};

function requiredAssets(tag) {
  return [
    "checksums.txt",
    ...["darwin", "linux", "windows"].flatMap((os) =>
      ["amd64", "arm64"].map((arch) =>
        `humblebee_${tag.slice(1)}_${os}_${arch}.${os === "windows" ? "zip" : "tar.gz"}`)),
    `HumbleBee_GUI_${tag}_darwin_arm64.zip`,
    `HumbleBee_GUI_${tag}_linux_amd64.tar.gz`,
    `HumbleBee_GUI_${tag}_windows_amd64.zip`,
    `HumbleBee_${tag}_sbom.cdx.json`
  ];
}

function compareVersions(left, right) {
  const a = left.slice(1).split(".").map(BigInt);
  const b = right.slice(1).split(".").map(BigInt);
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) return a[i] > b[i] ? 1 : -1;
  }
  return 0;
}

async function updateWebsiteRelease({ source, website, tag, attempts = 20,
  sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms)), log = () => {} }) {
  if (!stableVersion.test(tag)) return { updated: false, reason: "Not a stable release tag" };

  for (let attempt = 0; attempt < attempts; attempt++) {
    const { data: latest } = await source.rest.repos.getLatestRelease(sourceRepo);
    if (latest.tag_name !== tag || latest.draft || latest.prerelease) {
      return { updated: false, reason: "Release is not the latest stable release" };
    }
    const { data: release } = await source.rest.repos.getReleaseByTag({ ...sourceRepo, tag });
    if (release.draft || release.prerelease || !release.published_at) {
      throw new Error("Release is not published and stable");
    }
    const available = new Set(release.assets.filter((asset) => asset.state === "uploaded" && asset.size > 0).map((asset) => asset.name));
    const missing = requiredAssets(tag).filter((name) => !available.has(name));
    if (missing.length) {
      if (attempt === attempts - 1) throw new Error(`Release assets are missing: ${missing.join(", ")}`);
      log(`Waiting for release assets: ${missing.join(", ")}`);
      await sleep(30_000);
      continue;
    }

    for (let retry = 0; retry < 3; retry++) {
      const { data: file } = await website.rest.repos.getContent(targetFile);
      if (file.type !== "file" || file.encoding !== "base64") throw new Error("Unexpected website version file");
      const content = Buffer.from(file.content, "base64").toString("utf8");
      const match = /^version: (v\d+\.\d+\.\d+)\r?\n?$/.exec(content);
      if (!match || !stableVersion.test(match[1])) throw new Error("Unexpected website version format");
      if (compareVersions(match[1], tag) >= 0) {
        return { updated: false, reason: "Website already uses this or a newer release" };
      }
      const { data: currentLatest } = await source.rest.repos.getLatestRelease(sourceRepo);
      if (currentLatest.tag_name !== tag) return { updated: false, reason: "A newer release was published" };
      try {
        const { data: result } = await website.rest.repos.createOrUpdateFileContents({
          owner: targetFile.owner, repo: targetFile.repo, path: targetFile.path,
          branch: targetFile.ref, sha: file.sha,
          message: `fix: point HumbleBee downloads to ${tag}`,
          content: Buffer.from(`version: ${tag}\n`).toString("base64")
        });
        return { updated: true, commit: result.commit.sha };
      } catch (error) {
        if (error.status !== 409 || retry === 2) throw error;
      }
    }
  }
  throw new Error("No release asset checks were performed");
}

module.exports = { requiredAssets, updateWebsiteRelease };
