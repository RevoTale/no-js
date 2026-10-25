package httpserver

import "strings"

// Already compressed media has no useful second gzip layer. Keep unknown types
// on the existing compression path; notably SVG and WebAssembly remain compressible.
func alreadyCompressedContentType(value string) bool {
	mediaType, _, _ := strings.Cut(value, ";")
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/avif", "image/jxl", "image/heic", "image/heif",
		"image/vnd.microsoft.icon", "font/woff", "font/woff2", "application/font-woff", "application/zip",
		"application/gzip", "application/x-gzip", "application/x-7z-compressed", "application/x-xz",
		"audio/mpeg", "audio/ogg", "audio/aac", "video/mp4", "video/webm":
		return true
	default:
		return false
	}
}
