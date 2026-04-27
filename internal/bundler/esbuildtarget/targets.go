package esbuildtarget

import (
	"fmt"
	"sort"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

const DefaultTarget = "es2020"

func Parse(values []string) (api.Target, []api.Engine, error) {
	if len(values) == 0 {
		values = []string{DefaultTarget}
	}

	target := api.DefaultTarget
	engines := []api.Engine{}
	for _, value := range values {
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized == "" {
			return 0, nil, fmt.Errorf("browser target cannot be empty")
		}
		if parsed, ok := targetByName[normalized]; ok {
			target = parsed
			continue
		}
		matchedEngine := false
		for prefix, engine := range engineByName {
			if !strings.HasPrefix(normalized, prefix) {
				continue
			}
			version := strings.TrimSpace(normalized[len(prefix):])
			if version == "" {
				return 0, nil, fmt.Errorf("browser target %q is missing a version", value)
			}
			engines = append(engines, api.Engine{Name: engine, Version: version})
			matchedEngine = true
			break
		}
		if matchedEngine {
			continue
		}
		return 0, nil, fmt.Errorf("unsupported browser target %q; use %s", value, supportedTargetSummary())
	}
	return target, engines, nil
}

var targetByName = map[string]api.Target{
	"esnext": api.ESNext,
	"es5":    api.ES5,
	"es6":    api.ES2015,
	"es2015": api.ES2015,
	"es2016": api.ES2016,
	"es2017": api.ES2017,
	"es2018": api.ES2018,
	"es2019": api.ES2019,
	"es2020": api.ES2020,
	"es2021": api.ES2021,
	"es2022": api.ES2022,
	"es2023": api.ES2023,
	"es2024": api.ES2024,
	"es2025": api.ES2025,
}

var engineByName = map[string]api.EngineName{
	"chrome":  api.EngineChrome,
	"deno":    api.EngineDeno,
	"edge":    api.EngineEdge,
	"firefox": api.EngineFirefox,
	"hermes":  api.EngineHermes,
	"ie":      api.EngineIE,
	"ios":     api.EngineIOS,
	"node":    api.EngineNode,
	"opera":   api.EngineOpera,
	"rhino":   api.EngineRhino,
	"safari":  api.EngineSafari,
}

func supportedTargetSummary() string {
	values := make([]string, 0, len(targetByName)+len(engineByName))
	for target := range targetByName {
		values = append(values, target)
	}
	for engine := range engineByName {
		values = append(values, engine+"N")
	}
	sort.Strings(values)
	return strings.Join(values, ", ")
}
