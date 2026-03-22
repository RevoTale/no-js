package bundler

import (
	"github.com/RevoTale/no-js/framework/i18n"
)
const defaultPublicDirName = "public"
const defaultAppRouteModule = "internal/web/app"
const defaultGenModule = "internal/web/gen"
const defaultResolverModule = "internal/web/resolvers"
// ServerBundlerConfig defines the config for bundler.
// It only defines how to bundle the app. Do not use it for app runtime in any case!
// To achieve the greatest perfomance of the target app, ServerBundlerConfig generates
// the server features on-demand, omitting the unused features and modules.
// Such architecture gives small binary size, RAM and CPU usage.
type ServerBundlerConfig struct {
	// RootDir is a root directory of the target project. Must be an absolute path. 
	// By default is the directory where command i called.
	RootDir string 
	// I18n defined the localization config
	I18n i18n.Config
	// PublicDirName is directory name for the files which shoudl
	// be served as-is without any mnodification.
	// Relative to the root. Should start with no slashes or dots.
	PublicDirName string
	// PublicDirRequestPathPrefix is the the prefix for routing where should the files from PublicDirName be served
	PublicDirRequestPathPrefix string
}

type GenerationPaths struct {
	// AppRoot is an router root directory  path
	AppRoot       string
	// GenRoot is a generated output directory
	GenRoot       string
		// GenImportRoot is a generated output module name
	GenImportRoot string
	// ResolverRoot is a resolvers root. There will be one auto generate files and data resolver should be implemented by the developer.
	ResolverRoot  string
	// AppModulePath is a project app path defined by go.mod
	AppModulePath string
}

