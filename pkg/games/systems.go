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
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
)

const (
	CategoryArcade              = "Arcade"
	CategoryConsole             = "Console"
	CategoryComputer            = "Computer"
	CategoryHandheld            = "Handheld"
	CategoryOther               = "Other"
	ManufacturerEntex           = "Entex"
	ManufacturerEmerson         = "Emerson"
	ManufacturerMattel          = "Mattel"
	ManufacturerBally           = "Bally"
	ManufacturerAtari           = "Atari"
	ManufacturerColeco          = "Coleco"
	ManufacturerSega            = "Sega"
	ManufacturerNintendo        = "Nintendo"
	ManufacturerNEC             = "NEC"
	ManufacturerSNK             = "SNK"
	ManufacturerBandai          = "Bandai"
	ManufacturerVTech           = "VTech"
	ManufacturerCasio           = "Casio"
	ManufacturerWatara          = "Watara"
	ManufacturerMagnavox        = "Magnavox"
	ManufacturerFairchild       = "Fairchild"
	ManufacturerGCE             = "GCE"
	ManufacturerBitCorp         = "Bit Corporation"
	ManufacturerCommodore       = "Commodore"
	ManufacturerAmstrad         = "Amstrad"
	ManufacturerAcorn           = "Acorn"
	ManufacturerApple           = "Apple"
	ManufacturerBenesse         = "Benesse"
	ManufacturerSony            = "Sony"
	ManufacturerInterton        = "Interton"
	ManufacturerTandy           = "Tandy"
	ManufacturerIBM             = "IBM"
	ManufacturerApogee          = "Apogee"
	ManufacturerElektronika     = "Elektronika"
	ManufacturerCambridge       = "Cambridge"
	ManufacturerInteract        = "Interact"
	ManufacturerJupiter         = "Jupiter"
	ManufacturerVideoTechnology = "Video Technology"
	ManufacturerMicrosoft       = "Microsoft"
	ManufacturerPhilips         = "Philips"
	ManufacturerPanasonic       = "Panasonic"
)

//nolint:govet // Field order preserves legacy JSON output.
type MglParams struct {
	Delay      int
	Method     string
	Index      int
	resetDelay int
	resetHold  int
}

//nolint:govet // Field order preserves legacy JSON output.
type Slot struct {
	Label string
	Exts  []string
	Mgl   *MglParams
}

//nolint:govet // Field order preserves legacy JSON output.
type System struct {
	Id             string //nolint:revive // Legacy Remote JSON field name.
	Name           string // US
	Category       string
	ReleaseDate    string // US
	Manufacturer   string
	Alias          []string
	SetName        string
	SetNameSameDir bool
	Folder         []string
	Rbf            string //nolint:revive // Legacy Remote JSON field name.
	Slots          []Slot
	extensions     []string
}

type systemMetadata struct {
	Name         string   `json:"name"`
	Category     string   `json:"category,omitempty"`
	ReleaseDate  string   `json:"releaseDate,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
}

//go:embed system_metadata.gen.json
var rawSystemMetadata []byte

var systemMetadataByID = loadSystemMetadata()

func loadSystemMetadata() map[string]systemMetadata {
	var generated struct {
		Systems map[string]systemMetadata `json:"systems"`
	}
	if err := json.Unmarshal(rawSystemMetadata, &generated); err != nil {
		panic(fmt.Errorf("decode system display metadata: %w", err))
	}
	return generated.Systems
}

func slotsFromCatalog(slots []catalog.Slot) []Slot {
	converted := make([]Slot, len(slots))
	for i := range slots {
		converted[i] = Slot{
			Label: slots[i].Label,
			Exts:  append([]string(nil), slots[i].Exts...),
		}
		if slots[i].Mgl != nil {
			converted[i].Mgl = &MglParams{
				Delay:      slots[i].Mgl.Delay,
				Method:     slots[i].Mgl.Method,
				Index:      slots[i].Mgl.Index,
				resetDelay: slots[i].Mgl.ResetDelay,
				resetHold:  slots[i].Mgl.ResetHold,
			}
		}
	}
	return converted
}

func slotsToCatalog(slots []Slot) []catalog.Slot {
	converted := make([]catalog.Slot, len(slots))
	for i := range slots {
		converted[i] = catalog.Slot{
			Label: slots[i].Label,
			Exts:  append([]string(nil), slots[i].Exts...),
		}
		if slots[i].Mgl != nil {
			converted[i].Mgl = &catalog.MGLParams{
				Delay:      slots[i].Mgl.Delay,
				Method:     slots[i].Mgl.Method,
				Index:      slots[i].Mgl.Index,
				ResetDelay: slots[i].Mgl.resetDelay,
				ResetHold:  slots[i].Mgl.resetHold,
			}
		}
	}
	return converted
}

func systemFromCatalog(core *catalog.Core) System {
	metadata := systemMetadataByID[core.ID]
	name := metadata.Name
	if name == "" {
		name = core.ID
	}

	return System{
		Id:             core.ID,
		Name:           name,
		Category:       metadata.Category,
		ReleaseDate:    metadata.ReleaseDate,
		Manufacturer:   metadata.Manufacturer,
		Alias:          append([]string(nil), metadata.Aliases...),
		SetName:        core.SetName,
		SetNameSameDir: core.SetNameSameDir,
		Folder:         append([]string(nil), core.Folders...),
		Rbf:            core.RBF,
		Slots:          slotsFromCatalog(core.Slots),
		extensions:     append([]string(nil), core.Extensions...),
	}
}

func buildSystems() map[string]System {
	definitions := catalog.All()
	systems := make(map[string]System, len(definitions))
	for i := range definitions {
		definition := &definitions[i]
		systems[definition.ID] = systemFromCatalog(definition)
	}
	return systems
}

// Systems combines Zaparoo's canonical MiSTer launch catalog with mrext-only
// display metadata used by legacy menus and Remote responses.
var Systems = buildSystems()

func buildCoreGroups() map[string][]System {
	definitions := catalog.Groups()
	groups := make(map[string][]System, len(definitions))
	for id, members := range definitions {
		converted := make([]System, len(members))
		for i := range members {
			converted[i] = systemFromCatalog(&members[i])
		}
		groups[id] = converted
	}
	return groups
}

// CoreGroups preserves legacy group lookup while sourcing membership and MGL
// slot composition from Zaparoo's canonical catalog.
var CoreGroups = buildCoreGroups()

// CatalogCore converts a legacy system value to the canonical launch model.
func CatalogCore(system *System) catalog.Core {
	return catalog.Core{
		ID:             system.Id,
		SetName:        system.SetName,
		RBF:            system.Rbf,
		Slots:          slotsToCatalog(system.Slots),
		SetNameSameDir: system.SetNameSameDir,
	}
}

func PathToMglDef(system *System, path string) (*MglParams, error) {
	params, err := catalog.PathToMGLDef(&catalog.Core{ID: system.Id, Slots: slotsToCatalog(system.Slots)}, path)
	if err != nil {
		return nil, fmt.Errorf("resolve catalog MGL parameters: %w", err)
	}
	if params == nil {
		return nil, nil //nolint:nilnil // No matching MGL slot is a valid optional result.
	}
	return &MglParams{
		Delay:      params.Delay,
		Method:     params.Method,
		Index:      params.Index,
		resetDelay: params.ResetDelay,
		resetHold:  params.ResetHold,
	}, nil
}
