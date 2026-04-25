package metagen

import (
	"context"
	"strings"
)

type managedStylesheetsContextKey struct{}
type assetBasePathContextKey struct{}

func WithManagedStylesheets(ctx context.Context, stylesheets []string) context.Context {
	normalized := normalizeStylesheets(stylesheets)
	if len(normalized) == 0 {
		return ctx
	}

	existing := ManagedStylesheetsFromContext(ctx)
	merged := normalizeStylesheets(append(existing, normalized...))
	if len(merged) == 0 {
		return ctx
	}

	return context.WithValue(ctx, managedStylesheetsContextKey{}, merged)
}

func ManagedStylesheetsFromContext(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}

	stylesheets, ok := ctx.Value(managedStylesheetsContextKey{}).([]string)
	if !ok || len(stylesheets) == 0 {
		return nil
	}

	out := make([]string, len(stylesheets))
	copy(out, stylesheets)
	return out
}

func MergeManagedStylesheets(ctx context.Context, meta Metadata) Metadata {
	stylesheets := ManagedStylesheetsFromContext(ctx)
	if len(stylesheets) == 0 {
		return meta
	}

	return Merge(Metadata{Stylesheets: stylesheets}, meta)
}

func WithAssetBasePath(ctx context.Context, basePath string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	normalized := normalizeAssetBasePath(basePath)
	if normalized == "" {
		return ctx
	}
	return context.WithValue(ctx, assetBasePathContextKey{}, normalized)
}

func AssetBasePathFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	basePath, ok := ctx.Value(assetBasePathContextKey{}).(string)
	if !ok {
		return ""
	}
	return normalizeAssetBasePath(basePath)
}

func AssetURL(ctx context.Context, assetPath string) string {
	trimmed := strings.TrimSpace(assetPath)
	trimmed = strings.ReplaceAll(trimmed, `\`, `/`)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}

	basePath := AssetBasePathFromContext(ctx)
	if basePath == "" {
		return ""
	}
	return basePath + strings.TrimLeft(trimmed, "/")
}

func MergeManagedClientAssets(ctx context.Context, meta Metadata, assets ClientAssets) Metadata {
	resolved := Metadata{}
	for _, href := range assets.Stylesheets {
		if url := AssetURL(ctx, href); url != "" {
			resolved.Stylesheets = append(resolved.Stylesheets, url)
		}
	}
	for _, src := range assets.ModuleScripts {
		if url := AssetURL(ctx, src); url != "" {
			resolved.ModuleScripts = append(resolved.ModuleScripts, url)
		}
	}
	if len(resolved.Stylesheets) == 0 && len(resolved.ModuleScripts) == 0 {
		return meta
	}
	return Merge(meta, resolved)
}

func normalizeAssetBasePath(basePath string) string {
	trimmed := strings.TrimSpace(basePath)
	trimmed = strings.ReplaceAll(trimmed, `\`, `/`)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") &&
		!strings.HasPrefix(trimmed, "http://") &&
		!strings.HasPrefix(trimmed, "https://") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/") + "/"
}
