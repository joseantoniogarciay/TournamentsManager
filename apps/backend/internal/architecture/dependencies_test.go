// Package architecture verifies the dependency rules of the modular monolith.
package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/joseantoniogarciay/TournamentsManager/apps/backend"

func TestBusinessPackagesDoNotImportAdapters(t *testing.T) {
	root := backendRoot(t)
	for _, name := range []string{"access", "accounts", "federated", "legal", "notifications", "registration", "suggestions", "tournaments"} {
		assertNoImportPrefix(t, filepath.Join(root, "internal", name), modulePath+"/internal/adapters/")
	}
}

func TestHTTPAdapterDoesNotImportPostgres(t *testing.T) {
	root := backendRoot(t)
	assertNoImportPrefix(t, filepath.Join(root, "internal", "adapters", "http"), modulePath+"/internal/adapters/postgres")
}

func backendRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}

func assertNoImportPrefix(t *testing.T, directory, forbiddenPrefix string) {
	t.Helper()
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", directory, err)
	}
	for _, pkg := range packages {
		for filename, file := range pkg.Files {
			for _, importSpec := range file.Imports {
				path, err := strconv.Unquote(importSpec.Path.Value)
				if err != nil {
					t.Fatalf("decode import in %s: %v", filename, err)
				}
				if strings.HasPrefix(path, forbiddenPrefix) {
					t.Errorf("%s imports forbidden dependency %s", filename, path)
				}
			}
		}
	}
}
