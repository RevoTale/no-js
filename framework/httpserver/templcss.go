package httpserver

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/RevoTale/no-js/framework/metagen"
	frameworkstaticassets "github.com/RevoTale/no-js/framework/staticassets"
	"github.com/a-h/templ"
)

const (
	defaultTemplCSSAssetPath = "styles/templ.css"
	privateTemplCSSPath      = "/__no_js_internal__/templ.css"
)

type TemplCSSConfig struct {
	Manifest  frameworkstaticassets.Manifest
	AssetPath string
	Classes   []templ.CSSClass
}

func applyTemplCSS(next http.Handler, staticPrefix string, cfg *TemplCSSConfig) (http.Handler, error) {
	if cfg == nil {
		return next, nil
	}

	stylesheetURL, err := templCSSStylesheetURL(staticPrefix, *cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve templ css stylesheet URL: %w", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := metagen.WithManagedStylesheets(r.Context(), []string{stylesheetURL})
		next.ServeHTTP(w, r.WithContext(ctx))
	})

	if len(cfg.Classes) == 0 {
		return handler, nil
	}

	cssMiddleware := templ.NewCSSMiddleware(handler, cfg.Classes...)
	cssMiddleware.Path = privateTemplCSSPath
	return cssMiddleware, nil
}

func templCSSStylesheetURL(staticPrefix string, cfg TemplCSSConfig) (string, error) {
	if strings.TrimSpace(cfg.Manifest.Hash) == "" {
		return "", fmt.Errorf("manifest hash is required")
	}

	assetPath := normalizeTemplCSSAssetPath(cfg.AssetPath)
	prefix := strings.TrimSpace(staticPrefix)
	if prefix == "" {
		prefix = cfg.Manifest.VersionedURLPrefix(defaultStaticPrefix)
	} else {
		prefix = frameworkstaticassets.NormalizeURLPrefix(prefix)
	}

	return path.Join(prefix, assetPath), nil
}

func normalizeTemplCSSAssetPath(assetPath string) string {
	trimmed := strings.TrimSpace(assetPath)
	trimmed = strings.ReplaceAll(trimmed, `\`, `/`)
	trimmed = strings.TrimPrefix(trimmed, "/")
	if trimmed == "" {
		return defaultTemplCSSAssetPath
	}
	return trimmed
}
