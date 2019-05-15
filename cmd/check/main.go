package main

import (
	"github.com/matloob/analysistalk/cmpequal"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(cmpequal.Analyzer)
}