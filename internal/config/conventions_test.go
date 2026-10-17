package config

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// Settings must not drift from where they're documented and deployed. This
// reads config.go itself: os.Getenv("X") is a required setting, getenv("X", ...)
// one with a default.
func TestSettingsAreDocumentedAndDeployed(t *testing.T) {
	src := read(t, "config.go")
	readme := read(t, "../../README.md")
	deployment := read(t, "../../deploy/base/deployment.yaml")

	required := regexp.MustCompile(`os\.Getenv\("(\w+)"\)`).FindAllStringSubmatch(src, -1)
	optional := regexp.MustCompile(`getenv\("(\w+)",`).FindAllStringSubmatch(src, -1)
	if len(required)+len(optional) == 0 {
		t.Fatal("found no settings in config.go; has the way it reads them changed?")
	}

	for _, m := range append(required, optional...) {
		if !strings.Contains(readme, "`"+m[1]+"`") {
			t.Errorf("%s is missing from README.md's Configuration table", m[1])
		}
	}
	for _, m := range required {
		if !strings.Contains(deployment, "name: "+m[1]) {
			t.Errorf("%s is required but deploy/base/deployment.yaml doesn't set it", m[1])
		}
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
