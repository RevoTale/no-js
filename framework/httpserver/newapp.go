package httpserver

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/RevoTale/no-js/framework"
	frameworkdiscovery "github.com/RevoTale/no-js/framework/discovery"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/staticassets"
	"github.com/a-h/templ"
)

const defaultStaticManifestPath = "web/assets-build/manifest.json"
const defaultPublicDir = "web/public"

type AppBundle[C any] struct {
	Context       C
	ExactHandlers []framework.RouteHandler[C]
	Handlers      []framework.RouteHandler[C]
	Discovery     *frameworkdiscovery.Bundle[C]
	I18n          *frameworki18n.Config
	ResolveRoot   func(r *http.Request) *url.URL
	NotFoundPage  func(
		appCtx C,
		r *http.Request,
		notFoundContext framework.NotFoundContext,
	) (templ.Component, error)
	TemplCSSClasses               func() []templ.CSSClass
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
	ServerErrorPage     func(err error) templ.Component
	LogServerError      func(err error)
	LogServerErrorEvent func(event ServerErrorEvent)
	LogResolverTiming   func(event framework.ResolverTiming)
	EnableResolverDebug bool
	DisableHealth       bool
	HealthPath          string
	HealthBody          string
}

func NewApp[C any](cfg Config[C]) (http.Handler, error) {
	app := cfg.App
	custom := cfg.Custom

	if appContextIsNil(app.Context) {
		return nil, fmt.Errorf("app context is required")
	}

	staticAssetsCfg := custom.StaticAssets
	if staticAssetsCfg == nil {
		staticAssetsCfg = &StaticAssetsConfig{
			ManifestPath: defaultStaticManifestPath,
			URLPrefix:    defaultStaticPrefix,
		}
	}

	staticAssets, err := loadStaticAssets(staticAssetsCfg.ManifestPath, staticAssetsCfg.URLPrefix)
	if err != nil {
		return nil, fmt.Errorf("resolve static assets: %w", err)
	}
	staticMount := staticAssets.Mount
	if app.OnStaticAssetBasePathResolved != nil {
		app.OnStaticAssetBasePathResolved(staticBasePath(staticMount, staticAssetsCfg.URLPrefix))
	}

	publicFiles, err := resolvePublicFiles(custom.PublicFiles)
	if err != nil {
		return nil, fmt.Errorf("resolve public files: %w", err)
	}

	var templCSSCfg *TemplCSSConfig
	if app.TemplCSSClasses != nil && strings.TrimSpace(staticAssets.Dir) != "" {
		classes := app.TemplCSSClasses()
		stylesheetFile := filepath.Join(staticAssets.Dir, filepath.FromSlash(defaultTemplCSSAssetPath))
		if len(classes) > 0 && pathIsFile(stylesheetFile) {
			templCSSCfg = &TemplCSSConfig{
				Manifest:  staticAssets.Manifest,
				AssetPath: defaultTemplCSSAssetPath,
				Classes:   classes,
			}
		}
	}

	return New(Config[C]{
		AppContext:          app.Context,
		ExactHandlers:       app.ExactHandlers,
		Handlers:            app.Handlers,
		Discovery:           app.Discovery,
		I18n:                app.I18n,
		ResolveRoot:         app.ResolveRoot,
		PublicFiles:         publicFiles,
		MountExtraRoutes:    custom.ExtraRoutes,
		MainMiddlewares:     custom.MainMiddlewares,
		Static:              staticMount,
		TemplCSS:            templCSSCfg,
		CachePolicies:       custom.CachePolicies,
		NotFoundPage:        app.NotFoundPage,
		ServerErrorPage:     custom.ServerErrorPage,
		LogServerError:      custom.LogServerError,
		LogServerErrorEvent: custom.LogServerErrorEvent,
		LogResolverTiming:   custom.LogResolverTiming,
		EnableResolverDebug: custom.EnableResolverDebug,
		DisableHealth:       custom.DisableHealth,
		HealthPath:          custom.HealthPath,
		HealthBody:          custom.HealthBody,
	})
}

func appContextIsNil[C any](value C) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
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

type resolvedStaticAssets struct {
	Mount    StaticMount
	Manifest staticassets.Manifest
	Dir      string
}

func loadStaticAssets(manifestPath string, basePrefix string) (resolvedStaticAssets, error) {
	trimmedManifestPath := strings.TrimSpace(manifestPath)
	if trimmedManifestPath == "" || !pathIsFile(trimmedManifestPath) {
		return resolvedStaticAssets{}, nil
	}

	manifest, err := staticassets.ReadManifest(trimmedManifestPath)
	if err != nil {
		return resolvedStaticAssets{}, fmt.Errorf("load static manifest %q: %w", trimmedManifestPath, err)
	}

	staticDir := filepath.Clean(filepath.Dir(trimmedManifestPath))
	info, statErr := os.Stat(staticDir)
	if statErr != nil {
		return resolvedStaticAssets{}, fmt.Errorf("stat static build dir %q: %w", staticDir, statErr)
	}
	if !info.IsDir() {
		return resolvedStaticAssets{}, fmt.Errorf("static build dir %q is not a directory", staticDir)
	}

	versionedPrefix := manifest.VersionedURLPrefix(basePrefix)
	return resolvedStaticAssets{
		Mount: StaticMount{
			URLPrefix: versionedPrefix,
			Dir:       staticDir,
		},
		Manifest: manifest,
		Dir:      staticDir,
	}, nil
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
