package imageloader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoaderURL_DisabledReturnsOriginal(t *testing.T) {
	t.Parallel()

	loader := New(false)
	const src = "/images/hello world.webp"
	assert.Equal(t, src, loader.URL(src, 640))
}

func TestLoaderURL_RewritesCDNS3Width(t *testing.T) {
	t.Parallel()

	loader := New(true)
	got := loader.URL("/cdn/image/s3/828/files/pic.webp", 1080)
	want := "/cdn/image/s3/1080/files/pic.webp"
	assert.Equal(t, want, got)
}

func TestLoaderURL_BuildsRelativeEndpointAndEncodesSpaces(t *testing.T) {
	t.Parallel()

	loader := New(true)
	got := loader.URL("/images/hello world.webp", 640)
	want := "/cdn/image/relative/640/images/hello%20world.webp"
	assert.Equal(t, want, got)
}

func TestLoaderURL_NormalizesToNearestAllowedWidth(t *testing.T) {
	t.Parallel()

	loader := New(true)
	assert.Equal(t, "/cdn/image/relative/48/images/pic.webp", loader.URL("/images/pic.webp", 40))
	assert.Equal(t, "/cdn/image/relative/828/images/pic.webp", loader.URL("/images/pic.webp", 768))
}

func TestLoaderResponsiveSrcSet_UsesCMSDeviceWidths(t *testing.T) {
	t.Parallel()

	loader := New(true)
	got, err := loader.ResponsiveSrcSet("/images/pic.webp", 1080)
	require.NoError(t, err)
	want := "/cdn/image/relative/16/images/pic.webp 16w, " +
		"/cdn/image/relative/32/images/pic.webp 32w, " +
		"/cdn/image/relative/48/images/pic.webp 48w, " +
		"/cdn/image/relative/64/images/pic.webp 64w, " +
		"/cdn/image/relative/96/images/pic.webp 96w, " +
		"/cdn/image/relative/128/images/pic.webp 128w, " +
		"/cdn/image/relative/256/images/pic.webp 256w, " +
		"/cdn/image/relative/384/images/pic.webp 384w, " +
		"/cdn/image/relative/450/images/pic.webp 450w, " +
		"/cdn/image/relative/530/images/pic.webp 530w, " +
		"/cdn/image/relative/640/images/pic.webp 640w, " +
		"/cdn/image/relative/750/images/pic.webp 750w, " +
		"/cdn/image/relative/828/images/pic.webp 828w, " +
		"/cdn/image/relative/1080/images/pic.webp 1080w"
	assert.Equal(t, want, got)
}

func TestLoaderResponsiveSrcSet_SmallWidthUsesAllowedSizes(t *testing.T) {
	t.Parallel()

	loader := New(true)
	got, err := loader.ResponsiveSrcSet("/images/pic.webp", 40)
	require.NoError(t, err)
	want := "/cdn/image/relative/16/images/pic.webp 16w, " +
		"/cdn/image/relative/32/images/pic.webp 32w, " +
		"/cdn/image/relative/48/images/pic.webp 48w"
	assert.Equal(t, want, got)
}

func TestLoaderResponsiveSrcSet_RoundsUpToAllowedTargetWidth(t *testing.T) {
	t.Parallel()

	loader := New(true)
	got, err := loader.ResponsiveSrcSet("/images/pic.webp", 768)
	require.NoError(t, err)
	want := "/cdn/image/relative/16/images/pic.webp 16w, " +
		"/cdn/image/relative/32/images/pic.webp 32w, " +
		"/cdn/image/relative/48/images/pic.webp 48w, " +
		"/cdn/image/relative/64/images/pic.webp 64w, " +
		"/cdn/image/relative/96/images/pic.webp 96w, " +
		"/cdn/image/relative/128/images/pic.webp 128w, " +
		"/cdn/image/relative/256/images/pic.webp 256w, " +
		"/cdn/image/relative/384/images/pic.webp 384w, " +
		"/cdn/image/relative/450/images/pic.webp 450w, " +
		"/cdn/image/relative/530/images/pic.webp 530w, " +
		"/cdn/image/relative/640/images/pic.webp 640w, " +
		"/cdn/image/relative/750/images/pic.webp 750w, " +
		"/cdn/image/relative/828/images/pic.webp 828w"
	assert.Equal(t, want, got)
}

func TestLoaderThumb_Uses1080AndScalesHeight(t *testing.T) {
	t.Parallel()

	loader := New(true)
	url, width, height := loader.Thumb("/images/hello.webp", 1200, 630)

	assert.Equal(t, "/cdn/image/relative/1080/images/hello.webp", url)
	assert.Equal(t, 1080, width)
	assert.Equal(t, 567, height)
}
