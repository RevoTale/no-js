
// Config defines build-time framework inputs.
// It must not be used as application runtime configuration.
type Config struct {
	I18n frameworki18n.Config

	AppDir   string
	GenDir   string
	Resolver string

	PublicDirName              string
	PublicDirRequestPathPrefix string
}

// ProjectLayout describes the resolved application layout on disk and in module space.
type ProjectLayout struct {
	RootDir string

	AppDir          string
	GeneratedDir    string
	GeneratedImport string
	ResolverDir     string

	PublicDir               string
	PublicRequestPathPrefix string

	AppModulePath string
}
// Config defines build-time framework inputs.
// It must not be used as application runtime configuration.
type Config struct {
	// I18n carries the runtime-owned locale configuration used during build-time generation.
	I18n frameworki18n.Config

	// AppDir is the route tree directory relative to the app root.
	// When empty, internal/web/app is used.
	AppDir string

	// GenDir is the generated route output directory relative to the app root.
	// When empty, internal/web/gen is used.
	GenDir string

	// Resolver is the handwritten resolver directory relative to the app root.
	// When empty, internal/web/resolvers is used.
	Resolver string

	// PublicDirName is the public files directory relative to the app root.
	// When empty, public is used.
	PublicDirName string

	// PublicDirRequestPathPrefix is the URL prefix used when serving public files.
	// When empty, / is used.
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

	// GeneratedImport is the module-relative import path for GeneratedDir.
	GeneratedImport string

	// ResolverDir is the absolute filesystem path to the handwritten resolver directory.
	ResolverDir string

	// PublicDir is the absolute filesystem path to the public files directory.
	PublicDir string

	// PublicRequestPathPrefix is the normalized request-path prefix for PublicDir.
	PublicRequestPathPrefix string

	// AppModulePath is the Go module path declared in the consuming application's go.mod.
	AppModulePath string
}