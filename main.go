// The coveragecollector is a simple cli tool for print the total coverate of each package for a
// test run against multiple packages with the coverage, e.g.:
//  go tool -coverprofile cover.out coverpkg ./... ./...
//
// To use this cli, simply run the tests and provide the coverage file name to the cli.
// Future plans:
//  - Support more than just set mode.
//  - Support more than one file.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/cover"
)

var (
	ErrInvalidMode     = errors.New("coverage collector only supported set mode")
	ErrNoProfiles      = errors.New("no cover profiles proved")
	ErrTooManyProfiles = errors.New("only one cover profile is supported")
)

type CoverageCollector struct {
	files [][]*cover.Profile
}

func NewCoverageCollector(files ...[]*cover.Profile) *CoverageCollector {
	return &CoverageCollector{files}
}

func (c *CoverageCollector) Validate() error {
	if len(c.files) == 0 {
		return ErrNoProfiles
	}
	if len(c.files) > 1 {
		return ErrTooManyProfiles
	}
	for _, f := range c.files {
		for _, p := range f {
			if p.Mode != "set" {
				return ErrInvalidMode
			}
		}
	}
	return nil
}

type PackageCoverage struct {
	Package string
	Files   []FileCoverage
}

type FileCoverage struct {
	FileName string
	// Can contain repeats but in the original order
	Blocks []cover.ProfileBlock
}

func (fc *FileCoverage) UniqueBlocks() []cover.ProfileBlock {
	type key struct {
		startLine, startCol int
	}
	unique := map[key]*cover.ProfileBlock{}
	for _, b := range fc.Blocks {
		k := key{startLine: b.StartLine, startCol: b.StartCol}
		if bb, ok := unique[k]; !ok {
			b2 := b
			unique[k] = &b2
		} else {
			bb.Count += b.Count
		}
	}

	uniqueBlocks := make([]cover.ProfileBlock, len(unique))
	for _, b := range unique {
		uniqueBlocks = append(uniqueBlocks, *b)
	}

	sort.Slice(uniqueBlocks, func(i, j int) bool {
		if uniqueBlocks[i].StartLine != uniqueBlocks[j].StartLine {
			return uniqueBlocks[i].StartLine < uniqueBlocks[j].StartLine
		}
		return uniqueBlocks[i].StartCol < uniqueBlocks[j].StartCol
	})
	return uniqueBlocks
}

func (p *PackageCoverage) Coverage() float64 {
	totalStmts := 0
	totalCoveredStmts := 0
	for _, f := range p.Files {
		for _, b := range f.UniqueBlocks() {
			totalStmts += b.NumStmt
			if b.Count > 0 {
				totalCoveredStmts += b.NumStmt
			}
		}
	}
	return float64(totalCoveredStmts) / float64(totalStmts)
}

func (c *CoverageCollector) CollectPackages() []*PackageCoverage {
	packages := map[string]*PackageCoverage{}
	for _, f := range c.files {
	Loop:
		for _, p := range f {
			pkg := filepath.Dir(p.FileName)
			if pCov, ok := packages[pkg]; !ok {
				packages[pkg] = &PackageCoverage{
					Package: pkg,
					Files: []FileCoverage{{
						FileName: p.FileName,
						Blocks:   p.Blocks,
					}},
				}
			} else {
				for i, ff := range pCov.Files {
					if ff.FileName == p.FileName {
						pCov.Files[i].Blocks = append(pCov.Files[i].Blocks, p.Blocks...)
						continue Loop
					}
				}
				pCov.Files = append(pCov.Files, FileCoverage{
					FileName: p.FileName,
					Blocks:   p.Blocks,
				})
				sort.Slice(pCov.Files, func(i, j int) bool { return pCov.Files[i].FileName < pCov.Files[j].FileName })
			}
		}
	}

	packageSlice := make([]*PackageCoverage, 0, len(packages))
	for _, p := range packages {
		packageSlice = append(packageSlice, p)
	}

	sort.Slice(packageSlice, func(i, j int) bool { return packageSlice[i].Package < packageSlice[j].Package })
	return packageSlice
}

func main() {

	files := [][]*cover.Profile{}
	for _, a := range os.Args[1:] {
		profiles, err := cover.ParseProfiles(a)
		if err != nil {
			reportErr(err)
		}
		files = append(files, profiles)
	}

	c := NewCoverageCollector(files...)
	if err := c.Validate(); err != nil {
		reportErr(err)
	}

	packages := c.CollectPackages()
	maxWidth := 0
	for _, p := range packages {
		if len(p.Package) > maxWidth {
			maxWidth = len(p.Package)
		}
	}
	for _, p := range packages {
		padding := strings.Repeat(" ", maxWidth-len(p.Package)+1)
		fmt.Printf("%s%scoverage (%.1f)%%\n", p.Package, padding, 100.0*p.Coverage())
	}
}

func reportErr(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(1)
}
