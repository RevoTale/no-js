package seo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"blog/framework"
	frameworki18n "blog/framework/i18n"
	"blog/framework/metagen"
	"blog/internal/notes"
	"blog/internal/web/appcore"
	webi18n "blog/internal/web/i18n"
)

const (
	organizationURL = "https://github.com/RevoTale/blog"
)

func MetaGenRootPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
) (metagen.Metadata, error) {
	view, err := appcore.LoadNotesPage(ctx, appCtx, r, framework.EmptyParams{})
	if err != nil {
		return metagen.Metadata{}, err
	}
	description := localizeSEO(
		appCtx,
		view.LocaleCode(),
		webi18n.KeySeoRootDescription,
		"Browse the latest notes, tales, and micro-tales.",
		nil,
	)
	return notesListingMetadata(appCtx, r, view, description, "website", nil, true)
}

func MetaGenTalesPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
) (metagen.Metadata, error) {
	view, err := appcore.LoadNotesTalesPage(ctx, appCtx, r, framework.EmptyParams{})
	if err != nil {
		return metagen.Metadata{}, err
	}
	description := localizeSEO(
		appCtx,
		view.LocaleCode(),
		webi18n.KeySeoTalesDescription,
		"Read long-form tales from the blog feed.",
		nil,
	)
	return notesListingMetadata(appCtx, r, view, description, "website", nil, true)
}

func MetaGenMicroTalesPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
) (metagen.Metadata, error) {
	view, err := appcore.LoadNotesMicroTalesPage(ctx, appCtx, r, framework.EmptyParams{})
	if err != nil {
		return metagen.Metadata{}, err
	}
	description := localizeSEO(
		appCtx,
		view.LocaleCode(),
		webi18n.KeySeoMicroTalesDescription,
		"Read short-form micro-tales from the blog feed.",
		nil,
	)
	return notesListingMetadata(appCtx, r, view, description, "website", nil, true)
}

func MetaGenTagPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	slug string,
) (metagen.Metadata, error) {
	view, err := appcore.LoadTagPage(ctx, appCtx, r, framework.SlugParams{Slug: slug})
	if err != nil {
		return metagen.Metadata{}, err
	}
	description := localizeSEO(
		appCtx,
		view.LocaleCode(),
		webi18n.KeySeoTagDescription,
		"Browse notes for this tag.",
		map[string]any{
			"Tag": strings.TrimSpace(strings.TrimPrefix(view.PageTitle, "#")),
		},
	)
	return notesListingMetadata(appCtx, r, view, description, "website", nil, true)
}

func MetaGenChannelsPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
) (metagen.Metadata, error) {
	view, err := appcore.LoadChannelsPage(ctx, appCtx, r, framework.EmptyParams{})
	if err != nil {
		return metagen.Metadata{}, err
	}
	description := localizeSEO(
		appCtx,
		view.LocaleCode(),
		webi18n.KeySeoChannelsDescription,
		"Browse available channels and filters for the blog feed.",
		nil,
	)
	return notesListingMetadata(
		appCtx,
		r,
		view,
		description,
		"website",
		&metagen.Robots{Index: metagen.Bool(false), Follow: metagen.Bool(true)},
		false,
	)
}

func MetaGenAuthorPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	slug string,
) (metagen.Metadata, error) {
	view, err := appcore.LoadAuthorPage(ctx, appCtx, r, framework.SlugParams{Slug: slug})
	if err != nil {
		return metagen.Metadata{}, err
	}

	site := siteInfo(appCtx, view.LocaleCode())
	title := titledPage(view.PageTitle, site.Name)
	description := localizeSEO(
		appCtx,
		view.LocaleCode(),
		webi18n.KeySeoAuthorDescription,
		"Browse notes by this author.",
		map[string]any{
			"Author": strings.TrimSpace(view.PageTitle),
		},
	)
	if view.ActiveAuthor != nil && strings.TrimSpace(view.ActiveAuthor.Bio) != "" {
		description = strings.TrimSpace(view.ActiveAuthor.Bio)
	}

	alternates, alternatesErr := buildAlternates(appCtx, r, view.LocaleCode())
	if alternatesErr != nil {
		return metagen.Metadata{}, alternatesErr
	}
	canonicalURL := strings.TrimSpace(alternates.Canonical)

	var authorName string
	var authorSlug string
	var image *metagen.OpenGraphImage
	if view.ActiveAuthor != nil {
		authorName = strings.TrimSpace(view.ActiveAuthor.Name)
		authorSlug = strings.TrimSpace(view.ActiveAuthor.Slug)
		image = authorAvatarImage(appCtx, view.ActiveAuthor)
	}

	openGraph := &metagen.OpenGraph{
		Type:        "profile",
		URL:         canonicalURL,
		SiteName:    site.Name,
		Title:       title,
		Description: description,
		Locale:      view.LocaleCode(),
	}
	twitter := &metagen.Twitter{
		Card:        "summary",
		Title:       title,
		Description: description,
	}
	if image != nil {
		openGraph.Images = []metagen.OpenGraphImage{*image}
		twitter.Card = "summary_large_image"
		twitter.Images = []string{image.URL}
	}

	jsonLD := []metagen.JSONLDDocument{
		organizationDocument(site),
	}
	if authorName != "" {
		person := metagen.JSONLDDocument{
			"@context":    "https://schema.org",
			"@type":       "Person",
			"name":        authorName,
			"description": description,
			"url":         canonicalURL,
		}
		if image != nil {
			person["image"] = image.URL
		}
		jsonLD = append(jsonLD, person)
	}

	authors := []metagen.Author{}
	if authorName != "" {
		authors = append(authors, metagen.Author{
			Name: authorName,
			URL:  absoluteLocalizedURL(appCtx, view.LocaleCode(), "/author/"+authorSlug),
		})
	}

	return metagen.Normalize(metagen.Metadata{
		Title:       title,
		Description: description,
		Alternates:  alternates,
		OpenGraph:   openGraph,
		Twitter:     twitter,
		Authors:     authors,
		Publisher:   site.Publisher,
		Pinterest:   &metagen.Pinterest{RichPin: metagen.Bool(true)},
		JSONLD:      jsonLD,
	}), nil
}

