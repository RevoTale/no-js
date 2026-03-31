package discovery

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/RevoTale/no-js/framework"
)

const (
	RobotsPath          = "/robots.txt"
	FeedPath            = "/feed.xml"
	SitemapPath         = "/sitemap.xml"
	SitemapIndexPath    = "/sitemap-index"
	SitemapIndexXMLPath = "/sitemap-index.xml"

	defaultRobotsCachePolicy       = "public, max-age=3600, s-maxage=3600"
	defaultFeedCachePolicy         = "public, max-age=3600, s-maxage=3600"
	defaultSitemapCachePolicy      = "public, max-age=3600, s-maxage=3600"
	defaultSitemapIndexCachePolicy = "public, max-age=3600, s-maxage=3600, " +
		"stale-while-revalidate=9000, stale-if-error=86400"

	contentTypePlainText      = "text/plain; charset=utf-8"
	contentTypeRSSXML         = "application/rss+xml; charset=utf-8"
	contentTypeApplicationXML = "application/xml; charset=utf-8"
)

// Bundle contains the optional reserved discovery conventions generated from
// web/routes.
type Bundle[C interface{}] struct {
	Robots           func(runtime framework.RuntimeContext[C], r *http.Request) (Robots, error)
	Sitemap          func(runtime framework.RuntimeContext[C], r *http.Request) ([]SitemapEntry, error)
	GenerateSitemaps func(runtime framework.RuntimeContext[C], r *http.Request) ([]SitemapID, error)
	SitemapByID      func(runtime framework.RuntimeContext[C], r *http.Request, id string) ([]SitemapEntry, error)
	Feed             func(runtime framework.RuntimeContext[C], r *http.Request) (FeedDocument, error)
}

// Robots is the app-facing robots.txt model returned from web/routes/robots.go.
// The framework renders it into plain text using RFC 9309-style groups plus the
// optional Sitemap and Host directives.
type Robots struct {
	// Rules maps to one or more robots.txt groups.
	Rules []RobotsRule
	// Sitemaps is rendered as repeated absolute "Sitemap:" lines.
	Sitemaps []string
	// Host is rendered as a trailing "Host:" line. It is a common extension, not
	// part of RFC 9309 itself.
	Host string
}

// RobotsRule describes one robots.txt group keyed by a User-agent value.
type RobotsRule struct {
	// UserAgent maps to the "User-agent:" line. Empty falls back to "*".
	UserAgent string
	// Allow maps to repeated "Allow:" lines inside the group.
	Allow []string
	// Disallow maps to repeated "Disallow:" lines inside the group.
	Disallow []string
}

// SitemapEntry is one <url> record returned from web/routes/sitemap.go or
// SitemapByID. The framework serializes it into the XML sitemap protocol and
// related extensions.
type SitemapEntry struct {
	// URL maps to <loc> and should be an absolute canonical URL.
	URL string
	// Alternates maps hreflang codes to absolute URLs and is rendered as
	// <xhtml:link rel="alternate" ...>.
	Alternates map[string]string
	// Images adds image sitemap extension entries for the URL.
	Images []SitemapImage
	// LastModified maps to <lastmod>.
	LastModified *time.Time
	// ChangeFrequency maps to <changefreq>.
	ChangeFrequency string
	// Priority maps to <priority>. Use values in the sitemap protocol range.
	Priority *float64
}

// SitemapImage is one image sitemap extension entry for a SitemapEntry.
type SitemapImage struct {
	// URL maps to <image:loc> and should be absolute.
	URL string
}

// SitemapID identifies one generated sitemap chunk returned from
// GenerateSitemaps and later resolved by SitemapByID.
type SitemapID struct {
	// ID is the stable application key passed back into SitemapByID.
	ID string
	// Path is the request path for the chunk when the framework should derive the
	// public location from the current request.
	Path string
	// Location overrides the absolute URL emitted in sitemap indexes.
	Location string
}

