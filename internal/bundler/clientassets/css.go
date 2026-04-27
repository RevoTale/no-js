package clientassets

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

func transformCSSModule(
	layout projectlayout.ProjectLayout,
	filePath string,
	content string,
	classAllocator *generatedClassAllocator,
) ([]cssClass, string, error) {
	mapper := newCSSClassMapper(layout, filePath, classAllocator)
	parser := css.NewParser(parse.NewInputString(content), false)
	streamRewriter := newCSSTokenStreamRewriter(mapper)
	builder := strings.Builder{}
	for {
		grammar, tokenType, data := parser.Next()
		switch grammar {
		case css.ErrorGrammar:
			streamRewriter.flush(&builder)
			err := parser.Err()
			if errors.Is(err, io.EOF) {
				return mapper.classes(), builder.String(), nil
			}
			return nil, "", fmt.Errorf("parse css %q: %w", filePath, err)
		case css.TokenGrammar:
			streamRewriter.write(&builder, tokenType, data)
		case css.CommentGrammar:
			streamRewriter.flush(&builder)
			builder.Write(data)
		case css.AtRuleGrammar:
			streamRewriter.flush(&builder)
			builder.Write(data)
			writeCSSAtRuleTokens(&builder, data, parser.Values(), mapper)
			builder.WriteByte(';')
		case css.BeginAtRuleGrammar:
			streamRewriter.flush(&builder)
			builder.Write(data)
			writeCSSAtRuleTokens(&builder, data, parser.Values(), mapper)
			builder.WriteByte('{')
		case css.EndAtRuleGrammar:
			streamRewriter.flush(&builder)
			builder.WriteByte('}')
		case css.QualifiedRuleGrammar:
			streamRewriter.flush(&builder)
			builder.Write(data)
			writeCSSTokens(&builder, parser.Values(), mapper)
			builder.WriteByte(',')
		case css.BeginRulesetGrammar:
			streamRewriter.flush(&builder)
			writeCSSTokens(&builder, parser.Values(), mapper)
			builder.WriteByte('{')
		case css.EndRulesetGrammar:
			streamRewriter.flush(&builder)
			builder.WriteByte('}')
		case css.DeclarationGrammar, css.CustomPropertyGrammar:
			streamRewriter.flush(&builder)
			builder.Write(data)
			builder.WriteByte(':')
			writeRawCSSTokens(&builder, parser.Values())
			builder.WriteByte(';')
		}
	}
}

func writeCSSTokens(builder *strings.Builder, tokens []css.Token, mapper *cssClassMapper) {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if isCSSClassDot(token) && i+1 < len(tokens) && tokens[i+1].TokenType == css.IdentToken {
			class := mapper.classFor(string(tokens[i+1].Data))
			builder.WriteByte('.')
			builder.WriteString(class.Generated)
			i++
			continue
		}
		builder.Write(token.Data)
	}
}

func writeCSSAtRuleTokens(builder *strings.Builder, name []byte, tokens []css.Token, mapper *cssClassMapper) {
	if atRulePreludeCanContainSelectors(name) {
		writeCSSTokens(builder, tokens, mapper)
		return
	}
	writeRawCSSTokens(builder, tokens)
}

func atRulePreludeCanContainSelectors(name []byte) bool {
	switch strings.ToLower(strings.TrimSpace(string(name))) {
	case "@scope", "@supports":
		return true
	default:
		return false
	}
}

func writeRawCSSTokens(builder *strings.Builder, tokens []css.Token) {
	for _, token := range tokens {
		builder.Write(token.Data)
	}
}

func isCSSClassDot(token css.Token) bool {
	return token.TokenType == css.DelimToken && string(token.Data) == "."
}

func newCSSTokenStreamRewriter(mapper *cssClassMapper) *cssTokenStreamRewriter {
	return &cssTokenStreamRewriter{
		mapper: mapper,
	}
}

func (rewriter *cssTokenStreamRewriter) write(
	builder *strings.Builder,
	tokenType css.TokenType,
	data []byte,
) {
	if rewriter.pendingClassDot {
		if tokenType == css.IdentToken {
			class := rewriter.mapper.classFor(string(data))
			builder.WriteByte('.')
			builder.WriteString(class.Generated)
			rewriter.pendingClassDot = false
			return
		}
		builder.WriteByte('.')
		rewriter.pendingClassDot = false
	}
	if tokenType == css.DelimToken && string(data) == "." {
		rewriter.pendingClassDot = true
		return
	}
	builder.Write(data)
}

func (rewriter *cssTokenStreamRewriter) flush(builder *strings.Builder) {
	if !rewriter.pendingClassDot {
		return
	}
	builder.WriteByte('.')
	rewriter.pendingClassDot = false
}

func newGeneratedClassAllocator() *generatedClassAllocator {
	return &generatedClassAllocator{
		used: map[string]string{},
	}
}

func (allocator *generatedClassAllocator) className(
	layout projectlayout.ProjectLayout,
	filePath string,
	className string,
) string {
	if allocator == nil {
		allocator = newGeneratedClassAllocator()
	}
	key := generatedClassKey(layout, filePath, className)
	base := hashedGeneratedClassName(key)
	candidate := base
	for suffix := 2; ; suffix++ {
		owner, ok := allocator.used[candidate]
		if !ok || owner == key {
			allocator.used[candidate] = key
			return candidate
		}
		candidate = base + "_" + strconv.Itoa(suffix)
	}
}

func newCSSClassMapper(
	layout projectlayout.ProjectLayout,
	filePath string,
	allocator *generatedClassAllocator,
) *cssClassMapper {
	return &cssClassMapper{
		layout:     layout,
		filePath:   filePath,
		allocator:  allocator,
		filePrefix: pascalIdentifier(strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))),
		seen:       map[string]cssClass{},
		constants:  map[string]int{},
	}
}

func (mapper *cssClassMapper) classFor(className string) cssClass {
	if class, ok := mapper.seen[className]; ok {
		return class
	}
	constant := mapper.filePrefix + pascalIdentifier(className) + "Class"
	if count := mapper.constants[constant]; count > 0 {
		constant = constant + strconv.Itoa(count+1)
	}
	mapper.constants[constant]++
	class := cssClass{
		Original:  className,
		Constant:  constant,
		Generated: mapper.allocator.className(mapper.layout, mapper.filePath, className),
	}
	mapper.seen[className] = class
	return class
}

func (mapper *cssClassMapper) classes() []cssClass {
	classNames := make([]string, 0, len(mapper.seen))
	for className := range mapper.seen {
		classNames = append(classNames, className)
	}
	sort.Strings(classNames)
	classes := make([]cssClass, 0, len(classNames))
	for _, className := range classNames {
		classes = append(classes, mapper.seen[className])
	}
	return classes
}

func generatedClassKey(layout projectlayout.ProjectLayout, filePath string, className string) string {
	relative, err := filepath.Rel(layout.RootDir, filePath)
	if err != nil {
		relative = filePath
	}
	return filepath.ToSlash(relative) + "\x00" + className
}

func hashedGeneratedClassName(key string) string {
	hash := sha1.Sum([]byte(key))
	encoded := hex.EncodeToString(hash[:])
	return "n_" + encoded[:8]
}
