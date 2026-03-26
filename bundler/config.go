package bundler

import (
	"fmt"
	"strings"
)

const (
	defaultBundleConfigFileName    = "no-js.bundle.yaml"
	bundleConfigVersion            = 1
	defaultAppDir                  = "internal/web/app"
	defaultGeneratedDir            = "internal/web/gen"
	defaultResolverDir             = "internal/web/resolvers"
	defaultRuntimeDir              = "internal/web/runtime"
	defaultBootstrapDir            = "internal/web/bootstrap"
	defaultI18nDir                 = "internal/web/i18n"
	defaultPublicDirName           = "public"
	defaultPublicRequestPathPrefix = "/"
	defaultStaticSourceDir         = "internal/web/static"
	defaultStaticOutDir            = "internal/web/static-build"
	defaultStaticManifestFileName  = "manifest.json"
	defaultStaticRuntimeURLPrefix  = "/_assets/"
)

type FeatureMode string

const (
	FeatureAuto     FeatureMode = "auto"
	FeatureEnabled  FeatureMode = "enabled"
	FeatureDisabled FeatureMode = "disabled"
)

func (mode FeatureMode) Resolve(autoValue bool) bool {
	switch mode {
	case "", FeatureAuto:
		return autoValue
	case FeatureEnabled:
		return true
	case FeatureDisabled:
		return false
	default:
		return autoValue
	}
}

func (mode *FeatureMode) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}

	normalized := FeatureMode(strings.TrimSpace(strings.ToLower(raw)))
	switch normalized {
	case "", FeatureAuto:
		*mode = FeatureAuto
	case FeatureEnabled, FeatureDisabled:
		*mode = normalized
	default:
		return fmt.Errorf("invalid feature mode %q: expected auto, enabled, or disabled", raw)
	}

	return nil
}

type ServerFeaturesConfig struct {
	// I18nRouting controls whether generated server bootstrap wires locale routing middleware.
	I18nRouting FeatureMode `yaml:"i18n_routing"`

	// StaticAssets controls whether generated server bootstrap wires static asset manifest loading and mounting.
	StaticAssets FeatureMode `yaml:"static_assets"`

	// PublicFiles controls whether generated server bootstrap wires public-file serving middleware.
	PublicFiles FeatureMode `yaml:"public_files"`

	// HealthEndpoint controls whether generated server bootstrap exposes a health endpoint.
	HealthEndpoint FeatureMode `yaml:"health_endpoint"`
}

type ProjectConfig struct {
	// AppDir is the route tree directory relative to the application root.
	AppDir string `yaml:"app_dir"`

	// GenDir is the generated route output directory relative to the application root.
	GenDir string `yaml:"gen_dir"`

	// ResolverDir is the handwritten resolver directory relative to the application root.
	ResolverDir string `yaml:"resolver_dir"`

	// RuntimeDir is the application runtime contract package relative to the application root.
	RuntimeDir string `yaml:"runtime_dir"`

	// BootstrapDir is the app-owned bootstrap package relative to the application root.
	BootstrapDir string `yaml:"bootstrap_dir"`

	// I18nDir is the application i18n package relative to the application root.
	I18nDir string `yaml:"i18n_dir"`

	// PublicDir is the public files directory relative to the application root.
	PublicDir string `yaml:"public_dir"`
}

type StaticAssetsConfig struct {
	// SourceDir is the source directory for bundled static assets relative to the application root.
	SourceDir string `yaml:"source_dir"`

	// OutDir is the output directory for bundled static assets relative to the application root.
	OutDir string `yaml:"out_dir"`

	// ManifestPath is the manifest file path relative to the application root.
	ManifestPath string `yaml:"manifest_path"`
}

type PublicFilesConfig struct {
	// RequestPathPrefix is the request-path prefix used when serving public files.
	RequestPathPrefix string `yaml:"request_path_prefix"`
}

type ServerConfig struct {
	Features ServerFeaturesConfig `yaml:"features"`
}

// Config defines build-time inputs for project layout resolution and code generation.
// It must not be used as application runtime configuration.
type Config struct {
	Version int `yaml:"version"`

	Project      ProjectConfig      `yaml:"project"`
	Server       ServerConfig       `yaml:"server"`
	StaticAssets StaticAssetsConfig `yaml:"static_assets"`
	PublicFiles  PublicFilesConfig  `yaml:"public_files"`
}

type ServerFeatures struct {
	I18nRouting    bool
	StaticAssets   bool
	PublicFiles    bool
	HealthEndpoint bool
}

type StaticAssetsLayout struct {
	SourceDir    string
	OutDir       string
	ManifestPath string
}

// ProjectLayout describes the resolved application layout on disk and in module space.
type ProjectLayout struct {
	// RootDir is the absolute filesystem path to the consuming application root.
	RootDir string

	// ConfigPath is the absolute filesystem path to the discovered bundle config file.
	// It is empty when defaults are used without a bundle config file.
	ConfigPath string

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

	// BootstrapDir is the absolute filesystem path to the app-owned bootstrap package.
	BootstrapDir string

	// BootstrapImport is the module-relative import path that corresponds to BootstrapDir.
	BootstrapImport string

	// I18nDir is the absolute filesystem path to the handwritten i18n package.
	I18nDir string

	// I18nImport is the module-relative import path that corresponds to I18nDir.
	I18nImport string

	// PublicDir is the absolute filesystem path to the public files directory.
	PublicDir string

	// PublicRequestPathPrefix is the normalized request-path prefix used when serving PublicDir.
	PublicRequestPathPrefix string

	StaticAssets   StaticAssetsLayout
	ServerFeatures ServerFeatures

	// AppModulePath is the Go module path declared in the consuming application's go.mod.
	AppModulePath string
}
