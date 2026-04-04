package i18n

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"slices"
	"sort"
	"strings"
)

type LocaleLink struct {
	Code   string
	Label  string
	Href   string
	Active bool
}

type Context[K ~string] interface {
	Locale() string
	Locales() []string
	DefaultLocale() string
	T(key K, vars map[string]any) string
	Path(path string) string
	PathFor(locale string, path string) string
	URL(path string) *url.URL
	URLFor(locale string, path string) *url.URL
	SwitchURL(locale string) *url.URL
	LocaleLinks(alternates map[string]string) []LocaleLink
}

type Runtime[K ~string] struct {
	cfg              Config
	catalog          *Catalog
	compiledMessages map[string]map[K]CompiledMessage
	defaultMessages  map[K]string
}

type contextState[K ~string] struct {
	runtime      *Runtime[K]
	request      *http.Request
	root         *url.URL
	locale       string
	currentPath  string
	currentQuery url.Values
}

func NewRuntime[K ~string](cfg Config, catalog *Catalog, defaultMessages map[K]string) (*Runtime[K], error) {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Runtime[K]{
		cfg:             normalized,
		catalog:         catalog,
		defaultMessages: cloneDefaultMessages(defaultMessages),
	}, nil
}

func NewStaticRuntime[K ~string](
	cfg Config,
	compiledMessages map[string]map[K]CompiledMessage,
	defaultMessages map[K]string,
) (*Runtime[K], error) {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Runtime[K]{
		cfg:              normalized,
		compiledMessages: cloneCompiledMessages(compiledMessages),
		defaultMessages:  cloneDefaultMessages(defaultMessages),
	}, nil
}

func (runtime *Runtime[K]) Config() Config {
	if runtime == nil {
		return Config{}
	}
	return runtime.cfg
}

func (runtime *Runtime[K]) Localize(locale string, key K, vars map[string]any) string {
	if runtime == nil {
		return strings.TrimSpace(string(key))
	}

	normalizedLocale := normalizeLocale(locale)
	if normalizedLocale == "" || !slices.Contains(runtime.cfg.Locales, normalizedLocale) {
		normalizedLocale = runtime.cfg.DefaultLocale
	}

	if message, ok := runtime.lookupCompiledMessage(normalizedLocale, key); ok {
		return message.Render(vars)
	}
	if normalizedLocale != runtime.cfg.DefaultLocale {
		if message, ok := runtime.lookupCompiledMessage(runtime.cfg.DefaultLocale, key); ok {
			return message.Render(vars)
		}
	}

	fallback := strings.TrimSpace(runtime.defaultMessages[key])
	if runtime.catalog == nil {
		if fallback != "" {
			return fallback
		}
		return strings.TrimSpace(string(key))
	}

	return runtime.catalog.Localize(normalizedLocale, string(key), vars, fallback)
}

func (runtime *Runtime[K]) Context(r *http.Request, root *url.URL) Context[K] {
	if runtime == nil {
		return contextState[K]{
			request:      r,
			root:         cloneURL(root),
			currentPath:  "/",
			currentQuery: url.Values{},
		}
	}

	requestCtx := context.Background()
	if r != nil {
		requestCtx = r.Context()
	}
	info, _ := RequestInfoFromContext(requestCtx)
	locale := normalizeLocale(info.Locale)
	if locale == "" || !slices.Contains(runtime.cfg.Locales, locale) {
		locale = runtime.cfg.DefaultLocale
	}

	currentPath := strings.TrimSpace(info.StrippedPath)
	if currentPath == "" {
		if r != nil && r.URL != nil {
			currentPath = r.URL.Path
		}
	}

	currentQuery := url.Values{}
	if r != nil && r.URL != nil {
		currentQuery = r.URL.Query()
	}

	return contextState[K]{
		runtime:      runtime,
		request:      r,
		root:         cloneURL(root),
		locale:       locale,
		currentPath:  NormalizePath(currentPath),
		currentQuery: cloneQuery(currentQuery),
	}
}

func (runtime *Runtime[K]) lookupCompiledMessage(locale string, key K) (CompiledMessage, bool) {
	if runtime == nil || len(runtime.compiledMessages) == 0 {
		return CompiledMessage{}, false
	}

	localeMessages, ok := runtime.compiledMessages[normalizeLocale(locale)]
	if !ok {
		return CompiledMessage{}, false
	}

	message, ok := localeMessages[key]
	if !ok {
		return CompiledMessage{}, false
	}
	return message, true
}

func cloneDefaultMessages[K ~string](defaultMessages map[K]string) map[K]string {
	clonedDefaults := make(map[K]string, len(defaultMessages))
	for key, value := range defaultMessages {
		clonedDefaults[key] = strings.TrimSpace(value)
	}
	return clonedDefaults
}

