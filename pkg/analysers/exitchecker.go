package analysers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	specPackage      = "main"
	specImport       = "\"os\""
	specDeclaration  = "main"
	specFunction     = "Exit"
	specFunctionPrfx = "os"
)

var ExitCheckAnalyser = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for calling os.Exit()",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		haveSpecImport := false
		ast.Inspect(file, func(node ast.Node) bool {
			switch nodeType := node.(type) {
			case *ast.File:
				if nodeType.Name.Name != specPackage {
					return false
				}
			case *ast.ImportSpec:
				if nodeType.Path.Value == specImport {
					haveSpecImport = true
				}
			case *ast.FuncDecl:
				if !haveSpecImport {
					return false
				}
				if nodeType.Name.Name != specDeclaration {
					return false
				}

				return true
			case *ast.CallExpr:
				f, ok := nodeType.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkg, ok := f.X.(*ast.Ident)
				if !ok {
					return true
				}
				if pkg.Name == specFunctionPrfx && f.Sel.Name == specFunction {
					pass.Reportf(pkg.NamePos, "calling os.Exit in function main")
					return false
				}
			}
			return true
		})
	}
	//nolint:nilnil // необходимо возвращать так: Result должен иметь тип указателя
	return nil, nil
}