// FeedDocument is the app-facing RSS channel model returned from
// web/routes/feed.go. The framework renders it as an RSS 2.0 feed with an Atom
// self link when SelfURL is set.
type FeedDocument struct {
	// Title maps to <channel><title>.
	Title string
	// Link maps to <channel><link>.
	Link string
	// Description maps to <channel><description>.
	Description string
	// Language maps to <channel><language>.
	Language string
	// LastBuildDate maps to <channel><lastBuildDate>.
	LastBuildDate *time.Time
	// Generator maps to <channel><generator>.
	Generator string
	// Copyright maps to <channel><copyright>.
	Copyright string
	// SelfURL is emitted as the Atom <link rel="self" type="application/rss+xml">.
	SelfURL string
	// Items maps to repeated <item> nodes.
	Items []FeedItem
}

// FeedItem is one RSS 2.0 <item> entry inside a FeedDocument.
type FeedItem struct {
	// Title maps to <item><title>.
	Title string
	// Link maps to <item><link>.
	Link string
	// GUID maps to <item><guid>.
	GUID string
	// Description maps to <item><description>.
	Description string
	// Author maps to <item><author>.
	Author string
	// PublishedAt maps to <item><pubDate>.
	PublishedAt *time.Time
	// Categories maps to repeated <item><category> elements.
	Categories []string
}

func MaybeServe[C interface{}](
	runtime framework.RuntimeContext[C],
	bundle *Bundle[C],
	logServerError func(error),
	w http.ResponseWriter,
	r *http.Request,
) bool {
	if bundle == nil || r == nil || r.URL == nil {
		return false
	}

	requestPath := normalizePath(r.URL.Path)
	switch requestPath {
	case RobotsPath:
		if bundle.Robots == nil {
			return false
		}
		if !isReadMethod(r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return true
		}

		document, err := bundle.Robots(runtime, r)
		if err != nil {
			writeInternalServerError(logServerError, w, fmt.Errorf("resolve robots: %w", err))
			return true
		}

		payload := renderRobotsTXT(document)
		writeResponse(w, r, http.StatusOK, defaultRobotsCachePolicy, contentTypePlainText, payload)
		return true
	case FeedPath:
		if bundle.Feed == nil {
			return false
		}
		if !isReadMethod(r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return true
		}

		document, err := bundle.Feed(runtime, r)
		if err != nil {
			writeInternalServerError(logServerError, w, fmt.Errorf("resolve feed: %w", err))
			return true
		}

		payload, err := renderFeedXML(document)
		if err != nil {
			writeInternalServerError(logServerError, w, fmt.Errorf("render feed: %w", err))
			return true
		}

		writeResponse(w, r, http.StatusOK, defaultFeedCachePolicy, contentTypeRSSXML, payload)
		return true
	case SitemapPath:
		if bundle.Sitemap != nil {
			if !isReadMethod(r.Method) {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return true
			}

			entries, err := bundle.Sitemap(runtime, r)
			if err != nil {
				writeInternalServerError(logServerError, w, fmt.Errorf("resolve sitemap: %w", err))
				return true
			}

			payload, err := renderSitemapXML(entries)
			if err != nil {
				writeInternalServerError(logServerError, w, fmt.Errorf("render sitemap: %w", err))
				return true
			}

			writeResponse(w, r, http.StatusOK, defaultSitemapCachePolicy, contentTypeApplicationXML, payload)
			return true
		}
	case SitemapIndexPath, SitemapIndexXMLPath:
		if bundle.Sitemap == nil && bundle.GenerateSitemaps == nil {
			return false
		}
		if !isReadMethod(r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return true
		}

		ids, err := sitemapIDsForRequest(runtime, bundle, r)
		if err != nil {
			writeInternalServerError(logServerError, w, fmt.Errorf("resolve sitemap index: %w", err))
			return true
		}

		locations := make([]string, 0, len(ids)+1)
		seen := map[string]struct{}{}
		if bundle.Sitemap != nil {
			rootLocation := requestAbsoluteURL(r, SitemapPath)
			if rootID, ok := matchSitemapIDByPath(ids, SitemapPath); ok {
				rootLocation = resolveSitemapLocation(rootID, r)
			}
			appendUniqueNonEmpty(&locations, seen, rootLocation)
		}
		for _, id := range ids {
			if normalizePath(id.Path) == SitemapPath {
				continue
			}
			appendUniqueNonEmpty(&locations, seen, resolveSitemapLocation(id, r))
		}

		payload, err := renderSitemapIndexXML(locations)
		if err != nil {
			writeInternalServerError(logServerError, w, fmt.Errorf("render sitemap index: %w", err))
			return true
		}

		writeResponse(w, r, http.StatusOK, defaultSitemapIndexCachePolicy, contentTypeApplicationXML, payload)
		return true
	}

	if bundle.GenerateSitemaps == nil || bundle.SitemapByID == nil {
		return false
	}
	if !isReadMethod(r.Method) {
		return false
	}

	ids, err := sitemapIDsForRequest(runtime, bundle, r)
	if err != nil {
		writeInternalServerError(logServerError, w, fmt.Errorf("resolve sitemap ids: %w", err))
		return true
	}
	matchedID, ok := matchSitemapIDByPath(ids, requestPath)
	if !ok {
		return false
	}

	entries, err := bundle.SitemapByID(runtime, r, strings.TrimSpace(matchedID.ID))
	if err != nil {
		writeInternalServerError(logServerError, w, fmt.Errorf("resolve sitemap %q: %w", matchedID.ID, err))
		return true
	}

	payload, err := renderSitemapXML(entries)
	if err != nil {
		writeInternalServerError(logServerError, w, fmt.Errorf("render sitemap %q: %w", matchedID.ID, err))
		return true
	}

	writeResponse(w, r, http.StatusOK, defaultSitemapCachePolicy, contentTypeApplicationXML, payload)
	return true
}