func cloneCompiledMessages[K ~string](input map[string]map[K]CompiledMessage) map[string]map[K]CompiledMessage {
	if len(input) == 0 {
		return nil
	}

	out := make(map[string]map[K]CompiledMessage, len(input))
	for locale, messages := range input {
		normalizedLocale := normalizeLocale(locale)
		if normalizedLocale == "" {
			continue
		}
		clonedMessages := make(map[K]CompiledMessage, len(messages))
		for key, message := range messages {
			clonedParts := append([]CompiledMessagePart(nil), message.Parts...)
			clonedMessages[key] = CompiledMessage{Parts: clonedParts}
		}
		out[normalizedLocale] = clonedMessages
	}
	return out
}

func (ctx contextState[K]) Locale() string {
	return ctx.locale
}

func (ctx contextState[K]) Locales() []string {
	if ctx.runtime == nil {
		return nil
	}
	return append([]string(nil), ctx.runtime.cfg.Locales...)
}

func (ctx contextState[K]) DefaultLocale() string {
	if ctx.runtime == nil {
		return ""
	}
	return ctx.runtime.cfg.DefaultLocale
}

func (ctx contextState[K]) T(key K, vars map[string]any) string {
	if ctx.runtime == nil {
		return strings.TrimSpace(string(key))
	}
	return ctx.runtime.Localize(ctx.locale, key, vars)
}

func (ctx contextState[K]) Path(pathValue string) string {
	return ctx.PathFor(ctx.locale, pathValue)
}

func (ctx contextState[K]) PathFor(locale string, pathValue string) string {
	if ctx.runtime == nil {
		return NormalizePath(pathValue)
	}
	return LocalizePath(ctx.runtime.cfg, normalizeLocale(locale), pathValue)
}

func (ctx contextState[K]) URL(pathValue string) *url.URL {
	return joinRootAndPath(ctx.root, ctx.Path(pathValue), nil)
}

func (ctx contextState[K]) URLFor(locale string, pathValue string) *url.URL {
	return joinRootAndPath(ctx.root, ctx.PathFor(locale, pathValue), nil)
}

func (ctx contextState[K]) SwitchURL(locale string) *url.URL {
	return joinRootAndPath(ctx.root, ctx.PathFor(locale, ctx.currentPath), ctx.currentQuery)
}

func (ctx contextState[K]) LocaleLinks(alternates map[string]string) []LocaleLink {
	if ctx.runtime == nil {
		return nil
	}

	ordered := orderedLocales(ctx.runtime.cfg)
	items := make([]LocaleLink, 0, len(ordered))
	for _, code := range ordered {
		href := strings.TrimSpace(alternates[code])
		if href == "" {
			switchURL := ctx.SwitchURL(code)
			if switchURL != nil {
				href = switchURL.String()
			}
		}
		if href == "" {
			continue
		}

		items = append(items, LocaleLink{
			Code:   code,
			Label:  displayLabel(ctx.runtime.cfg, code),
			Href:   href,
			Active: code == ctx.locale,
		})
	}

	return items
}

func cloneURL(value *url.URL) *url.URL {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneQuery(values url.Values) url.Values {
	if values == nil {
		return url.Values{}
	}
	clone := url.Values{}
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}

func joinRootAndPath(root *url.URL, pathValue string, query url.Values) *url.URL {
	if root == nil {
		return nil
	}

	normalizedPath := NormalizePath(pathValue)
	clone := *root
	base := strings.TrimSuffix(strings.TrimSpace(clone.Path), "/")

	if normalizedPath == "/" {
		if base == "" {
			clone.Path = "/"
		} else {
			clone.Path = base
		}
	} else {
		joined := path.Join(base, strings.TrimPrefix(normalizedPath, "/"))
		if !strings.HasPrefix(joined, "/") {
			joined = "/" + joined
		}
		clone.Path = joined
	}

	if query == nil {
		clone.RawQuery = ""
	} else {
		clone.RawQuery = cloneQuery(query).Encode()
	}
	clone.Fragment = ""
	return &clone
}

func orderedLocales(cfg Config) []string {
	out := make([]string, 0, len(cfg.Locales))
	seen := make(map[string]struct{}, len(cfg.Locales))

	for _, code := range cfg.DisplayOrder {
		if _, ok := seen[code]; ok {
			continue
		}
		if !slices.Contains(cfg.Locales, code) {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}

	for _, code := range cfg.Locales {
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}

	if len(out) == len(cfg.Locales) {
		return out
	}

	rest := make([]string, 0, len(cfg.Locales)-len(out))
	for _, code := range cfg.Locales {
		if _, ok := seen[code]; ok {
			continue
		}
		rest = append(rest, code)
	}
	sort.Strings(rest)
	return append(out, rest...)
}

func displayLabel(cfg Config, code string) string {
	if label := strings.TrimSpace(cfg.DisplayLabels[code]); label != "" {
		return label
	}
	if code == "" {
		return ""
	}
	return strings.ToUpper(code)
}
