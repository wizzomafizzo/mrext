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

package games

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
)

// ReadNeoGeoNames reads MiSTer's romsets.xml aliases and display titles.
// Malformed XML never returns a partial mapping. Empty aliases and titles are
// ignored, preserving Favorites' naming rules.
func ReadNeoGeoNames(reader io.Reader) (map[string]string, error) {
	names := make(map[string]string)
	decoder := xml.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return names, nil
		}
		if err != nil {
			return nil, fmt.Errorf("decode NeoGeo ROM sets: %w", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "romset" {
			continue
		}
		var rawNames, title string
		for _, attribute := range start.Attr {
			switch attribute.Name.Local {
			case "name":
				rawNames = attribute.Value
			case "altname":
				title = strings.TrimSpace(attribute.Value)
			}
		}
		if rawNames == "" || title == "" {
			continue
		}
		for _, rawName := range strings.Split(rawNames, ",") {
			name := strings.ToLower(strings.TrimSpace(rawName))
			if name != "" {
				names[name] = title
			}
		}
	}
}

// NeoGeoMGLOverride mounts a NeoGeo ROM set using the catalog's .neo slot.
// The caller supplies the MGL-relative path and is responsible for identifying
// directory/ZIP sets. Unlike runtime system hooks, this performs no file writes.
func NeoGeoMGLOverride(system *System, mountPath string) (string, error) {
	if system == nil || system.Id != "NeoGeo" {
		return "", errors.New("NeoGeo MGL override requires a NeoGeo system")
	}
	core := CatalogCore(system)
	// ZIPs and extracted sets use the same slot as .neo, but are not scan
	// extensions in the pinned catalog. Do not duplicate its mount parameters.
	params, err := catalog.PathToMGLDef(&core, ".neo")
	if err != nil {
		return "", fmt.Errorf("resolve NeoGeo ROM-set slot: %w", err)
	}
	escaped := strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;",
	).Replace(mountPath)
	//nolint:gocritic // XML attributes need XML escaping, not Go string quoting.
	override := fmt.Sprintf("\t<file delay=\"%d\" type=%q index=\"%d\" path=\"%s\"/>\n",
		params.Delay, params.Method, params.Index, escaped)
	if params.ResetDelay > 0 {
		override += fmt.Sprintf("\t<reset delay=\"%d\" hold=\"%d\"/>\n", params.ResetDelay, params.ResetHold)
	}
	return override, nil
}