func sitemapIDsForRequest[C interface{}](
	runtime framework.RuntimeContext[C],
	bundle *Bundle[C],
	r *http.Request,
) ([]SitemapID, error) {
	if bundle == nil || bundle.GenerateSitemaps == nil {
		return nil, nil
	}
	return bundle.GenerateSitemaps(runtime, r)
}

func matchSitemapIDByPath(ids []SitemapID, requestPath string) (SitemapID, bool) {
	normalizedRequestPath := normalizePath(requestPath)
	for _, id := range ids {
		if normalizePath(id.Path) == normalizedRequestPath {
			return id, true
		}
	}
	return SitemapID{}, false
}

func resolveSitemapLocation(id SitemapID, r *http.Request) string {
	location := strings.TrimSpace(id.Location)
	if location != "" {
		return location
	}
	return requestAbsoluteURL(r, id.Path)
}

func renderRobotsTXT(document Robots) string {
	lines := make([]string, 0, len(document.Rules)*4+len(document.Sitemaps)+1)
	rules := document.Rules
	if len(rules) == 0 {
		rules = []RobotsRule{{UserAgent: "*", Allow: []string{"/"}}}
	}

	for _, rule := range rules {
		userAgent := strings.TrimSpace(rule.UserAgent)
		if userAgent == "" {
			userAgent = "*"
		}
		lines = append(lines, "User-agent: "+userAgent)
		for _, allow := range rule.Allow {
			appendNonEmptyLine(&lines, "Allow: ", allow)
		}
		for _, disallow := range rule.Disallow {
			appendNonEmptyLine(&lines, "Disallow: ", disallow)
		}
	}

	for _, sitemap := range document.Sitemaps {
		appendNonEmptyLine(&lines, "Sitemap: ", sitemap)
	}
	appendNonEmptyLine(&lines, "Host: ", document.Host)

	return strings.Join(lines, "\n") + "\n"
}

func renderFeedXML(document FeedDocument) (string, error) {
	view := rssXMLTemplateData{
		XMLHeader:     xml.Header,
		Title:         strings.TrimSpace(document.Title),
		Link:          strings.TrimSpace(document.Link),
		Description:   strings.TrimSpace(document.Description),
		Language:      strings.TrimSpace(document.Language),
		LastBuildDate: formatRFC1123Z(document.LastBuildDate),
		Generator:     strings.TrimSpace(document.Generator),
		Copyright:     strings.TrimSpace(document.Copyright),
		FeedURL:       strings.TrimSpace(document.SelfURL),
		Items:         make([]rssItem, 0, len(document.Items)),
	}
	if view.LastBuildDate == "" {
		view.LastBuildDate = time.Now().UTC().Format(time.RFC1123Z)
	}

	for _, item := range document.Items {
		view.Items = append(view.Items, rssItem{
			Title:       strings.TrimSpace(item.Title),
			Link:        strings.TrimSpace(item.Link),
			GUID:        strings.TrimSpace(item.GUID),
			Description: strings.TrimSpace(item.Description),
			Author:      strings.TrimSpace(item.Author),
			PubDate:     formatRFC1123Z(item.PublishedAt),
			Categories:  compactNonEmptyStrings(item.Categories),
		})
	}

	return executeXMLTemplate(rssXMLTemplate, view)
}

