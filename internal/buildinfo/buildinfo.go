package buildinfo

// These values are injected via -ldflags at build time.
// Defaults are set for development builds.
var (
	Version   = "dev"
	Commit    = ""
	BuildTime = ""
)
