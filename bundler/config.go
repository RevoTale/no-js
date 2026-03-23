package bundler

import frameworki18n "github.com/RevoTale/no-js/framework/i18n"

const (
	defaultAppDir                  = "internal/web/app"
	defaultGeneratedDir            = "internal/web/gen"
	defaultResolverDir             = "internal/web/resolvers"
	defaultRuntimeDir              = "internal/web/runtime"
	defaultI18nDir                 = "internal/web/i18n"
	defaultPublicDirName           = "public"
	defaultPublicRequestPathPrefix = "/"
)

type FeatureMode string

const (
	FeatureAuto     FeatureMode = ""
	FeatureEnabled  FeatureMode = "enabled"
	FeatureDisabled FeatureMode = "disabled"
)

func (mode FeatureMode) EnabledByDefault(defaultValue bool) bool {
	switch mode {
	case FeatureEnabled:
		return true
	case FeatureDisabled:
		return false
	default:
		return defaultValue
	}
}

type ServerFeaturesConfig struct {
	// I18nRouting controls whether generated server bootstrap wires locale routing middleware.
	I18nRouting FeatureMode

	// StaticAssets controls whether generated server bootstrap reads and mounts the static asset manifest.
	StaticAssets FeatureMode

	// PublicFiles controls whether generated server bootstrap wires public-file serving middleware.
	PublicFiles FeatureMode
}

type ServerConfig struct {
	// RuntimeDir is the application runtime package relative to the application root.
	// When empty, internal/web/runtime is used.
	RuntimeDir string

	// I18nDir is the application i18n package relative to the application root.
	// When empty, internal/web/i18n is used.
	I18nDir string

	Features ServerFeaturesConfig
}

type ServerFeatures struct {
	I18nRouting  bool
	StaticAssets bool
	PublicFiles  bool
}

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

	Server ServerConfig
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

	// ResolverImport is the module-relative import path that corresponds to ResolverDir.
	ResolverImport string

	// RuntimeDir is the absolute filesystem path to the handwritten runtime contract package.
	RuntimeDir string

	// RuntimeImport is the module-relative import path that corresponds to RuntimeDir.
	RuntimeImport string

	// I18nDir is the absolute filesystem path to the handwritten i18n package.
	I18nDir string

	// I18nImport is the module-relative import path that corresponds to I18nDir.
	I18nImport string

	// PublicDir is the absolute filesystem path to the public files directory.
	PublicDir string

	// PublicRequestPathPrefix is the normalized request-path prefix used when serving PublicDir.
	PublicRequestPathPrefix string

	ServerFeatures ServerFeatures

	// AppModulePath is the Go module path declared in the consuming application's go.mod.
	AppModulePath string
}