func renderSitemapXML(entries []SitemapEntry) (string, error) {
	return executeXMLTemplate(sitemapXMLTemplate, newSitemapXMLTemplateData(entries))
}

func renderSitemapIndexXML(locations []string) (string, error) {
	return executeXMLTemplate(
		sitemapIndexXMLTemplate,
		sitemapIndexXMLTemplateData{
			XMLHeader: xml.Header,
			Locations: compactNonEmptyStrings(locations),
		},
	)
}

func writeInternalServerError(logServerError func(error), w http.ResponseWriter, err error) {
	if logServerError != nil {
		logServerError(err)
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func writeResponse(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	cachePolicy string,
	contentType string,
	payload string,
) {
	if strings.TrimSpace(cachePolicy) != "" {
		w.Header().Set("Cache-Control", strings.TrimSpace(cachePolicy))
	}
	if strings.TrimSpace(contentType) != "" {
		w.Header().Set("Content-Type", strings.TrimSpace(contentType))
	}
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	if r != nil && r.Method == http.MethodHead {
		return
	}
	_, _ = io.WriteString(w, payload)
}

func isReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func normalizePath(pathValue string) string {
	trimmed := strings.TrimSpace(pathValue)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return path.Clean(trimmed)
}

func requestAbsoluteURL(r *http.Request, routePath string) string {
	normalizedPath := normalizePath(routePath)
	if r == nil {
		return normalizedPath
	}

	host := strings.TrimSpace(r.Host)
	if host == "" && r.URL != nil {
		host = strings.TrimSpace(r.URL.Host)
	}
	if host == "" {
		return normalizedPath
	}

	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if comma := strings.IndexByte(scheme, ','); comma >= 0 {
		scheme = scheme[:comma]
	}
	scheme = strings.TrimSpace(scheme)
	if scheme == "" && r.URL != nil {
		scheme = strings.TrimSpace(r.URL.Scheme)
	}
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	out := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   normalizedPath,
	}
	return out.String()
}

func appendNonEmptyLine(lines *[]string, prefix string, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}
	*lines = append(*lines, prefix+trimmed)
}

func appendUniqueNonEmpty(values *[]string, seen map[string]struct{}, candidate string) {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		return
	}
	if _, ok := seen[trimmed]; ok {
		return
	}
	seen[trimmed] = struct{}{}
	*values = append(*values, trimmed)
}

func formatRFC1123Z(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC1123Z)
}

type rssItem struct {
	Title       string
	Link        string
	GUID        string
	Description string
	Author      string
	PubDate     string
	Categories  []string
}

type rssXMLTemplateData struct {
	XMLHeader     string
	Title         string
	Link          string
	Description   string
	Language      string
	LastBuildDate string
	Generator     string
	Copyright     string
	FeedURL       string
	Items         []rssItem
}

type sitemapAlternate struct {
	Locale string
	Href   string
}

type sitemapXMLEntry struct {
	Loc        string
	Alternates []sitemapAlternate
	Images     []string
	LastMod    string
	ChangeFreq string
	Priority   string
}

type sitemapXMLTemplateData struct {
	XMLHeader      string
	WithAlternates bool
	WithImages     bool
	Entries        []sitemapXMLEntry
}

type sitemapIndexXMLTemplateData struct {
	XMLHeader string
	Locations []string
}

