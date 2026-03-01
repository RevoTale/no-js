package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"blog/framework/httpserver"
	"blog/framework/staticassets"
	"blog/internal/config"
	"blog/internal/gql"
	"blog/internal/notes"
	"blog/internal/web/appcore"
	webgen "blog/internal/web/gen"
)

const immutableStaticCachePolicy = "public, max-age=31536000, immutable"

func main() {
	if err := run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func run() error {
	cfg := config.Load()

	manifest, err := staticassets.ReadManifest(cfg.StaticManifestPath)
	if err != nil {
		return fmt.Errorf(
			"load static manifest %q: %w (run staticassetsgen during build)",
			cfg.StaticManifestPath,
			err,
		)
	}

	appcore.SetStaticAssetBasePath(manifest.URLPrefix)

	graphqlClient := gql.NewClient(cfg)
	noteService := notes.NewService(graphqlClient, cfg.PageSize, cfg.RootURL)
	cachePolicies := httpserver.DefaultCachePolicies()
	if strings.TrimSpace(cfg.CacheLiveNavigation) != "" {
		cachePolicies.LiveNavigation = cfg.CacheLiveNavigation
	}
	cachePolicies.Static = immutableStaticCachePolicy

	// Serve static files from the manifest directory so routing cannot drift
	// to an unprocessed source dir if env configuration is stale.
	staticDir := filepath.Clean(filepath.Dir(cfg.StaticManifestPath))
	if info, statErr := os.Stat(staticDir); statErr != nil {
		return fmt.Errorf("stat static build dir %q: %w", staticDir, statErr)
	} else if !info.IsDir() {
		return fmt.Errorf("static build dir %q is not a directory", staticDir)
	}
	log.Printf(
		"static assets bundle loaded: hash=%s prefix=%s dir=%s manifest=%s",
		manifest.Hash,
		manifest.URLPrefix,
		staticDir,
		cfg.StaticManifestPath,
	)

	handler, err := httpserver.New(httpserver.Config[*appcore.Context]{
		AppContext:      appcore.NewContext(noteService),
		Handlers:        webgen.Handlers(webgen.NewRouteResolvers()),
		IsNotFoundError: appcore.IsNotFoundError,
		NotFoundPage:    webgen.NotFoundPage,
		Static: httpserver.StaticMount{
			URLPrefix: manifest.URLPrefix,
			Dir:       staticDir,
		},
		CachePolicies: cachePolicies,
		LogServerError: func(err error) {
			log.Printf("blog server error: %v", err)
		},
	})
	if err != nil {
		return fmt.Errorf("handler setup failed: %w", err)
	}

	log.Printf("blog server listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, handler); err != nil {
		return err
	}

	return nil
}
