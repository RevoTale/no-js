package bundler

import frameworki18n "github.com/RevoTale/no-js/framework/i18n"

const (
	defaultAppDir                  = "internal/web/app"
	defaultGeneratedDir            = "internal/web/gen"
	defaultResolverDir             = "internal/web/resolvers"
	defaultPublicDirName           = "public"
	defaultPublicRequestPathPrefix = "/"
)

// Config defines build-time inputs for project layout resolution and code generation.
// It must not be used as application runtime configuration.
type Config struct {
	// I18n carries the consuming application's locale configuration for generators that need it.
	I18n frameworki18n.Config

	// AppDir is the route tree directory relative to the application root.
	// When empty, internal/web/app is used.
	AppDir string

	// GenDir is the generated route output directory relative to the application root.
	// When empty, internal/web/gen is used.
	GenDir string

	// Resolver is the handwritten resolver directory relative to the application root.
	// When empty, internal/web/resolvers is used.
	Resolver string

	// PublicDirName is the public files directory relative to the application root.
	// Files in this directory are served as-is rather than fingerprinted as bundled assets.
	// When empty, public is used.
	PublicDirName string

	// PublicDirRequestPathPrefix is the request-path prefix used when serving PublicDirName.
	// Empty values resolve to /. Non-root trailing slashes are trimmed during normalization.
	PublicDirRequestPathPrefix string
}

// ProjectLayout describes the resolved application layout on disk and in module space.
type ProjectLayout struct {
	// RootDir is the absolute filesystem path to the consuming application root.
	RootDir string

	// AppDir is the absolute filesystem path to the route tree root.
	AppDir string

	// GeneratedDir is the absolute filesystem path where generated route code is written.
	GeneratedDir string

	// GeneratedImport is the module-relative import path that corresponds to GeneratedDir.
	GeneratedImport string

	// ResolverDir is the absolute filesystem path to the handwritten resolver directory.
	ResolverDir string

	// PublicDir is the absolute filesystem path to the public files directory.
	PublicDir string

	// PublicRequestPathPrefix is the normalized request-path prefix used when serving PublicDir.
	PublicRequestPathPrefix string

	// AppModulePath is the Go module path declared in the consuming application's go.mod.
	AppModulePath string
}
