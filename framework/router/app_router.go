package router

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

const pageTemplateName = "page.templ"

var (
	dynamicSegmentNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	slotSegmentNamePattern    = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
	reservedSegmentPattern    = regexp.MustCompile(`^_([a-z_]+)__([a-zA-Z][a-zA-Z0-9_]*)$`)
)

const (
	reservedGroupKind            = "group"
	reservedSlotKind             = "slot"
	reservedParamKind            = "param"
	reservedCatchAllKind         = "catchall"
	reservedOptionalCatchAllKind = "optional_catchall"
)

type SegmentKind string

const (
	SegmentStatic           SegmentKind = "static"
	SegmentDynamic          SegmentKind = "dynamic"
	SegmentCatchAll         SegmentKind = "catch_all"
	SegmentOptionalCatchAll SegmentKind = "optional_catch_all"
	SegmentGroup            SegmentKind = "group"
	SegmentSlot             SegmentKind = "slot"
)

type Segment struct {
	Kind SegmentKind
	Name string
}

func (segment Segment) RawPart() string {
	switch segment.Kind {
	case SegmentDynamic:
		return "_param__" + segment.Name
	case SegmentCatchAll:
		return "_catchall__" + segment.Name
	case SegmentOptionalCatchAll:
		return "_optional_catchall__" + segment.Name
	case SegmentGroup:
		return "_group__" + segment.Name
	case SegmentSlot:
		return "_slot__" + segment.Name
	default:
		return segment.Name
	}
}

func (segment Segment) ContributesToPublicPath() bool {
	switch segment.Kind {
	case SegmentGroup, SegmentSlot:
		return false
	default:
		return true
	}
}

func (segment Segment) IsParam() bool {
	switch segment.Kind {
	case SegmentDynamic, SegmentCatchAll, SegmentOptionalCatchAll:
		return true
	default:
		return false
	}
}

func (segment Segment) PublicPart() string {
	if !segment.ContributesToPublicPath() {
		return ""
	}
	return segment.RawPart()
}

func (segment Segment) PatternKeyPart() string {
	switch segment.Kind {
	case SegmentStatic:
		return segment.Name
	case SegmentDynamic:
		return ":"
	case SegmentCatchAll:
		return "*"
	case SegmentOptionalCatchAll:
		return "**"
	default:
		return ""
	}
}

type ParamValues map[string][]string

type AppRouteMatch struct {
	ID     string
	Params ParamValues
}

func (match AppRouteMatch) Param(name string) ([]string, bool) {
	if match.Params == nil {
		return nil, false
	}

	value, ok := match.Params[name]
	return slices.Clone(value), ok
}

type appRoute struct {
	id            string
	internalParts []Segment
	publicParts   []Segment
	patternKey    string
}

type AppRouter struct {
	routes []appRoute
}

func NewAppRouter(embedded fs.FS, root string) (*AppRouter, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("app router root cannot be empty")
	}

	routes := make([]appRoute, 0, 8)
	seenPattern := make(map[string]string)

	walkErr := fs.WalkDir(embedded, root, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if path.Base(filePath) != pageTemplateName {
			return nil
		}

		relPath := strings.TrimPrefix(filePath, root+"/")
		if relPath == filePath {
			return fmt.Errorf("compute route path for %q under root %q", filePath, root)
		}

		route, parseErr := parseAppRoute(relPath)
		if parseErr != nil {
			return parseErr
		}
		if len(route.publicParts) == 0 && containsSlotSegment(route.internalParts) {
			return nil
		}
		if containsSlotSegment(route.internalParts) {
			return nil
		}

		if existing, ok := seenPattern[route.patternKey]; ok {
			return fmt.Errorf("route pattern conflict: %q and %q", existing, route.id)
		}
		seenPattern[route.patternKey] = route.id
		routes = append(routes, route)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk app directory: %w", walkErr)
	}

	if len(routes) == 0 {
		return nil, errors.New("no page.templ routes found")
	}

	return &AppRouter{routes: routes}, nil
}

