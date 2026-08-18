package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codegym/backend/internal/model"
	"codegym/backend/internal/seed"
)

// exportlib dumps the built-in question bank into JSON files (one per tech
// stack) under data/library, which becomes the source of truth at runtime.
func main() {
	outDir := "data/library"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	all, err := seed.All()
	if err != nil {
		panic(err)
	}
	byTech := map[string][]model.Exercise{}
	for _, ex := range all {
		byTech[ex.TechStack] = append(byTech[ex.TechStack], ex)
	}
	techs := make([]string, 0, len(byTech))
	for t := range byTech {
		techs = append(techs, t)
	}
	sort.Strings(techs)

	total := 0
	for _, t := range techs {
		list := byTech[t]
		// keep source order (seed.All is deterministic per bank)
		path := filepath.Join(outDir, t+".json")
		data, err := json.MarshalIndent(list, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			panic(err)
		}
		total += len(list)
		fmt.Printf("%s: %d 题 -> %s\n", t, len(list), path)
	}
	fmt.Printf("共导出 %d 题到 %s\n", total, outDir)
	_ = strings.TrimSpace
}
