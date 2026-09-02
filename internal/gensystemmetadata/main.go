package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
)

const (
	coreModule        = "github.com/ZaparooProject/zaparoo-core/v2"
	coreRepository    = "https://github.com/ZaparooProject/zaparoo-core.git"
	coreRevision      = "01d7a1798ab4f10fdb32fa6c02a6c3d723a15cd4"
	defaultOutput     = "pkg/games/system_metadata.gen.json"
	coreSourceEnv     = "ZAPAROO_CORE_SOURCE"
	coreRepositoryEnv = "ZAPAROO_CORE_REPOSITORY"
	outputFormat      = 1
)

type generatedMetadata struct {
	Source  string                    `json:"source"`
	Format  int                       `json:"format"`
	Systems map[string]systemMetadata `json:"systems"`
}

type systemMetadata struct {
	Name         string   `json:"name"`
	Category     string   `json:"category,omitempty"`
	ReleaseDate  string   `json:"releaseDate,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
}

type sourceReader interface {
	ReadFile(name string) ([]byte, error)
}

type directorySource struct {
	root string
}

func (s directorySource) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.root, filepath.FromSlash(name)))
}

func gitOutput(args ...string) (string, error) {
	command := exec.Command("git", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func checkoutCoreSource() (string, error) {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache: %w", err)
	}
	cacheParent := filepath.Join(cacheRoot, "mrext", "system-metadata")
	checkout := filepath.Join(cacheParent, coreRevision)
	if revision, revisionErr := gitOutput("-C", checkout, "rev-parse", "HEAD"); revisionErr == nil && revision == coreRevision {
		return checkout, nil
	}
	if err := os.MkdirAll(cacheParent, 0o755); err != nil {
		return "", fmt.Errorf("create metadata cache: %w", err)
	}
	temporary, err := os.MkdirTemp(cacheParent, "checkout-")
	if err != nil {
		return "", fmt.Errorf("create temporary Core checkout: %w", err)
	}
	defer os.RemoveAll(temporary)

	repository := os.Getenv(coreRepositoryEnv)
	if repository == "" {
		repository = coreRepository
	}
	if _, err := gitOutput("init", "--quiet", temporary); err != nil {
		return "", err
	}
	if _, err := gitOutput("-C", temporary, "remote", "add", "origin", repository); err != nil {
		return "", err
	}
	if _, err := gitOutput("-C", temporary, "fetch", "--quiet", "--depth=1", "origin", coreRevision); err != nil {
		return "", err
	}
	if _, err := gitOutput("-C", temporary, "checkout", "--quiet", "--detach", "FETCH_HEAD"); err != nil {
		return "", err
	}
	revision, err := gitOutput("-C", temporary, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	if revision != coreRevision {
		return "", fmt.Errorf("Core source revision mismatch: got %s", revision)
	}
	if err := os.RemoveAll(checkout); err != nil {
		return "", fmt.Errorf("remove stale Core checkout: %w", err)
	}
	if err := os.Rename(temporary, checkout); err != nil {
		return "", fmt.Errorf("cache Core checkout: %w", err)
	}
	return checkout, nil
}

func stringValue(expr ast.Expr, constants map[string]string) (string, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		if value.Kind != token.STRING {
			return "", false
		}
		decoded, err := strconv.Unquote(value.Value)
		return decoded, err == nil
	case *ast.Ident:
		decoded, ok := constants[value.Name]
		return decoded, ok
	default:
		return "", false
	}
}

func parseAliases(data []byte) (map[string][]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "systemdefs.go", data, 0)
	if err != nil {
		return nil, fmt.Errorf("parse Core system definitions: %w", err)
	}

	constants := make(map[string]string)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, spec := range general.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range values.Names {
				if i >= len(values.Values) {
					continue
				}
				if value, ok := stringValue(values.Values[i], constants); ok {
					constants[name.Name] = value
				}
			}
		}
	}

	aliases := make(map[string][]string)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != 1 || values.Names[0].Name != "Systems" || len(values.Values) != 1 {
				continue
			}
			systems, ok := values.Values[0].(*ast.CompositeLit)
			if !ok {
				continue
			}
			for _, element := range systems.Elts {
				entry, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				id, ok := stringValue(entry.Key, constants)
				if !ok {
					continue
				}
				definition, ok := entry.Value.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, field := range definition.Elts {
					pair, ok := field.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					name, ok := pair.Key.(*ast.Ident)
					if !ok || name.Name != "Aliases" {
						continue
					}
					list, ok := pair.Value.(*ast.CompositeLit)
					if !ok {
						continue
					}
					for _, item := range list.Elts {
						if alias, ok := stringValue(item, constants); ok {
							aliases[id] = append(aliases[id], alias)
						}
					}
				}
			}
		}
	}
	return aliases, nil
}

func loadMetadata(source sourceReader) (map[string]systemMetadata, error) {
	definitions, err := source.ReadFile("pkg/database/systemdefs/systemdefs.go")
	if err != nil {
		return nil, fmt.Errorf("read Core system definitions: %w", err)
	}
	aliases, err := parseAliases(definitions)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]systemMetadata)
	for _, system := range catalog.All() {
		data, err := source.ReadFile("pkg/assets/systems/" + system.ID + ".json")
		if err != nil {
			return nil, fmt.Errorf("read metadata for %s: %w", system.ID, err)
		}
		var entry systemMetadata
		if err := json.Unmarshal(data, &entry); err != nil {
			return nil, fmt.Errorf("decode metadata for %s: %w", system.ID, err)
		}
		if entry.Name == "" || entry.Category == "" {
			return nil, fmt.Errorf("metadata for %s lacks name or category", system.ID)
		}
		entry.Aliases = append([]string(nil), aliases[system.ID]...)
		metadata[system.ID] = entry
	}
	return metadata, nil
}

func pinnedSourceLabel() string {
	return coreModule + "@" + coreRevision
}

func generatedOutputCurrent(path string) bool {
	current, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var generated generatedMetadata
	if err := json.Unmarshal(current, &generated); err != nil {
		return false
	}
	return generated.Source == pinnedSourceLabel() && generated.Format == outputFormat
}

func generate(source sourceReader, sourceLabel string) ([]byte, error) {
	metadata, err := loadMetadata(source)
	if err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(generatedMetadata{
		Source:  sourceLabel,
		Format:  outputFormat,
		Systems: metadata,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode generated metadata: %w", err)
	}
	return append(encoded, '\n'), nil
}

func main() {
	output := flag.String("output", defaultOutput, "generated Go output path")
	flag.Parse()

	var source sourceReader
	var sourceLabel string
	if sourceRoot := os.Getenv(coreSourceEnv); sourceRoot != "" {
		source = directorySource{root: sourceRoot}
		sourceLabel = "local " + sourceRoot
	} else {
		if generatedOutputCurrent(*output) {
			return
		}
		checkout, err := checkoutCoreSource()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		source = directorySource{root: checkout}
		sourceLabel = pinnedSourceLabel()
	}

	generated, err := generate(source, sourceLabel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if current, err := os.ReadFile(*output); err == nil && bytes.Equal(current, generated) {
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
