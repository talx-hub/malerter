package analysers_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/talx-hub/malerter/pkg/analysers"
)

func TestExitCheckAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "exitcheck")
}

func TestExitCheckAnalyzerNotMainFunc(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "foofunc")
}

func TestExitCheckAnalyzerNotMainPackage(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analysers.ExitCheckAnalyser, "otherpackage")
}
