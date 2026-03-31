package httpserver

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const defaultPublicFilesCachePolicy = "public, max-age=0"

type PublicFilesConfig struct {
	Dir               string
	RequestPathPrefix string
	CachePolicy       string
}

type publicFilesHandler struct {
	index       map[string]string
	cachePolicy string
}

func (cfg PublicFilesConfig) WithPublicFileCachePolicy(policy string) PublicFilesConfig {
	cfg.CachePolicy = strings.TrimSpace(policy)
	return cfg
}

func (cfg PublicFilesConfig) WithRequestPathPrefix(prefix string) PublicFilesConfig {
	cfg.RequestPathPrefix = strings.TrimSpace(prefix)
	return cfg
}

func newPublicFilesHandler(cfg PublicFilesConfig) (*publicFilesHandler, error) {
	publicDir := strings.TrimSpace(cfg.Dir)
	if publicDir == "" {
		return nil, fmt.Errorf("public dir is required")
	}

	info, err := os.Stat(publicDir)
	if err != nil {
		return nil, fmt.Errorf("stat public dir %q: %w", publicDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("public dir %q is not a directory", publicDir)
	}

	index, err := buildPublicFilesIndex(publicDir, normalizePublicFilesRequestPathPrefix(cfg.RequestPathPrefix))
	if err != nil {
		return nil, err
	}

	cachePolicy := strings.TrimSpace(cfg.CachePolicy)
	if cachePolicy == "" {
		cachePolicy = defaultPublicFilesCachePolicy
	}

	return &publicFilesHandler{
		index:       index,
		cachePolicy: cachePolicy,
	}, nil
}

func WithPublicFiles(cfg PublicFilesConfig) (func(http.Handler) http.Handler, error) {
	handler, err := newPublicFilesHandler(cfg)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if next == nil {
				return
			}
			if !handler.ServeHTTP(w, r) {
				next.ServeHTTP(w, r)
			}
		})
	}, nil
}

func (handler *publicFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	if handler == nil || r == nil || r.URL == nil {
		return false
	}

	publicPath := normalizePublicRequestPath(r.URL.Path)
	filePath, ok := handler.index[publicPath]
	if !ok {
		return false
	}
	if !isReadMethod(r.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true
	}

	if contentType := publicFileContentType(filePath); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	setCachePolicy(w, handler.cachePolicy)
	http.ServeFile(w, r, filePath)
	return true
}

func buildPublicFilesIndex(publicDir string, requestPathPrefix string) (map[string]string, error) {
	index := make(map[string]string)
	err := filepath.WalkDir(publicDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(publicDir, path)
		if err != nil {
			return fmt.Errorf("resolve public relative path for %q: %w", path, err)
		}
		normalizedRel := filepath.ToSlash(relPath)
		if strings.TrimSpace(normalizedRel) == "" || normalizedRel == "." {
			return nil
		}

		normalizedRel = strings.TrimPrefix(normalizedRel, "/")
		normalizedRel = strings.TrimPrefix(normalizedRel, "./")
		if normalizedRel == "" || strings.HasPrefix(normalizedRel, "../") || normalizedRel == ".." {
			return fmt.Errorf("invalid public file path %q", relPath)
		}

		publicPath := joinPublicRequestPath(requestPathPrefix, normalizedRel)
		index[publicPath] = path
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("index public dir %q: %w", publicDir, err)
	}
	return index, nil
}

func normalizePublicFilesRequestPathPrefix(pathValue string) string {
	trimmed := strings.TrimSpace(pathValue)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if trimmed != "/" && strings.HasSuffix(trimmed, "/") {
		trimmed = strings.TrimRight(trimmed, "/")
	}
	return trimmed
}

func joinPublicRequestPath(prefix string, relativePath string) string {
	trimmedRelative := strings.Trim(strings.TrimSpace(relativePath), "/")
	if trimmedRelative == "" {
		return normalizePublicFilesRequestPathPrefix(prefix)
	}

	normalizedPrefix := normalizePublicFilesRequestPathPrefix(prefix)
	if normalizedPrefix == "/" {
		return "/" + trimmedRelative
	}

	return normalizedPrefix + "/" + trimmedRelative
}

func normalizePublicRequestPath(pathValue string) string {
	trimmed := strings.TrimSpace(pathValue)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if !strings.HasPrefix(cleaned, "/") {
		return "/" + cleaned
	}
	return cleaned
}

func isReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func publicFileContentType(filePath string) string {
	extension := strings.ToLower(filepath.Ext(filePath))
	switch extension {
	case ".webmanifest":
		return "application/manifest+json"
	}

	return strings.TrimSpace(mime.TypeByExtension(extension))
}
