package bundler

import frameworki18n "github.com/RevoTale/no-js/framework/i18n"

const (
	defaultAppDir                  = "internal/web/app"
	defaultGeneratedDir            = "internal/web/gen"
	defaultResolverDir             = "internal/web/resolvers"
	defaultPublicDirName           = "public"
	defaultPublicRequestPathPrefix = "/"
)

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