func MetaGenNotePage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	slug string,
) (metagen.Metadata, error) {
	view, err := appcore.LoadNotePage(ctx, appCtx, r, framework.SlugParams{Slug: slug})
	if err != nil {
		return metagen.Metadata{}, err
	}

	site := siteInfo(appCtx, view.LocaleCode())
	title := titledPage(view.PageTitle, site.Name)
	description := strings.TrimSpace(view.Note.Description)
	if description == "" {
		description = localizeSEO(
			appCtx,
			view.LocaleCode(),
			webi18n.KeySeoNoteDescription,
			"Read this note from the blog archive.",
			map[string]any{"Title": strings.TrimSpace(view.Note.Title)},
		)
	}

	alternates, alternatesErr := buildAlternates(appCtx, r, view.LocaleCode())
	if alternatesErr != nil {
		return metagen.Metadata{}, alternatesErr
	}
	canonicalURL := strings.TrimSpace(alternates.Canonical)

	image := noteAttachmentImage(appCtx, view.Note.Attachment)
	openGraph := &metagen.OpenGraph{
		Type:        "article",
		URL:         canonicalURL,
		SiteName:    site.Name,
		Title:       title,
		Description: description,
		Locale:      view.LocaleCode(),
	}
	twitter := &metagen.Twitter{
		Card:        "summary",
		Title:       title,
		Description: description,
	}
	if image != nil {
		openGraph.Images = []metagen.OpenGraphImage{*image}
		twitter.Card = "summary_large_image"
		twitter.Images = []string{image.URL}
	}

	authors := make([]metagen.Author, 0, len(view.Note.Authors))
	authorItems := make([]map[string]any, 0, len(view.Note.Authors))
	for _, author := range view.Note.Authors {
		authorName := strings.TrimSpace(author.Name)
		authorSlug := strings.TrimSpace(author.Slug)
		if authorName == "" {
			continue
		}
		authors = append(authors, metagen.Author{
			Name: authorName,
			URL:  absoluteLocalizedURL(appCtx, view.LocaleCode(), "/author/"+authorSlug),
		})
		authorItems = append(authorItems, map[string]any{
			"@type": "Person",
			"name":  authorName,
			"url":   absoluteLocalizedURL(appCtx, view.LocaleCode(), "/author/"+authorSlug),
		})
	}

	mentions := make([]map[string]any, 0, len(view.Note.Tags))
	keywords := make([]string, 0, len(view.Note.Tags))
	for _, tag := range view.Note.Tags {
		tagName := strings.TrimSpace(tag.Title)
		if tagName == "" {
			tagName = strings.TrimSpace(tag.Name)
		}
		if tagName == "" {
			continue
		}
		keywords = append(keywords, tagName)
		mention := map[string]any{
			"@type": "Thing",
			"name":  tagName,
		}
		tagSlug := strings.TrimSpace(tag.Name)
		if tagSlug != "" {
			mention["url"] = absoluteLocalizedURL(appCtx, view.LocaleCode(), "/tag/"+tagSlug)
		}
		mentions = append(mentions, mention)
	}

	publishing := metagen.JSONLDDocument{
		"@context":         "https://schema.org",
		"@type":            "BlogPosting",
		"headline":         strings.TrimSpace(view.Note.Title),
		"description":      description,
		"mainEntityOfPage": canonicalURL,
		"url":              canonicalURL,
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  site.Publisher,
			"url":   site.RootURL,
		},
	}
	if strings.TrimSpace(view.Note.PublishedAt) != "" {
		publishing["datePublished"] = strings.TrimSpace(view.Note.PublishedAt)
	}
	if len(authorItems) > 0 {
		publishing["author"] = authorItems
	}
	if len(mentions) > 0 {
		publishing["mentions"] = mentions
		publishing["keywords"] = strings.Join(keywords, ", ")
	}
	if image != nil {
		publishing["image"] = image.URL
	}

	return metagen.Normalize(metagen.Metadata{
		Title:       title,
		Description: description,
		Alternates:  alternates,
		OpenGraph:   openGraph,
		Twitter:     twitter,
		Authors:     authors,
		Publisher:   site.Publisher,
		Pinterest:   &metagen.Pinterest{RichPin: metagen.Bool(true)},
		JSONLD: []metagen.JSONLDDocument{
			organizationDocument(site),
			publishing,
		},
	}), nil
}

func notesListingMetadata(
	appCtx *appcore.Context,
	r *http.Request,
	view appcore.NotesPageView,
	description string,
	openGraphType string,
	robots *metagen.Robots,
	includeListingJSONLD bool,
) (metagen.Metadata, error) {
	site := siteInfo(appCtx, view.LocaleCode())
	title := titledPage(view.PageTitle, site.Name)

	alternates, err := buildAlternates(appCtx, r, view.LocaleCode())
	if err != nil {
		return metagen.Metadata{}, err
	}
	canonicalURL := strings.TrimSpace(alternates.Canonical)

	image := firstListingImage(appCtx, view.Notes)
	openGraph := &metagen.OpenGraph{
		Type:        strings.TrimSpace(openGraphType),
		URL:         canonicalURL,
		SiteName:    site.Name,
		Title:       title,
		Description: description,
		Locale:      view.LocaleCode(),
	}
	twitter := &metagen.Twitter{
		Card:        "summary",
		Title:       title,
		Description: description,
	}
	if image != nil {
		openGraph.Images = []metagen.OpenGraphImage{*image}
		twitter.Card = "summary_large_image"
		twitter.Images = []string{image.URL}
	}

	jsonLD := []metagen.JSONLDDocument{
		organizationDocument(site),
	}
	if includeListingJSONLD {
		jsonLD = append(jsonLD, blogListingDocument(appCtx, view, canonicalURL, site, description))
	}

	return metagen.Normalize(metagen.Metadata{
		Title:       title,
		Description: description,
		Alternates:  alternates,
		Robots:      robots,
		OpenGraph:   openGraph,
		Twitter:     twitter,
		Publisher:   site.Publisher,
		JSONLD:      jsonLD,
	}), nil
}