func ParseDirectorySegments(routeDir string) ([]Segment, error) {
	if strings.TrimSpace(routeDir) == "" {
		return []Segment{}, nil
	}

	parts := strings.Split(routeDir, "/")
	segments := make([]Segment, 0, len(parts))
	for _, part := range parts {
		segment, err := ParseDirectorySegment(part)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

func ParseDirectorySegment(part string) (Segment, error) {
	trimmed := strings.TrimSpace(part)
	if trimmed == "" {
		return Segment{}, errors.New("route segment cannot be empty")
	}

	switch {
	case strings.HasPrefix(trimmed, "_"):
		return parseReservedDirectorySegment(trimmed)
	}

	if strings.ContainsAny(trimmed, "[]()@") {
		return Segment{}, fmt.Errorf("invalid static segment %q", part)
	}

	return Segment{Kind: SegmentStatic, Name: trimmed}, nil
}

func parseReservedDirectorySegment(part string) (Segment, error) {
	matches := reservedSegmentPattern.FindStringSubmatch(part)
	if len(matches) != 3 {
		return Segment{}, fmt.Errorf(
			"invalid reserved route segment %q; use _group__, _slot__, _param__, _catchall__, or _optional_catchall__",
			part,
		)
	}

	kind := matches[1]
	name := matches[2]
	if !dynamicSegmentNamePattern.MatchString(name) {
		return Segment{}, fmt.Errorf("invalid reserved route name %q", name)
	}

	switch kind {
	case reservedGroupKind:
		return Segment{Kind: SegmentGroup, Name: name}, nil
	case reservedSlotKind:
		if !slotSegmentNamePattern.MatchString(name) {
			return Segment{}, fmt.Errorf("invalid slot name %q", part)
		}
		return Segment{Kind: SegmentSlot, Name: name}, nil
	case reservedParamKind:
		return Segment{Kind: SegmentDynamic, Name: name}, nil
	case reservedCatchAllKind:
		return Segment{Kind: SegmentCatchAll, Name: name}, nil
	case reservedOptionalCatchAllKind:
		return Segment{Kind: SegmentOptionalCatchAll, Name: name}, nil
	default:
		return Segment{}, fmt.Errorf(
			"unknown reserved route segment %q; use _group__, _slot__, _param__, _catchall__, or _optional_catchall__",
			part,
		)
	}
}

func InternalRouteID(segments []Segment) string {
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		parts = append(parts, segment.RawPart())
	}
	return strings.Join(parts, "/")
}

func PatternKey(segments []Segment) string {
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		key := segment.PatternKeyPart()
		if key == "" {
			continue
		}
		parts = append(parts, key)
	}
	if len(parts) == 0 {
		return "/"
	}
	return "/" + strings.Join(parts, "/")
}

func MatchPathPattern(pattern string, requestPath string) (ParamValues, bool) {
	segments, err := ParsePatternSegments(pattern)
	if err != nil {
		return nil, false
	}
	return MatchSegments(segments, requestPath)
}

