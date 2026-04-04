package projectlayout

import (
	"fmt"
	"strings"
)

const (
	defaultBundleConfigFileName   = "no-js.bundle.yaml"
	bundleConfigVersion           = 1
	defaultRoutesDir              = "web/routes"
	defaultGeneratedDir           = "web/generated"
	defaultResolversDir           = "web/resolvers"
	defaultViewDir                = "web/view"
	defaultI18nDir                = "web/i18n"
	defaultI18nMessagesDir        = "messages"
	defaultAssetsDir              = "web/assets"
	defaultAssetsBuildDir         = "web/assets-build"
	defaultStaticManifestFileName = "manifest.json"
	defaultStaticRuntimeURLPrefix = "/_assets/"
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
	// I18nRouting controls whether project layout resolves locale routing support as enabled.
	I18nRouting FeatureMode `yaml:"i18n_routing"`

	// StaticAssets controls whether project layout resolves static asset support as enabled.
	StaticAssets FeatureMode `yaml:"static_assets"`

	// HealthEndpoint controls whether project layout resolves health endpoint support as enabled.
	HealthEndpoint FeatureMode `yaml:"health_endpoint"`
}

type BuiltInI18nConfig struct {
	// Mode controls whether generation should emit the built-in i18n bundle and typed keys.
	Mode FeatureMode `yaml:"mode"`
}

type ProjectConfig struct {
	// RoutesDir is the route tree directory relative to the application root.
	RoutesDir string `yaml:"routes_dir"`

	// GeneratedDir is the generated route output directory relative to the application root.
	GeneratedDir string `yaml:"generated_dir"`

	// ResolversDir is the handwritten resolver directory relative to the application root.
	ResolversDir string `yaml:"resolvers_dir"`

	// ViewDir is the application view contract package relative to the application root.
	ViewDir string `yaml:"view_dir"`

	// I18nDir is the application i18n package relative to the application root.
	I18nDir string `yaml:"i18n_dir"`

	// AssetsDir is the source directory for bundled static assets relative to the application root.
	AssetsDir string `yaml:"assets_dir"`

	// AssetsBuildDir is the output directory for bundled static assets relative to the application root.
	AssetsBuildDir string `yaml:"assets_build_dir"`
}

type ServerConfig struct {
	Features ServerFeaturesConfig `yaml:"features"`
}

type StaticAssetsConfig struct {
	// ManifestPath is the manifest file path relative to the application root.
	ManifestPath string `yaml:"manifest_path"`
}

// Config defines build-time inputs for project layout resolution and code generation.
// It must not be used as application runtime configuration.
type Config struct {
	Version int `yaml:"version"`

	Project      ProjectConfig      `yaml:"project"`
	Server       ServerConfig       `yaml:"server"`
	I18n         BuiltInI18nConfig  `yaml:"i18n"`
	StaticAssets StaticAssetsConfig `yaml:"static_assets"`
}

type ServerFeatures struct {
	I18nRouting    bool
	StaticAssets   bool
	HealthEndpoint bool
}

type BuiltInI18nLayout struct {
	Enabled     bool
	MessagesDir string
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

	// RoutesDir is the absolute filesystem path to the route tree root.
	RoutesDir string

	// RoutesImport is the module-relative import path that corresponds to RoutesDir.
	RoutesImport string

	// GeneratedDir is the absolute filesystem path where generated route code is written.
	GeneratedDir string

	// GeneratedImport is the module-relative import path that corresponds to GeneratedDir.
	GeneratedImport string

	// ResolversDir is the absolute filesystem path to the handwritten resolver directory.
	ResolversDir string

	// ResolversImport is the module-relative import path that corresponds to ResolversDir.
	ResolversImport string

	// ViewDir is the absolute filesystem path to the handwritten view contract package.
	ViewDir string

	// ViewImport is the module-relative import path that corresponds to ViewDir.
	ViewImport string

	// I18nDir is the absolute filesystem path to the handwritten i18n package.
	I18nDir string

	// I18nImport is the module-relative import path that corresponds to I18nDir.
	I18nImport string

	BuiltInI18n    BuiltInI18nLayout
	StaticAssets   StaticAssetsLayout
	ServerFeatures ServerFeatures

	// AppModulePath is the Go module path declared in the consuming application's go.mod.
	AppModulePath string
}
