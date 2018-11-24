package main

import (
	"github.com/lyft/protoc-gen-star"
)

func main() {
	pgs.Init().RegisterModule(&rightsGen{ModuleBase: pgs.ModuleBase{}}).Render()
}