func blogListingDocument(
	appCtx *appcore.Context,
	view appcore.NotesPageView,
	canonicalURL string,
	site siteMetadata,
	description string,
) metagen.JSONLDDocument {
	posts := make([]map[string]any, 0, len(view.Notes))
	for _, note := range view.Notes {
		noteTitle := strings.TrimSpace(note.Title)
		if noteTitle == "" {
			continue
		}
		noteURL := absoluteLocalizedURL(appCtx, view.LocaleCode(), "/note/"+strings.TrimSpace(note.Slug))
		post := map[string]any{
			"@type":            "BlogPosting",
			"headline":         noteTitle,
			"url":              noteURL,
			"mainEntityOfPage": noteURL,
			"description":      strings.TrimSpace(note.Description),
		}
		if strings.TrimSpace(note.PublishedAt) != "" {
			post["datePublished"] = strings.TrimSpace(note.PublishedAt)
		}
		if len(note.Authors) > 0 {
			authors := make([]map[string]any, 0, len(note.Authors))
			for _, author := range note.Authors {
				name := strings.TrimSpace(author.Name)
				slug := strings.TrimSpace(author.Slug)
				if name == "" {
					continue
				}
				authors = append(authors, map[string]any{
					"@type": "Person",
					"name":  name,
					"url":   absoluteLocalizedURL(appCtx, view.LocaleCode(), "/author/"+slug),
				})
			}
			if len(authors) > 0 {
				post["author"] = authors
			}
		}
		if note.Attachment != nil {
			if imageURL := absoluteMediaURL(appCtx, note.Attachment.URL); imageURL != "" {
				post["image"] = imageURL
			}
		}
		posts = append(posts, post)
	}

	return metagen.JSONLDDocument{
		"@context":         "https://schema.org",
		"@type":            "Blog",
		"name":             site.Name,
		"url":              canonicalURL,
		"description":      description,
		"publisher":        map[string]any{"@type": "Organization", "name": site.Publisher, "url": site.RootURL},
		"blogPost":         posts,
		"inLanguage":       view.LocaleCode(),
		"mainEntityOfPage": canonicalURL,
	}
}

func organizationDocument(site siteMetadata) metagen.JSONLDDocument {
	sameAs := []string{organizationURL}
	if rootURL := strings.TrimSpace(site.RootURL); rootURL != "" {
		sameAs = append([]string{rootURL}, sameAs...)
	}

	out := metagen.JSONLDDocument{
		"@context":    "https://schema.org",
		"@type":       "Organization",
		"name":        site.Publisher,
		"url":         site.RootURL,
		"description": site.Description,
		"sameAs":      sameAs,
	}
	return out
}

type siteMetadata struct {
	Name        string
	Description string
	Publisher   string
	RootURL     string
}

func siteInfo(appCtx *appcore.Context, locale string) siteMetadata {
	name := localizeSEO(appCtx, locale, webi18n.KeySeoSiteName, "blog", nil)
	description := localizeSEO(
		appCtx,
		locale,
		webi18n.KeySeoSiteDescription,
		"A multilingual note feed with tales and micro-tales.",
		nil,
	)
	publisher := localizeSEO(appCtx, locale, webi18n.KeySeoPublisherName, "RevoTale", nil)

	rootURL := ""
	if appCtx != nil {
		rootURL = strings.TrimSpace(appCtx.RootURL())
	}

	return siteMetadata{
		Name:        name,
		Description: description,
		Publisher:   publisher,
		RootURL:     rootURL,
	}
}

func localizeSEO(
	appCtx *appcore.Context,
	locale string,
	key webi18n.Key,
	fallback string,
	data map[string]any,
) string {
	normalizedKey := webi18n.Key(strings.TrimSpace(string(key)))
	if appCtx == nil {
		return strings.TrimSpace(fallback)
	}
	value := strings.TrimSpace(appCtx.T(locale, normalizedKey, data))
	if value == "" || value == string(normalizedKey) {
		return strings.TrimSpace(fallback)
	}
	return value
}

func titledPage(pageTitle string, siteName string) string {
	trimmedPage := strings.TrimSpace(pageTitle)
	trimmedSite := strings.TrimSpace(siteName)
	if trimmedSite == "" {
		trimmedSite = "blog"
	}
	if trimmedPage == "" {
		return trimmedSite
	}
	return trimmedPage + " :: " + trimmedSite
}

