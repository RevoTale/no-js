package i18n

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type CompiledMessagePart struct {
	Text string
	Arg  string
}

type CompiledMessage struct {
	Parts []CompiledMessagePart
}

type Bundle[K ~string] struct {
	runtime *Runtime[K]
}

func NewBundle[K ~string](cfg Config, catalog *Catalog, defaultMessages map[K]string) (*Bundle[K], error) {
	runtime, err := NewRuntime(cfg, catalog, defaultMessages)
	if err != nil {
		return nil, err
	}
	return &Bundle[K]{runtime: runtime}, nil
}

func NewStaticBundle[K ~string](
	cfg Config,
	compiledMessages map[string]map[K]CompiledMessage,
	defaultMessages map[K]string,
) (*Bundle[K], error) {
	runtime, err := NewStaticRuntime(cfg, compiledMessages, defaultMessages)
	if err != nil {
		return nil, err
	}
	return &Bundle[K]{runtime: runtime}, nil
}

func (bundle *Bundle[K]) Config() Config {
	if bundle == nil || bundle.runtime == nil {
		return Config{}
	}
	return bundle.runtime.Config()
}

func (bundle *Bundle[K]) Localize(locale string, key K, vars map[string]any) string {
	if bundle == nil || bundle.runtime == nil {
		return strings.TrimSpace(string(key))
	}
	return bundle.runtime.Localize(locale, key, vars)
}

func (bundle *Bundle[K]) Context(r *http.Request, root *url.URL) Context[K] {
	if bundle == nil || bundle.runtime == nil {
		return nil
	}
	return bundle.runtime.Context(r, root)
}

func CompileMessage(translation string) (CompiledMessage, error) {
	parts, err := compileMessageParts(strings.TrimSpace(translation))
	if err != nil {
		return CompiledMessage{}, err
	}
	return CompiledMessage{Parts: parts}, nil
}

func (message CompiledMessage) Render(vars map[string]any) string {
	if len(message.Parts) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, part := range message.Parts {
		if part.Arg != "" {
			if vars == nil {
				continue
			}
			value, ok := vars[part.Arg]
			if !ok || value == nil {
				continue
			}
			_, _ = fmt.Fprint(&builder, value)
			continue
		}
		builder.WriteString(part.Text)
	}

	return builder.String()
}

func compileMessageParts(translation string) ([]CompiledMessagePart, error) {
	matches := templateActionPattern.FindAllStringSubmatchIndex(translation, -1)
	if len(matches) == 0 {
		return []CompiledMessagePart{{Text: translation}}, nil
	}

	parts := make([]CompiledMessagePart, 0, len(matches)*2+1)
	cursor := 0
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		fullStart, fullEnd := match[0], match[1]
		actionStart, actionEnd := match[2], match[3]
		if fullStart > cursor {
			parts = append(parts, CompiledMessagePart{Text: translation[cursor:fullStart]})
		}

		action := strings.TrimSpace(translation[actionStart:actionEnd])
		if !templateFieldPattern.MatchString(action) {
			return nil, fmt.Errorf("unsupported template action %q", action)
		}

		parts = append(parts, CompiledMessagePart{Arg: strings.TrimPrefix(action, ".")})
		cursor = fullEnd
	}

	if cursor < len(translation) {
		parts = append(parts, CompiledMessagePart{Text: translation[cursor:]})
	}
	if len(parts) == 0 {
		parts = append(parts, CompiledMessagePart{Text: translation})
	}

	return parts, nil
}
