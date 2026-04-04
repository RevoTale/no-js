package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

const MessagesDir = "messages"

type Message struct {
	ID          string       `json:"id"`
	Translation string       `json:"translation"`
	Args        []MessageArg `json:"args,omitempty"`
}

type MessageArg struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var exportedIdentifierPattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)
var templateActionPattern = regexp.MustCompile(`{{(.*?)}}`)
var templateFieldPattern = regexp.MustCompile(`^\s*\.[A-Za-z_][A-Za-z0-9_]*\s*$`)

func DiscoverMessageFiles(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, MessagesDir)
	if err != nil {
		return nil, fmt.Errorf("read messages directory %q: %w", MessagesDir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}
		if entry.IsDir() {
			return nil, fmt.Errorf("messages directory %q must not contain subdirectories: %q", MessagesDir, name)
		}
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			return nil, fmt.Errorf("messages directory %q must contain only json files: %q", MessagesDir, name)
		}
		if localeFromPath(name) == "" {
			return nil, fmt.Errorf("message file %q must end with .<locale>.json", name)
		}
		files = append(files, path.Join(MessagesDir, name))
	}

	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("messages directory %q has no json files", MessagesDir)
	}

	return files, nil
}

func ValidateMessageKeyParity(fsys fs.FS, files []string, expectedKeys []string) error {
	return ValidateMessageCatalog(fsys, files, "", expectedKeys)
}

