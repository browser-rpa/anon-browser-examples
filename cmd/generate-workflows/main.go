// Generates public workflow fixtures from Go, without connecting to a client.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/browser-rpa/anon-browser-examples/scenarios"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	check := flag.Bool("check", false, "verify tracked JSON matches Go sources")
	out := flag.String("out", "workflows", "output directory")
	flag.Parse()
	all, err := scenarios.All()
	if err != nil {
		return err
	}
	if !*check {
		if err = os.MkdirAll(*out, 0755); err != nil {
			return err
		}
	}
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := json.MarshalIndent(all[name], "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		path := filepath.Join(*out, name)
		if *check {
			existing, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !bytes.Equal(existing, data) {
				return fmt.Errorf("%s is stale; run go run ./cmd/generate-workflows", path)
			}
		} else if err = os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	fmt.Printf("%d workflow examples verified/generated\n", len(all))
	return nil
}
