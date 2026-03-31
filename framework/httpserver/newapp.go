package httpserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/framework"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/staticassets"
	"github.com/a-h/templ"
)

const defaultStaticManifestPath = "web/assets-build/manifest.json"
const defaultPublicDir = "web/public"

type AppBundle[C interface{}] struct {
	Context                       C
	Handlers                      []framework.RouteHandler[C]
	I18n                          *frameworki18n.Config
	NotFoundPage                  func(notFoundContext framework.NotFoundContext) templ.Component
	OnStaticAssetBasePathResolved func(prefix string)
}

type StaticAssetsConfig struct {
	ManifestPath string
	URLPrefix    string
}

type CustomConfig struct {
	ExtraRoutes         func(*http.ServeMux) error
	MainMiddlewares     []func(http.Handler) http.Handler
	CachePolicies       CachePolicies
	StaticAssets        *StaticAssetsConfig
	PublicFiles         *PublicFilesConfig
	LogServerError      func(err error)
	LogResolverTiming   func(event framework.ResolverTiming)
	EnableResolverDebug bool
	DisableHealth       bool
	HealthPath          string
	HealthBody          string
}

func NewApp[C interface{}](cfg Config[C]) (http.Handler, error) {
	app := cfg.App
	custom := cfg.Custom

	staticAssetsCfg := custom.StaticAssets
	if staticAssetsCfg == nil {
		staticAssetsCfg = &StaticAssetsConfig{
			ManifestPath: defaultStaticManifestPath,
			URLPrefix:    defaultStaticPrefix,
		}
	}

	staticMount, err := loadStaticMount(staticAssetsCfg.ManifestPath, staticAssetsCfg.URLPrefix)
	if err != nil {
		return nil, fmt.Errorf("resolve static assets: %w", err)
	}
	if app.OnStaticAssetBasePathResolved != nil {
		app.OnStaticAssetBasePathResolved(staticBasePath(staticMount, staticAssetsCfg.URLPrefix))
	}

	publicFiles, err := resolvePublicFiles(custom.PublicFiles)
	if err != nil {
		return nil, fmt.Errorf("resolve public files: %w", err)
	}

	return New(Config[C]{
		AppContext:          app.Context,
		Handlers:            app.Handlers,
		I18n:                app.I18n,
		PublicFiles:         publicFiles,
		MountExtraRoutes:    custom.ExtraRoutes,
		MainMiddlewares:     custom.MainMiddlewares,
		Static:              staticMount,
		CachePolicies:       custom.CachePolicies,
		NotFoundPage:        app.NotFoundPage,
		LogServerError:      custom.LogServerError,
		LogResolverTiming:   custom.LogResolverTiming,
		EnableResolverDebug: custom.EnableResolverDebug,
		DisableHealth:       custom.DisableHealth,
		HealthPath:          custom.HealthPath,
		HealthBody:          custom.HealthBody,
	})
}

func resolvePublicFiles(publicFiles *PublicFilesConfig) (*PublicFilesConfig, error) {
	if publicFiles != nil {
		return publicFiles, validatePublicFilesConfig(*publicFiles)
	}

	if !pathIsDirectory(defaultPublicDir) {
		return nil, nil
	}

	cfg := PublicFilesConfig{Dir: defaultPublicDir}
	return &cfg, nil
}

func validatePublicFilesConfig(cfg PublicFilesConfig) error {
	trimmedDir := strings.TrimSpace(cfg.Dir)
	if trimmedDir == "" {
		return nil
	}
	if !pathIsDirectory(trimmedDir) {
		return fmt.Errorf("public dir %q is not a directory", trimmedDir)
	}
	return nil
}

func loadStaticMount(manifestPath string, basePrefix string) (StaticMount, error) {
	trimmedManifestPath := strings.TrimSpace(manifestPath)
	if trimmedManifestPath == "" || !pathIsFile(trimmedManifestPath) {
		return StaticMount{}, nil
	}

	manifest, err := staticassets.ReadManifest(trimmedManifestPath)
	if err != nil {
		return StaticMount{}, fmt.Errorf("load static manifest %q: %w", trimmedManifestPath, err)
	}

	staticDir := filepath.Clean(filepath.Dir(trimmedManifestPath))
	info, statErr := os.Stat(staticDir)
	if statErr != nil {
		return StaticMount{}, fmt.Errorf("stat static build dir %q: %w", staticDir, statErr)
	}
	if !info.IsDir() {
		return StaticMount{}, fmt.Errorf("static build dir %q is not a directory", staticDir)
	}

	versionedPrefix := manifest.VersionedURLPrefix(basePrefix)
	return StaticMount{URLPrefix: versionedPrefix, Dir: staticDir}, nil
}

func staticBasePath(mount StaticMount, basePrefix string) string {
	if strings.TrimSpace(mount.URLPrefix) != "" {
		return mount.URLPrefix
	}
	return basePrefix
}

func pathIsDirectory(path string) bool {
	info, err := os.Stat(filepath.Clean(path))
	return err == nil && info.IsDir()
}

func pathIsFile(path string) bool {
	info, err := os.Stat(filepath.Clean(path))
	return err == nil && !info.IsDir()
}
