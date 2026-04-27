package approutegen

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	frameworkrouter "github.com/RevoTale/no-js/framework/router"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func buildRouteMetas(pages []templateDef, paths projectlayout.ProjectLayout) ([]routeMeta, error) {
	_ = paths
	metas := make([]routeMeta, 0, len(pages))

	for _, page := range pages {
		pageViewType, err := parsePageViewType(page.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("route %q: %w", page.RouteID, err)
		}

		params, err := routeParamsFromSegments(page.PublicRouteID, page.PublicSegments)
		if err != nil {
			return nil, err
		}

		routeName := routeNameFromSegments(page.Segments)
		meta := routeMeta{
			RouteID:            page.RouteID,
			InternalRouteID:    page.InternalRouteID,
			PublicRouteID:      page.PublicRouteID,
			Segments:           slices.Clone(page.Segments),
			PublicSegments:     slices.Clone(page.PublicSegments),
			RelativeSegments:   slices.Clone(page.RelativeSegments),
			Namespace:          page.Namespace,
			SlotName:           page.SlotName,
			SlotOwnerRouteID:   page.SlotOwnerRouteID,
			SlotRootInternalID: page.SlotRootInternalID,
			RouteName:          routeName,
			ParamsTypeName:     routeName + "Params",
			Params:             params,
			Page:               page,
			PageViewType:       pageViewType,
		}

		metas = append(metas, meta)
	}

	sort.Slice(metas, func(i int, j int) bool {
		if metas[i].PublicRouteID != metas[j].PublicRouteID {
			return metas[i].PublicRouteID < metas[j].PublicRouteID
		}
		return metas[i].RouteID < metas[j].RouteID
	})

	return metas, nil
}

func parsePageViewType(pageTemplatePath string) (string, error) {
	source, err := os.ReadFile(pageTemplatePath)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", filepath.ToSlash(pageTemplatePath), err)
	}

	matches := pageViewTypePattern.FindStringSubmatch(string(source))
	if len(matches) < 2 {
		return "", fmt.Errorf("%q must declare templ Page(view <type>)", filepath.ToSlash(pageTemplatePath))
	}

	viewType := strings.TrimSpace(matches[1])
	if viewType == "" {
		return "", fmt.Errorf("%q has empty Page view type", filepath.ToSlash(pageTemplatePath))
	}
	if !strings.Contains(viewType, ".") {
		return "", fmt.Errorf("%q page view type %q must be package qualified", filepath.ToSlash(pageTemplatePath), viewType)
	}
	if !strings.HasPrefix(viewType, viewPackageName+".") {
		return "", fmt.Errorf(
			"%q page view type %q must be %s-qualified",
			filepath.ToSlash(pageTemplatePath),
			viewType,
			viewPackageName,
		)
	}
	return viewType, nil
}

func inspectRouteTemplateModels(routes *routeFiles) error {
	for routeID, layout := range routes.Layouts {
		modelType, err := parseLayoutTemplateModelType(layout, routes.LayoutSlots[layout.RouteID])
		if err != nil {
			return err
		}
		layout.ModelType = modelType
		routes.Layouts[routeID] = layout
		routes.Templates = updateTemplateModelType(routes.Templates, layout)
	}
	for routeID, layout := range routes.SlotLayouts {
		modelType, err := parseLayoutTemplateModelType(layout, nil)
		if err != nil {
			return err
		}
		layout.ModelType = modelType
		routes.SlotLayouts[routeID] = layout
		routes.Templates = updateTemplateModelType(routes.Templates, layout)
	}
	for routeID, fallback := range routes.Defaults {
		modelType, err := parseDefaultTemplateModelType(fallback.SourcePath)
		if err != nil {
			return err
		}
		fallback.ModelType = modelType
		routes.Defaults[routeID] = fallback
		routes.Templates = updateTemplateModelType(routes.Templates, fallback)
	}
	for routeID, notFound := range routes.NotFounds {
		modelType, err := parseNotFoundTemplateModelType(notFound.SourcePath)
		if err != nil {
			return err
		}
		notFound.ModelType = modelType
		routes.NotFounds[routeID] = notFound
		routes.Templates = updateTemplateModelType(routes.Templates, notFound)
	}
	return nil
}

