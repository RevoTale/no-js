package clientassets

import (
	"fmt"
	"strings"

	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func buildRouteBundle(
	layout projectlayout.ProjectLayout,
	index *assetIndex,
	route RouteSpec,
) (*routeBundle, []*cssBundle, error) {
	collector := newDependencyCollector(layout, index)
	for _, templatePath := range route.TemplatePaths {
		collector.addFile(templatePath)
	}

	cssSpecs := route.CSSBundles
	if len(cssSpecs) == 0 {
		cssSpecs = fallbackCSSBundleSpecs(route)
	}

	cssBundles := make([]*cssBundle, 0, len(cssSpecs))
	cssBundlePaths := make([]string, 0, len(cssSpecs))
	seenBundlePaths := map[string]struct{}{}
	for _, spec := range cssSpecs {
		assetPath, err := cssBundleAssetPath(layout, spec)
		if err != nil {
			return nil, nil, err
		}
		if _, ok := seenBundlePaths[assetPath]; !ok {
			seenBundlePaths[assetPath] = struct{}{}
			cssBundlePaths = append(cssBundlePaths, assetPath)
		}
		cssBundles = append(cssBundles, &cssBundle{
			AssetPath: assetPath,
			CSSFiles:  collectCSSBundleAssets(layout, index, spec.TemplatePaths),
		})
	}

	return &routeBundle{
		RouteID:        route.RouteID,
		CSSBundlePaths: cssBundlePaths,
		ScriptFiles:    sortedScriptAssets(collector.scripts),
	}, cssBundles, nil
}

func bundleClientAssetsByRoute(
	bundles map[string]*routeBundle,
	cssBundlesByPath map[string]*cssBundle,
) map[string]metagen.ClientAssets {
	assetsByRoute := make(map[string]metagen.ClientAssets, len(bundles))
	for routeID, bundle := range bundles {
		assetsByRoute[routeID] = bundleClientAssets(bundle, cssBundlesByPath)
	}
	return assetsByRoute
}

func bundleClientAssets(bundle *routeBundle, cssBundlesByPath map[string]*cssBundle) metagen.ClientAssets {
	assets := metagen.ClientAssets{}
	for _, assetPath := range bundle.CSSBundlePaths {
		cssBundle := cssBundlesByPath[assetPath]
		if cssBundle == nil || len(cssBundle.CSSFiles) == 0 {
			continue
		}
		assets.Stylesheets = append(assets.Stylesheets, assetPath)
	}
	for _, file := range bundle.ScriptFiles {
		assets.ModuleScripts = append(assets.ModuleScripts, file.AssetPath)
	}
	return assets
}

func fallbackCSSBundleSpecs(route RouteSpec) []CSSBundleSpec {
	specs := make([]CSSBundleSpec, 0, len(route.TemplatePaths))
	for _, templatePath := range route.TemplatePaths {
		if strings.TrimSpace(templatePath) == "" {
			continue
		}
		specs = append(specs, CSSBundleSpec{
			OwnerTemplatePath: templatePath,
			TemplatePaths:     []string{templatePath},
		})
	}
	return specs
}

func cssBundleAssetPath(layout projectlayout.ProjectLayout, spec CSSBundleSpec) (string, error) {
	ownerPath := strings.TrimSpace(spec.OwnerTemplatePath)
	if ownerPath == "" {
		return "", fmt.Errorf("css bundle owner template path is required")
	}
	return logicalCSSAssetPath(layout, ownerPath)
}

func collectCSSBundleAssets(
	layout projectlayout.ProjectLayout,
	index *assetIndex,
	templatePaths []string,
) []*cssAsset {
	files := []*cssAsset{}
	seen := map[*cssAsset]struct{}{}
	for _, templatePath := range templatePaths {
		collector := newDependencyCollector(layout, index)
		collector.addFile(templatePath)
		for _, css := range nonEmptyCSSAssets(sortedCSSAssets(collector.css)) {
			if _, ok := seen[css]; ok {
				continue
			}
			seen[css] = struct{}{}
			files = append(files, css)
		}
	}
	return files
}

func mergeCSSBundles(
	bundlesByPath map[string]*cssBundle,
	bundleOrder *[]string,
	bundles []*cssBundle,
) {
	for _, bundle := range bundles {
		if bundle == nil || strings.TrimSpace(bundle.AssetPath) == "" {
			continue
		}
		existing := bundlesByPath[bundle.AssetPath]
		if existing == nil {
			existing = &cssBundle{AssetPath: bundle.AssetPath}
			bundlesByPath[bundle.AssetPath] = existing
			*bundleOrder = append(*bundleOrder, bundle.AssetPath)
		}
		appendUniqueCSSAssets(&existing.CSSFiles, bundle.CSSFiles)
	}
}

func appendUniqueCSSAssets(target *[]*cssAsset, files []*cssAsset) {
	seen := map[*cssAsset]struct{}{}
	for _, file := range *target {
		seen[file] = struct{}{}
	}
	for _, file := range files {
		if file == nil {
			continue
		}
		if _, ok := seen[file]; ok {
			continue
		}
		seen[file] = struct{}{}
		*target = append(*target, file)
	}
}
