package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Ref: https://github.com/googleapis/google-cloud-go/blob/4b41a6f3b0e014221ff06595fd24fd7efb7d765a/internal/actions/cmd/changefinder/main.go#L142
func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}
	log.Printf("Root dir: %q", rootDir)

	changes, err := gitFilesChanges(rootDir)
	if err != nil {
		log.Fatalf("unable to get files changed: %v", err)
	}

	updatedSubmodulesSet := map[string]bool{}

	gex := regexp.MustCompile(`(services)\/([a-z-]+)\/.+`)
	for _, change := range changes {
		match := gex.MatchString(change)

		if !match {
			continue
		}
		log.Printf("%+v", match)
		pkg := strings.Split(change, "/")[1]
		log.Printf("update to path: %s", pkg)
		if updatedSubmodulesSet[pkg] {
			continue
		}
		updatedSubmodulesSet[pkg] = true
	}
	updatedSubmodule := []string{}
	for mod := range updatedSubmodulesSet {
		updatedSubmodule = append(updatedSubmodule, mod)
	}
	b, err := json.Marshal(updatedSubmodule)
	if err != nil {
		log.Fatalf("unable to marshal submodules: %v", err)
	}
	fmt.Printf("::set-output name=submodules::%s", b)
}

func gitFilesChanges(dir string) ([]string, error) {
	c := exec.Command("git", "diff", "--name-only", "origin/main")
	c.Dir = dir
	b, err := c.Output()
	if err != nil {
		return nil, err
	}
	b = bytes.TrimSpace(b)
	log.Printf("Files changed:\n%s", b)
	return strings.Split(string(b), "\n"), nil
}
