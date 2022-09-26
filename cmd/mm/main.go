package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

func main() {
	activePaths := flag.Bool("active-paths", false, "print active system paths")
	allPaths := flag.Bool("all-paths", false, "print all detected system paths")
	filterSystems := flag.String("s", "all", "restrict operation to systems (comma separated)")
	timed := flag.Bool("t", false, "show how long operation took")
	flag.Parse()

	start := time.Now()

	var selectedSystems []games.System
	if *filterSystems == "all" {
		selectedSystems = games.AllSystems()
	} else {
		filterIds := strings.Split(*filterSystems, ",")
		for _, filterId := range filterIds {
			system, err := games.LookupSystem(filterId)
			if err != nil {
				continue
			} else {
				selectedSystems = append(selectedSystems, *system)
			}
		}
	}

	if *activePaths {
		paths := games.GetActiveSystemPaths(selectedSystems)
		for _, path := range paths {
			fmt.Printf("%s:%s\n", path.System.Id, path.Path)
		}
	} else if *allPaths {
		systemPaths := games.GetSystemPaths()
		filteredPaths := make(map[string][]string)

		for systemId, paths := range systemPaths {
			for _, system := range selectedSystems {
				if system.Id == systemId {
					filteredPaths[systemId] = paths
				}
			}
		}

		for k, v := range filteredPaths {
			for _, p := range v {
				fmt.Printf("%s:%s\n", k, p)
			}
		}
	}

	if *timed {
		seconds := int(time.Since(start).Seconds())
		milliseconds := int(time.Since(start).Milliseconds())
		remainder := milliseconds % int(time.Second)
		fmt.Printf("Operation took %d.%ds\n", int(seconds), remainder)
	}

	os.Exit(0)
}