func updateTemplateModelType(templates []templateDef, updated templateDef) []templateDef {
	for idx, tpl := range templates {
		if tpl.Kind == updated.Kind && tpl.RouteID == updated.RouteID && tpl.Namespace == updated.Namespace {
			templates[idx].ModelType = updated.ModelType
		}
	}
	return templates
}

func parseLayoutTemplateModelType(layout templateDef, slotNames []string) (string, error) {
	layoutTemplatePath := layout.SourcePath
	source, err := os.ReadFile(layoutTemplatePath)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", filepath.ToSlash(layoutTemplatePath), err)
	}

	matches := layoutSignaturePattern.FindStringSubmatch(string(source))
	if len(matches) < 2 {
		return "", fmt.Errorf(
			"%q must declare %s",
			filepath.ToSlash(layoutTemplatePath),
			expectedLayoutSignature(layout, slotNames),
		)
	}
	params := splitSignatureParams(matches[1])
	expected := expectedLayoutParams(layout, slotNames)
	if len(params) != len(expected) {
		return "", fmt.Errorf(
			"%q must declare %s",
			filepath.ToSlash(layoutTemplatePath),
			expectedLayoutSignature(layout, slotNames),
		)
	}
	for idx := range expected {
		if !layoutParamMatches(params[idx], expected[idx]) {
			return "", fmt.Errorf(
				"%q must declare %s",
				filepath.ToSlash(layoutTemplatePath),
				expectedLayoutSignature(layout, slotNames),
			)
		}
	}

	modelParamIndex := 0
	if layout.RouteID == "" {
		modelParamIndex = 1
	}
	_, modelType, ok := splitNamedParam(params[modelParamIndex])
	if !ok {
		return "", fmt.Errorf(
			"%q must declare %s",
			filepath.ToSlash(layoutTemplatePath),
			expectedLayoutSignature(layout, slotNames),
		)
	}
	if err := validateViewQualifiedType(layoutTemplatePath, "layout model", modelType); err != nil {
		return "", err
	}
	return modelType, nil
}

func expectedLayoutParams(layout templateDef, slotNames []string) []string {
	params := make([]string, 0, len(slotNames)+3)
	if layout.RouteID == "" {
		params = append(params, "meta metagen.Metadata")
	}
	params = append(params, "model view.<LayoutView>", "child templ.Component")
	for _, slotName := range slotNames {
		params = append(params, slotName+" templ.Component")
	}
	return params
}

func expectedLayoutSignature(layout templateDef, slotNames []string) string {
	return "templ Layout(" + strings.Join(expectedLayoutParams(layout, slotNames), ", ") + ")"
}

func layoutParamMatches(actual string, expected string) bool {
	if expected == "model view.<LayoutView>" {
		name, paramType, ok := splitNamedParam(actual)
		return ok && token.IsIdentifier(name) && strings.HasPrefix(paramType, viewPackageName+".")
	}
	return actual == expected
}

func splitNamedParam(param string) (string, string, bool) {
	parts := strings.Fields(strings.TrimSpace(param))
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func splitSignatureParams(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	params := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		params = append(params, trimmed)
	}
	return params
}

func validateRootTemplateSignature(rootTemplatePath string) error {
	source, err := os.ReadFile(rootTemplatePath)
	if err != nil {
		return fmt.Errorf("read %q: %w", filepath.ToSlash(rootTemplatePath), err)
	}

	if len(rootTemplateSignaturePattern.FindStringSubmatch(string(source))) < 1 {
		return fmt.Errorf(
			"%q must declare templ RootLayout(meta metagen.Metadata, locale string, child templ.Component)",
			filepath.ToSlash(rootTemplatePath),
		)
	}

	return nil
}

func parseDefaultTemplateModelType(defaultTemplatePath string) (string, error) {
	source, err := os.ReadFile(defaultTemplatePath)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", filepath.ToSlash(defaultTemplatePath), err)
	}

	matches := defaultSignaturePattern.FindStringSubmatch(string(source))
	if len(matches) < 2 {
		return "", fmt.Errorf(
			"%q must declare templ Default(model view.<DefaultView>)",
			filepath.ToSlash(defaultTemplatePath),
		)
	}

	viewType := strings.TrimSpace(matches[1])
	if err := validateViewQualifiedType(defaultTemplatePath, "default view", viewType); err != nil {
		return "", err
	}

	return viewType, nil
}