func newSitemapXMLTemplateData(entries []SitemapEntry) sitemapXMLTemplateData {
	view := sitemapXMLTemplateData{
		XMLHeader: xml.Header,
		Entries:   make([]sitemapXMLEntry, 0, len(entries)),
	}

	for _, entry := range entries {
		location := strings.TrimSpace(entry.URL)
		if location == "" {
			continue
		}

		outEntry := sitemapXMLEntry{
			Loc:        location,
			Alternates: make([]sitemapAlternate, 0, len(entry.Alternates)),
			Images:     make([]string, 0, len(entry.Images)),
			LastMod:    formatRFC3339(entry.LastModified),
			ChangeFreq: strings.TrimSpace(entry.ChangeFrequency),
		}
		if entry.Priority != nil {
			outEntry.Priority = strings.TrimSpace(fmt.Sprintf("%.1f", *entry.Priority))
		}

		keys := make([]string, 0, len(entry.Alternates))
		for key := range entry.Alternates {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			href := strings.TrimSpace(entry.Alternates[key])
			if href == "" {
				continue
			}
			outEntry.Alternates = append(outEntry.Alternates, sitemapAlternate{
				Locale: strings.TrimSpace(key),
				Href:   href,
			})
		}
		for _, image := range entry.Images {
			location := strings.TrimSpace(image.URL)
			if location == "" {
				continue
			}
			outEntry.Images = append(outEntry.Images, location)
		}

		if len(outEntry.Alternates) > 0 {
			view.WithAlternates = true
		}
		if len(outEntry.Images) > 0 {
			view.WithImages = true
		}
		view.Entries = append(view.Entries, outEntry)
	}

	return view
}

func formatRFC3339(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func compactNonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func executeXMLTemplate(template *texttemplate.Template, data any) (string, error) {
	var buffer bytes.Buffer
	if err := template.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

var xmlTemplateFunctions = texttemplate.FuncMap{
	"xml": xmlEscape,
}

var rssXMLTemplate = texttemplate.Must(
	texttemplate.New("rss.xml").Funcs(xmlTemplateFunctions).Parse(
		`{{.XMLHeader}}<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>{{xml .Title}}</title>
    <link>{{xml .Link}}</link>
    <description>{{xml .Description}}</description>
    <language>{{xml .Language}}</language>
    <lastBuildDate>{{xml .LastBuildDate}}</lastBuildDate>
    <generator>{{xml .Generator}}</generator>
    <copyright>{{xml .Copyright}}</copyright>
    <atom:link href="{{xml .FeedURL}}" rel="self" type="application/rss+xml"/>
{{- range .Items}}
    <item>
      <title>{{xml .Title}}</title>
      <link>{{xml .Link}}</link>
      <guid>{{xml .GUID}}</guid>
{{- if .Description}}
      <description>{{xml .Description}}</description>
{{- end}}
{{- if .Author}}
      <author>{{xml .Author}}</author>
{{- end}}
{{- if .PubDate}}
      <pubDate>{{xml .PubDate}}</pubDate>
{{- end}}
{{- range .Categories}}
      <category>{{xml .}}</category>
{{- end}}
    </item>
{{- end}}
  </channel>
</rss>
`,
	),
)

var sitemapXMLTemplate = texttemplate.Must(
	texttemplate.New("sitemap.xml").Funcs(xmlTemplateFunctions).Parse(
		`{{.XMLHeader}}<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"{{if .WithAlternates}}
 xmlns:xhtml="http://www.w3.org/1999/xhtml"{{end}}{{if .WithImages}}
 xmlns:image="http://www.google.com/schemas/sitemap-image/1.1"{{end}}>
{{- range .Entries}}
  <url>
    <loc>{{xml .Loc}}</loc>
{{- range .Alternates}}
    <xhtml:link rel="alternate" hreflang="{{xml .Locale}}" href="{{xml .Href}}"/>
{{- end}}
{{- if .LastMod}}
    <lastmod>{{xml .LastMod}}</lastmod>
{{- end}}
{{- if .ChangeFreq}}
    <changefreq>{{xml .ChangeFreq}}</changefreq>
{{- end}}
{{- if .Priority}}
    <priority>{{xml .Priority}}</priority>
{{- end}}
{{- range .Images}}
    <image:image><image:loc>{{xml .}}</image:loc></image:image>
{{- end}}
  </url>
{{- end}}
</urlset>
`,
	),
)

var sitemapIndexXMLTemplate = texttemplate.Must(
	texttemplate.New("sitemap-index.xml").Funcs(xmlTemplateFunctions).Parse(
		`{{.XMLHeader}}<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{- range .Locations}}
  <sitemap><loc>{{xml .}}</loc></sitemap>
{{- end}}
</sitemapindex>
`,
	),
)

func xmlEscape(value string) string {
	var builder strings.Builder
	if err := xml.EscapeText(&builder, []byte(value)); err != nil {
		return value
	}
	return builder.String()
}