func ValidateMessageCatalog(fsys fs.FS, files []string, canonicalRef string, expectedKeys []string) error {
	expectedKeySet := buildExpectedKeySet(expectedKeys)
	if len(expectedKeySet) == 0 {
		return fmt.Errorf("expected key set is empty")
	}

	localeMessages, canonicalLocale, err := LoadMessageDefinitions(fsys, files, canonicalRef)
	if err != nil {
		return err
	}

	canonicalMessages, ok := localeMessages[canonicalLocale]
	if !ok || len(canonicalMessages) == 0 {
		return fmt.Errorf("canonical locale %q has no messages", canonicalLocale)
	}

	canonicalByID := make(map[string]Message, len(canonicalMessages))
	for _, message := range canonicalMessages {
		canonicalByID[message.ID] = message
	}

	locales := make([]string, 0, len(localeMessages))
	for locale := range localeMessages {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	for _, locale := range locales {
		messages := localeMessages[locale]
		localeKeys := make(map[string]struct{}, len(messages))
		for _, message := range messages {
			localeKeys[message.ID] = struct{}{}
			canonicalMessage, ok := canonicalByID[message.ID]
			if !ok {
				continue
			}
			if err := validatePlaceholderParity(canonicalMessage, message, locale == canonicalLocale); err != nil {
				return fmt.Errorf("locale %q message %q: %w", locale, message.ID, err)
			}
		}

		missing := make([]string, 0, 4)
		for key := range expectedKeySet {
			if _, ok := localeKeys[key]; !ok {
				missing = append(missing, key)
			}
		}
		sort.Strings(missing)

		extra := make([]string, 0, 4)
		for key := range localeKeys {
			if _, ok := expectedKeySet[key]; !ok {
				extra = append(extra, key)
			}
		}
		sort.Strings(extra)

		if len(missing) > 0 || len(extra) > 0 {
			return fmt.Errorf("locale %q key parity mismatch: missing=%v extra=%v", locale, missing, extra)
		}
	}

	return nil
}

func LoadMessageDefinitions(fsys fs.FS, files []string, canonicalRef string) (map[string][]Message, string, error) {
	grouped, err := groupMessageFilesByLocale(files)
	if err != nil {
		return nil, "", err
	}

	canonicalLocale := resolveCanonicalLocale(canonicalRef, files)
	if canonicalLocale == "" {
		return nil, "", fmt.Errorf("canonical locale is required")
	}
	if _, ok := grouped[canonicalLocale]; !ok {
		return nil, "", fmt.Errorf("canonical locale %q is missing from message files", canonicalLocale)
	}

	out := make(map[string][]Message, len(grouped))
	for locale, localeFiles := range grouped {
		merged := make([]Message, 0)
		for _, file := range localeFiles {
			content, err := fs.ReadFile(fsys, file)
			if err != nil {
				return nil, "", fmt.Errorf("read locale file %q: %w", file, err)
			}

			var messages []Message
			if locale == canonicalLocale {
				messages, err = ParseCanonicalMessages(content)
			} else {
				messages, err = ParseLocaleMessages(content)
			}
			if err != nil {
				return nil, "", fmt.Errorf("parse locale file %q: %w", file, err)
			}

			merged, err = mergeMessages(merged, messages)
			if err != nil {
				return nil, "", fmt.Errorf("parse locale file %q: %w", file, err)
			}
		}
		out[locale] = merged
	}

	return out, canonicalLocale, nil
}

func ParseCanonicalMessages(data []byte) ([]Message, error) {
	entries, err := parseMessages(data, "canonical")
	if err != nil {
		return nil, err
	}

	for index, entry := range entries {
		placeholders, err := extractPlaceholders(entry.Translation)
		if err != nil {
			return nil, fmt.Errorf("entry %d (%q): %w", index, entry.ID, err)
		}
		if err := validateArgs(entry.Args); err != nil {
			return nil, fmt.Errorf("entry %d (%q): %w", index, entry.ID, err)
		}
		if err := validateArgsAgainstPlaceholders(entry.Args, placeholders); err != nil {
			return nil, fmt.Errorf("entry %d (%q): %w", index, entry.ID, err)
		}
	}

	return entries, nil
}

func ParseLocaleMessages(data []byte) ([]Message, error) {
	entries, err := parseMessages(data, "locale")
	if err != nil {
		return nil, err
	}
	for index, entry := range entries {
		if len(entry.Args) > 0 {
			return nil, fmt.Errorf("entry %d (%q) must not define args outside the canonical locale", index, entry.ID)
		}
		if _, err := extractPlaceholders(entry.Translation); err != nil {
			return nil, fmt.Errorf("entry %d (%q): %w", index, entry.ID, err)
		}
	}
	return entries, nil
}

func parseMessages(data []byte, label string) ([]Message, error) {
	var entries []Message
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse %s messages json: %w", label, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("%s messages are empty", label)
	}

	seenIDs := make(map[string]struct{}, len(entries))
	out := make([]Message, 0, len(entries))
	for index, entry := range entries {
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Translation = strings.TrimSpace(entry.Translation)
		if entry.ID == "" {
			return nil, fmt.Errorf("entry %d has empty id", index)
		}
		if _, exists := seenIDs[entry.ID]; exists {
			return nil, fmt.Errorf("duplicate message id %q", entry.ID)
		}
		seenIDs[entry.ID] = struct{}{}
		out = append(out, entry)
	}

	sort.Slice(out, func(i int, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func buildExpectedKeySet(keys []string) map[string]struct{} {
	out := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = struct{}{}
	}
	return out
}

func localeFromPath(pathValue string) string {
	fileName := strings.TrimSpace(path.Base(strings.TrimSpace(pathValue)))
	if !strings.HasSuffix(strings.ToLower(fileName), ".json") {
		return ""
	}

	baseName := strings.TrimSuffix(fileName, path.Ext(fileName))
	segments := strings.Split(baseName, ".")
	if len(segments) == 0 {
		return ""
	}

	locale := normalizeLocale(segments[len(segments)-1])
	if !localeCodePattern.MatchString(locale) {
		return ""
	}
	return locale
}

func groupMessageFilesByLocale(files []string) (map[string][]string, error) {
	grouped := make(map[string][]string)
	for _, file := range files {
		trimmed := strings.TrimSpace(file)
		if trimmed == "" {
			continue
		}
		locale := localeFromPath(trimmed)
		if locale == "" {
			return nil, fmt.Errorf("message file %q must end with .<locale>.json", trimmed)
		}
		grouped[locale] = append(grouped[locale], trimmed)
	}

	for locale := range grouped {
		sort.Strings(grouped[locale])
	}
	if len(grouped) == 0 {
		return nil, fmt.Errorf("message files are empty")
	}
	return grouped, nil
}

func resolveCanonicalLocale(canonicalRef string, files []string) string {
	if locale := localeFromPath(canonicalRef); locale != "" {
		return locale
	}

	normalizedRef := normalizeLocale(canonicalRef)
	if normalizedRef != "" {
		return normalizedRef
	}

	return detectCanonicalLocale(files)
}

func mergeMessages(base []Message, incoming []Message) ([]Message, error) {
	merged := make(map[string]Message, len(base)+len(incoming))
	for _, message := range base {
		merged[message.ID] = message
	}
	for _, message := range incoming {
		if _, exists := merged[message.ID]; exists {
			return nil, fmt.Errorf("duplicate message id %q", message.ID)
		}
		merged[message.ID] = message
	}

	out := make([]Message, 0, len(merged))
	for _, message := range merged {
		out = append(out, message)
	}
	sort.Slice(out, func(i int, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func validateArgs(args []MessageArg) error {
	seen := make(map[string]struct{}, len(args))
	for _, arg := range args {
		name := strings.TrimSpace(arg.Name)
		if !exportedIdentifierPattern.MatchString(name) {
			return fmt.Errorf("invalid arg name %q", arg.Name)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("duplicate arg name %q", name)
		}
		seen[name] = struct{}{}

		typeName := strings.TrimSpace(arg.Type)
		if !isSupportedArgType(typeName) {
			return fmt.Errorf("unsupported arg type %q", arg.Type)
		}
	}
	return nil
}

func validateArgsAgainstPlaceholders(args []MessageArg, placeholders []string) error {
	argNames := make(map[string]struct{}, len(args))
	for _, arg := range args {
		argNames[strings.TrimSpace(arg.Name)] = struct{}{}
	}

	for _, placeholder := range placeholders {
		if _, ok := argNames[placeholder]; !ok {
			return fmt.Errorf("undeclared placeholder %q", placeholder)
		}
	}
	for name := range argNames {
		if !slicesContains(placeholders, name) {
			return fmt.Errorf("declared arg %q is unused", name)
		}
	}
	return nil
}

func validatePlaceholderParity(canonical Message, localized Message, isCanonical bool) error {
	canonicalPlaceholders, err := extractPlaceholders(canonical.Translation)
	if err != nil {
		return err
	}
	localizedPlaceholders, err := extractPlaceholders(localized.Translation)
	if err != nil {
		return err
	}
	if len(canonicalPlaceholders) != len(localizedPlaceholders) {
		return fmt.Errorf("placeholder mismatch: expected=%v actual=%v", canonicalPlaceholders, localizedPlaceholders)
	}
	for index := range canonicalPlaceholders {
		if canonicalPlaceholders[index] != localizedPlaceholders[index] {
			return fmt.Errorf("placeholder mismatch: expected=%v actual=%v", canonicalPlaceholders, localizedPlaceholders)
		}
	}
	if isCanonical {
		return validateArgsAgainstPlaceholders(canonical.Args, canonicalPlaceholders)
	}
	return nil
}

func extractPlaceholders(translation string) ([]string, error) {
	matches := templateActionPattern.FindAllStringSubmatch(translation, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(matches))
	for _, match := range matches {
		action := strings.TrimSpace(match[1])
		if !templateFieldPattern.MatchString(action) {
			return nil, fmt.Errorf("unsupported template action %q", action)
		}
		placeholders = append(placeholders, strings.TrimPrefix(strings.TrimSpace(action), "."))
	}
	sort.Strings(placeholders)
	return placeholders, nil
}

func isSupportedArgType(typeName string) bool {
	switch strings.TrimSpace(typeName) {
	case "string", "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"time.Time":
		return true
	default:
		return false
	}
}

func detectCanonicalLocale(files []string) string {
	locales := make([]string, 0, len(files))
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		locale := localeFromPath(file)
		if locale == "" {
			continue
		}
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	for _, locale := range locales {
		if locale == "en" {
			return locale
		}
	}
	if len(locales) == 0 {
		return ""
	}
	return locales[0]
}

func slicesContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
