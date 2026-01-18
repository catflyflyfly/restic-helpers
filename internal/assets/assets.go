// Package assets provides embedded files for initialization.
// These files are compiled into the binary so they can be written
// to the user's config directory on first run.
package assets

import (
	_ "embed"
)

// Base config files

//go:embed example/core.exclude.txt
var CoreExclude string

//go:embed example/config.toml
var DefaultConfig string

//go:embed example/secret.toml
var DefaultSecret string

// Per-repository config files

//go:embed example/repo/name.txt
var RepoName string

//go:embed example/repo/password.txt
var RepoPassword string

//go:embed example/repo/paths.txt
var RepoPaths string

//go:embed example/repo/exclude.txt
var RepoExclude string

//go:embed example/repo/healthcheck.txt
var RepoHealthcheck string
