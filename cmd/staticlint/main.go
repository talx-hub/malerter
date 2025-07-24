// Package main запускает набор статических анализаторов Go-кода, включая стандартные,
// сторонние (staticcheck, go-critic, errcheck) и пользовательские.
//
// Он предназначен для интеграции с `go vet` через флаг `-vettool`.
package main

import (
	"github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/talx-hub/malerter/pkg/analysers"
)

// main инициализирует и запускает набор статических анализаторов.
// Используется как точка входа для multichecker.
// Включает:
//   - пользовательские анализаторы (analysers.ExitCheckAnalyser),
//   - стандартные анализаторы (printf, shadow и др.),
//   - сторонние анализаторы (staticcheck, simple, stylecheck, quickfix),
//   - go-critic (go-critic/analyzer),
//   - errcheck.
//
// Пример использования:
//
//	go build -o ./bin/ma-linter ./cmd/staticlint
//	go vet -vettool=./bin/ma-linter ./...
func main() {
	// checks определяет включаемые по имени анализаторы из пакетов staticcheck, simple, stylecheck, quickfix.
	// Только анализаторы, чьи имена присутствуют в этом map'е, будут добавлены в финальный список.
	checks := map[string]bool{
		"S1000":  true,
		"ST1000": true,
		"QF1001": true,
	}

	mychecks := []*analysis.Analyzer{
		analysers.ExitCheckAnalyser,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		analyzer.Analyzer,
		errcheck.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range quickfix.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(mychecks...)
}
