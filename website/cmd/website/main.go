package main

//go:generate go run scripts/build_docs.go

import (
	"github.com/N3moAhead/bombahead/website/internal/cfg"
	"github.com/N3moAhead/bombahead/website/internal/db"
	"github.com/N3moAhead/bombahead/website/internal/router"
)

func main() {
	cfg := cfg.Load()
	db.Init(cfg)
	router.Start(cfg)
}