func ParsePatternSegments(pattern string) ([]Segment, error) {
	trimmed := strings.Trim(strings.TrimSpace(pattern), "/")
	if trimmed == "" {
		return []Segment{}, nil
	}

	parts := strings.Split(trimmed, "/")
	segments := make([]Segment, 0, len(parts))
	for _, part := range parts {
		segment, err := ParseDirectorySegment(part)
		if err != nil {
			return nil, err
		}
		if !segment.ContributesToPublicPath() {
			return nil, fmt.Errorf("pattern segment %q does not contribute to the public path", part)
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

func MatchSegments(pattern []Segment, requestPath string) (ParamValues, bool) {
	requestSegments := splitPathSegments(requestPath)
	params, matched := matchSegments(pattern, requestSegments)
	if !matched {
		return nil, false
	}
	return params, true
}

func (router *AppRouter) Match(requestPath string) (AppRouteMatch, bool) {
	requestSegments := splitPathSegments(requestPath)
	matched := make([]appRoute, 0, 2)
	valuesByRoute := make(map[string]ParamValues)

	for _, route := range router.routes {
		params, ok := matchSegments(route.publicParts, requestSegments)
		if !ok {
			continue
		}
		matched = append(matched, route)
		valuesByRoute[route.id] = params
	}
	if len(matched) == 0 {
		return AppRouteMatch{}, false
	}

	best := matched[0]
	for _, candidate := range matched[1:] {
		if compareRouteSpecificity(candidate.publicParts, best.publicParts) > 0 {
			best = candidate
		}
	}

	params := valuesByRoute[best.id]
	if len(params) == 0 {
		return AppRouteMatch{ID: best.id}, true
	}
	return AppRouteMatch{ID: best.id, Params: params}, true
}

func parseAppRoute(relPath string) (appRoute, error) {
	cleaned := path.Clean(strings.TrimSpace(relPath))
	if cleaned == "" || cleaned == "." {
		return appRoute{}, errors.New("route path cannot be empty")
	}

	routeDir := ""
	if cleaned != pageTemplateName {
		suffix := "/" + pageTemplateName
		if !strings.HasSuffix(cleaned, suffix) {
			return appRoute{}, fmt.Errorf("route file %q must end with %q", relPath, suffix)
		}
		routeDir = strings.TrimSuffix(cleaned, suffix)
	}

	internalParts, err := ParseDirectorySegments(routeDir)
	if err != nil {
		return appRoute{}, fmt.Errorf("route file %q: %w", relPath, err)
	}
	publicParts := PublicSegments(internalParts)

	return appRoute{
		id:            InternalRouteID(internalParts),
		internalParts: internalParts,
		publicParts:   publicParts,
		patternKey:    PatternKey(internalParts),
	}, nil
}

func PublicSegments(segments []Segment) []Segment {
	out := make([]Segment, 0, len(segments))
	for _, segment := range segments {
		if !segment.ContributesToPublicPath() {
			continue
		}
		out = append(out, segment)
	}
	return out
}

func matchSegments(pattern []Segment, requestSegments []string) (ParamValues, bool) {
	params := make(ParamValues)
	patternIdx := 0
	requestIdx := 0

	for patternIdx < len(pattern) {
		segment := pattern[patternIdx]
		switch segment.Kind {
		case SegmentStatic:
			if requestIdx >= len(requestSegments) || requestSegments[requestIdx] != segment.Name {
				return nil, false
			}
			patternIdx++
			requestIdx++
		case SegmentDynamic:
			if requestIdx >= len(requestSegments) {
				return nil, false
			}
			params[segment.Name] = []string{requestSegments[requestIdx]}
			patternIdx++
			requestIdx++
		case SegmentCatchAll:
			if requestIdx >= len(requestSegments) {
				return nil, false
			}
			params[segment.Name] = slices.Clone(requestSegments[requestIdx:])
			patternIdx = len(pattern)
			requestIdx = len(requestSegments)
		case SegmentOptionalCatchAll:
			if requestIdx >= len(requestSegments) {
				params[segment.Name] = nil
			} else {
				params[segment.Name] = slices.Clone(requestSegments[requestIdx:])
				requestIdx = len(requestSegments)
			}
			patternIdx = len(pattern)
		default:
			return nil, false
		}
	}

	if requestIdx != len(requestSegments) {
		return nil, false
	}

	if len(params) == 0 {
		return nil, true
	}
	return params, true
}

func compareRouteSpecificity(left []Segment, right []Segment) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for idx := 0; idx < limit; idx++ {
		leftWeight := segmentSpecificity(left[idx])
		rightWeight := segmentSpecificity(right[idx])
		if leftWeight == rightWeight {
			continue
		}
		if leftWeight > rightWeight {
			return 1
		}
		return -1
	}

	switch {
	case len(left) == len(right):
		return 0
	case len(left) < len(right):
		if remainingOnlyOptionalCatchAll(right[limit:]) {
			return 1
		}
		return -1
	default:
		if remainingOnlyOptionalCatchAll(left[limit:]) {
			return -1
		}
		return 1
	}
}

func segmentSpecificity(segment Segment) int {
	switch segment.Kind {
	case SegmentStatic:
		return 4
	case SegmentDynamic:
		return 3
	case SegmentCatchAll:
		return 2
	case SegmentOptionalCatchAll:
		return 1
	default:
		return 0
	}
}

func remainingOnlyOptionalCatchAll(segments []Segment) bool {
	if len(segments) == 0 {
		return false
	}
	for _, segment := range segments {
		if segment.Kind != SegmentOptionalCatchAll {
			return false
		}
	}
	return true
}

func containsSlotSegment(segments []Segment) bool {
	for _, segment := range segments {
		if segment.Kind == SegmentSlot {
			return true
		}
	}
	return false
}

func splitPathSegments(raw string) []string {
	cleaned := path.Clean("/" + strings.TrimSpace(raw))
	if cleaned == "/" {
		return []string{}
	}

	trimmed := strings.Trim(cleaned, "/")
	if trimmed == "" {
		return []string{}
	}

	return strings.Split(trimmed, "/")
}
