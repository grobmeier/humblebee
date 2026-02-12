package cli

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

var buildInfo = BuildInfo{
	Version: "dev",
	Commit:  "none",
	Date:    "unknown",
}

func SetBuildInfo(info BuildInfo) {
	if info.Version != "" {
		buildInfo.Version = info.Version
	}
	if info.Commit != "" {
		buildInfo.Commit = info.Commit
	}
	if info.Date != "" {
		buildInfo.Date = info.Date
	}
}