func buildAlternates(appCtx *appcore.Context, r *http.Request, locale string) (metagen.Alternates, error) {
	if appCtx == nil {
		return metagen.Alternates{}, fmt.Errorf("app context is required")
	}

	rootURL := strings.TrimSpace(appCtx.RootURL())
	if rootURL == "" {
		return metagen.Alternates{}, fmt.Errorf("BLOG_ROOT_URL is required for metadata alternates")
	}

	cfg, err := frameworki18n.NormalizeConfig(appCtx.I18nConfig())
	if err != nil {
		return metagen.Alternates{}, fmt.Errorf("normalize i18n config: %w", err)
	}

	return metagen.BuildAlternates(rootURL, cfg, locale, requestPathWithQuery(r), nil)
}

func requestPathWithQuery(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "/"
	}
	pathValue := strings.TrimSpace(r.URL.Path)
	if pathValue == "" {
		pathValue = "/"
	}
	if strings.TrimSpace(r.URL.RawQuery) == "" {
		return pathValue
	}
	return pathValue + "?" + strings.TrimSpace(r.URL.RawQuery)
}

func firstListingImage(appCtx *appcore.Context, notes []notes.NoteSummary) *metagen.OpenGraphImage {
	for _, note := range notes {
		if note.Attachment == nil {
			continue
		}
		return noteAttachmentImage(appCtx, note.Attachment)
	}
	return nil
}

func noteAttachmentImage(appCtx *appcore.Context, attachment *notes.Attachment) *metagen.OpenGraphImage {
	if attachment == nil {
		return nil
	}
	imageURL := absoluteMediaURL(appCtx, attachment.URL)
	if imageURL == "" {
		return nil
	}
	return &metagen.OpenGraphImage{
		URL:    imageURL,
		Alt:    strings.TrimSpace(attachment.Alt),
		Width:  attachment.Width,
		Height: attachment.Height,
	}
}

func authorAvatarImage(appCtx *appcore.Context, author *notes.Author) *metagen.OpenGraphImage {
	if author == nil || author.Avatar == nil {
		return nil
	}
	imageURL := absoluteMediaURL(appCtx, author.Avatar.URL)
	if imageURL == "" {
		return nil
	}
	return &metagen.OpenGraphImage{
		URL:    imageURL,
		Alt:    strings.TrimSpace(author.Avatar.Alt),
		Width:  author.Avatar.Width,
		Height: author.Avatar.Height,
	}
}

func absoluteMediaURL(appCtx *appcore.Context, rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return ""
	}
	if parsed.IsAbs() && strings.TrimSpace(parsed.Host) != "" {
		return parsed.String()
	}

	root := ""
	if appCtx != nil {
		root = strings.TrimSpace(appCtx.RootURL())
	}
	return joinRootAndPath(root, parsed.Path)
}

func absoluteLocalizedURL(appCtx *appcore.Context, locale string, strippedPath string) string {
	localizedPath := appcore.LocalizeAppPath(locale, strippedPath)
	root := ""
	if appCtx != nil {
		root = strings.TrimSpace(appCtx.RootURL())
	}
	return joinRootAndPath(root, localizedPath)
}

func joinRootAndPath(rootURL string, routePath string) string {
	trimmedPath := strings.TrimSpace(routePath)
	if trimmedPath == "" {
		trimmedPath = "/"
	}
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	parsedRoot, err := url.Parse(strings.TrimSpace(rootURL))
	if err != nil || !parsedRoot.IsAbs() || strings.TrimSpace(parsedRoot.Host) == "" {
		return trimmedPath
	}

	base := strings.TrimSuffix(strings.TrimSpace(parsedRoot.Path), "/")
	if trimmedPath == "/" {
		if base == "" {
			parsedRoot.Path = "/"
		} else {
			parsedRoot.Path = base
		}
		parsedRoot.RawQuery = ""
		parsedRoot.Fragment = ""
		return parsedRoot.String()
	}

	joined := path.Join(base, strings.TrimPrefix(trimmedPath, "/"))
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	parsedRoot.Path = joined
	parsedRoot.RawQuery = ""
	parsedRoot.Fragment = ""
	return parsedRoot.String()
}
