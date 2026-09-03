// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
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
	coreRevision      = "97f142e67329526f59529db09f58010da72f8c70"
	defaultOutput     = "pkg/games/system_metadata.gen.json"
	coreSourceEnv     = "ZAPAROO_CORE_SOURCE"
	coreRepositoryEnv = "ZAPAROO_CORE_REPOSITORY"
	outputFormat      = 1
)

type generatedMetadata struct {
	Systems map[string]systemMetadata `json:"systems"`
	Source  string                    `json:"source"`
	Format  int                       `json:"format"`
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
	// #nosec G304,G703 -- name is an internal path within explicitly selected Core source.
	data, err := os.ReadFile(filepath.Join(s.root, filepath.FromSlash(name)))
	if err != nil {
		return nil, fmt.Errorf("read Core source file: %w", err)
	}
	return data, nil
}

func gitOutput(args ...string) (string, error) {
	// #nosec G204,G702 -- arguments are fixed generator operations plus configured repository path.
	command := exec.CommandContext(context.Background(), "git", args...)
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
	cachedRevision, revisionErr := gitOutput("-C", checkout, "rev-parse", "HEAD")
	if revisionErr == nil && cachedRevision == coreRevision {
		return checkout, nil
	}
	if mkdirErr := os.MkdirAll(cacheParent, 0o750); mkdirErr != nil {
		return "", fmt.Errorf("create metadata cache: %w", mkdirErr)
	}
	temporary, err := os.MkdirTemp(cacheParent, "checkout-")
	if err != nil {
		return "", fmt.Errorf("create temporary Core checkout: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(temporary); removeErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, removeErr)
		}
	}()

	repository := os.Getenv(coreRepositoryEnv)
	if repository == "" {
		repository = coreRepository
	}
	if _, initErr := gitOutput("init", "--quiet", temporary); initErr != nil {
		return "", initErr
	}
	if _, remoteErr := gitOutput("-C", temporary, "remote", "add", "origin", repository); remoteErr != nil {
		return "", remoteErr
	}
	if _, fetchErr := gitOutput(
		"-C", temporary, "fetch", "--quiet", "--depth=1", "origin", coreRevision,
	); fetchErr != nil {
		return "", fetchErr
	}
	if _, checkoutErr := gitOutput(
		"-C", temporary, "checkout", "--quiet", "--detach", "FETCH_HEAD",
	); checkoutErr != nil {
		return "", checkoutErr
	}
	revision, err := gitOutput("-C", temporary, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	if revision != coreRevision {
		return "", fmt.Errorf("core source revision mismatch: got %s", revision)
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

func metadataForSystem(
	source sourceReader,
	system *catalog.Core,
	aliases map[string][]string,
) (systemMetadata, error) {
	path := "pkg/assets/systems/" + system.ID + ".json"
	data, err := source.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		// Display assets are optional. Keep Core's operational catalog independent
		// from metadata needed only by this legacy application.
		return systemMetadata{
			Name:    system.ID,
			Aliases: append([]string(nil), aliases[system.ID]...),
		}, nil
	}
	if err != nil {
		return systemMetadata{}, fmt.Errorf("read metadata for %s: %w", system.ID, err)
	}

	var entry systemMetadata
	if err := json.Unmarshal(data, &entry); err != nil {
		return systemMetadata{}, fmt.Errorf("decode metadata for %s: %w", system.ID, err)
	}
	if entry.Name == "" || entry.Category == "" {
		return systemMetadata{}, fmt.Errorf("metadata for %s lacks name or category", system.ID)
	}
	entry.Aliases = append([]string(nil), aliases[system.ID]...)
	return entry, nil
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
	systems := catalog.All()
	for i := range systems {
		system := &systems[i]
		entry, err := metadataForSystem(source, system, aliases)
		if err != nil {
			return nil, err
		}
		metadata[system.ID] = entry
	}
	return metadata, nil
}

func pinnedSourceLabel() string {
	return coreModule + "@" + coreRevision
}

func generatedOutputCurrent(path string) bool {
	// #nosec G304 -- path is explicit generator output.
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
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		source = directorySource{root: checkout}
		sourceLabel = pinnedSourceLabel()
	}

	generated, err := generate(source, sourceLabel)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if current, err := os.ReadFile(*output); err == nil && bytes.Equal(current, generated) {
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o750); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// #nosec G306,G703 -- generated metadata is a world-readable build input at explicit output path.
	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
