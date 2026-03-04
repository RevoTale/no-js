package imageloader

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

const thumbWidth = 1080
const markdownSizesValue = "(max-width: 660px) 100vw, 672px"

var cdnS3PathPattern = regexp.MustCompile(`((?:^|/)cdn/image/s3/)(\d+)(/)`)

var deviceSizes = []int{384, 450, 530, 640, 828, 1080, 1920}

type Loader struct {
	enabled bool
}

func New(enabled bool) Loader {
	return Loader{
		enabled: enabled,
	}
}

func (l Loader) Enabled() bool {
	return l.enabled
}

func (l Loader) URL(src string, width int) string {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" {
		return ""
	}
	if !l.enabled {
		return trimmed
	}

	encodedSrc := strings.ReplaceAll(trimmed, " ", "%20")
	targetWidth := normalizeWidth(width)

	if cdnS3PathPattern.MatchString(encodedSrc) {
		replacement := fmt.Sprintf("${1}%d${3}", targetWidth)
		return cdnS3PathPattern.ReplaceAllString(encodedSrc, replacement)
	}

	relativePath := strings.TrimLeft(encodedSrc, "/")
	return fmt.Sprintf("/cdn/image/relative/%d/%s", targetWidth, relativePath)
}

func (l Loader) ResponsiveSrcSet(src string, maxWidth int) (string,error) {
	if !l.enabled {
		return "", errors.New("loader not enabled")
	}

	widths, err := responsiveWidths(maxWidth)
	if nil != err {
		return "",err
	}
	if len(widths) == 0 {
		return src,nil
	}
	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		url := l.URL(src, width)
		if url == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %dw", url, width))
	}

	return strings.Join(parts, ", "), nil
}

func (l Loader) Thumb(src string, originalWidth int, originalHeight int) (string, int, int) {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" {
		return "", 0, 0
	}

	if !l.enabled {
		return trimmed, positiveOrZero(originalWidth), positiveOrZero(originalHeight)
	}

	outURL := l.URL(trimmed, thumbWidth)
	outHeight := 0
	if originalWidth > 0 && originalHeight > 0 {
		outHeight = int(math.Round(float64(thumbWidth*originalHeight) / float64(originalWidth)))
	}
	if outHeight < 0 {
		outHeight = 0
	}

	return outURL, thumbWidth, outHeight
}

func MarkdownSizes() string {
	return markdownSizesValue
}

func normalizeWidth(width int) int {
	if width > 0 {
		return width
	}
	return thumbWidth
}

func cutSmallerSizes(size int) []int  {
	out := make([]int, 0, len(deviceSizes))
	for _, ds := range deviceSizes {
		if ds <= size {
			out = append(out, ds)
		}
	}

	return out

}

func responsiveWidths(maxWidth int) ([]int,error) {
	if maxWidth < 0 {
		return nil, errors.New("width negative")
	}
	if maxWidth == 0 {
		return append([]int(nil), deviceSizes...), nil
	}
	out := cutSmallerSizes(maxWidth)
	if len(out) != 0 {
 		return out, nil
	}
	if len(deviceSizes) == 0 {
		return  nil, errors.New("device sizes is null")
	}
	return append([]int{}, deviceSizes[0]) , nil
}

func positiveOrZero(value int) int {
	if value > 0 {
		return value
	}
	return 0
}