func parseNotFoundTemplateModelType(notFoundTemplatePath string) (string, error) {
	source, err := os.ReadFile(notFoundTemplatePath)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", filepath.ToSlash(notFoundTemplatePath), err)
	}

	matches := notFoundSignaturePattern.FindStringSubmatch(string(source))
	if len(matches) < 2 {
		return "", fmt.Errorf(
			"%q must declare templ NotFound(model view.<NotFoundView>, path string)",
			filepath.ToSlash(notFoundTemplatePath),
		)
	}

	viewType := strings.TrimSpace(matches[1])
	if err := validateViewQualifiedType(notFoundTemplatePath, "404 view", viewType); err != nil {
		return "", err
	}

	return viewType, nil
}

func validateViewQualifiedType(templatePath string, label string, viewType string) error {
	if strings.TrimSpace(viewType) == "" {
		return fmt.Errorf("%q has empty %s type", filepath.ToSlash(templatePath), label)
	}
	if !strings.Contains(viewType, ".") {
		return fmt.Errorf("%q %s type %q must be package qualified", filepath.ToSlash(templatePath), label, viewType)
	}
	if !strings.HasPrefix(viewType, viewPackageName+".") {
		return fmt.Errorf(
			"%q %s type %q must be %s-qualified",
			filepath.ToSlash(templatePath),
			label,
			viewType,
			viewPackageName,
		)
	}
	return nil
}

func validateNoDocumentTags(templatePath string) error {
	source, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read %q: %w", filepath.ToSlash(templatePath), err)
	}
	content := string(source)
	checks := []struct {
		label   string
		pattern *regexp.Regexp
	}{
		{label: "<html>", pattern: htmlTagPattern},
		{label: "<head>", pattern: headTagPattern},
		{label: "<body>", pattern: bodyTagPattern},
	}
	for _, check := range checks {
		if check.pattern.MatchString(content) {
			return fmt.Errorf(
				"%q contains %q; only root.templ may define document-level tags",
				filepath.ToSlash(templatePath),
				check.label,
			)
		}
	}
	return nil
}

func routeNameFromSegments(segments []routeSegment) string {
	if len(segments) == 0 {
		return "Root"
	}

	builder := strings.Builder{}
	for _, segment := range segments {
		switch segment.Kind {
		case frameworkrouter.SegmentDynamic:
			builder.WriteString("Param")
		case frameworkrouter.SegmentCatchAll:
			builder.WriteString("CatchAll")
		case frameworkrouter.SegmentOptionalCatchAll:
			builder.WriteString("OptionalCatchAll")
		case frameworkrouter.SegmentGroup:
			builder.WriteString("Group")
		case frameworkrouter.SegmentSlot:
			builder.WriteString("Slot")
		}
		builder.WriteString(pascalToken(segment.Name))
	}

	name := builder.String()
	if name == "" {
		return "Root"
	}
	return name
}

func routeParamsFromSegments(routeID string, segments []routeSegment) ([]routeParamDef, error) {
	params := make([]routeParamDef, 0, len(segments))
	seen := make(map[string]struct{})

	for _, segment := range segments {
		if !segment.IsParam() {
			continue
		}

		fieldName := pascalToken(segment.Name)
		if fieldName == "" {
			return nil, fmt.Errorf("route %q has invalid param name %q", routeID, segment.Name)
		}
		if _, ok := seen[fieldName]; ok {
			return nil, fmt.Errorf("route %q has duplicate param field %q", routeID, fieldName)
		}
		seen[fieldName] = struct{}{}

		params = append(params, routeParamDef{
			Name:      segment.Name,
			FieldName: fieldName,
			Type:      paramTypeForSegment(segment),
		})
	}

	return params, nil
}

func paramTypeForSegment(segment routeSegment) string {
	switch segment.Kind {
	case frameworkrouter.SegmentCatchAll, frameworkrouter.SegmentOptionalCatchAll:
		return "[]string"
	default:
		return "string"
	}
}
