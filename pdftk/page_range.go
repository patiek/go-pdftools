package pdftk

import (
	"fmt"
	"strings"
)

// PageRangeQualifier optionally chooses the pages (even vs odd) to include.
type PageRangeQualifier string

const (
	Odd  PageRangeQualifier = "odd"
	Even PageRangeQualifier = "even"
)

// PageRangeRotation sets the page rotation as follows (in degrees).
// The left, right, and down rotations make relative adjustments to page
// and must follow other page range adjustments.
//
// 	north: 0, east: 90, south: 180, west: 270
// 	left: -90, right: +90, down: +180
type PageRangeRotation string

const (
	North PageRangeRotation = "north"
	East  PageRangeRotation = "east"
	South PageRangeRotation = "south"
	West  PageRangeRotation = "west"
	Left  PageRangeRotation = "left"
	Right PageRangeRotation = "right"
	Down  PageRangeRotation = "down"
)

type PageRange struct {
	// The file handle name to reference (a key in InputFileMap).
	FileHandleName string

	// Optional rotation of page range
	Rotation PageRangeRotation

	// BeginPage must not be 0 for EndPage to have effect.
	// Omitting BeginPage will use entire document
	// Omitting EndPage will go from BeginPage to end of document.
	// For exactly one page, set BeginPage and EndPage to same page.
	// First page starts at 1.
	BeginPage uint32
	EndPage   uint32
	Qualifier PageRangeQualifier
}

// String representation suitable for parameter input.
func (pr *PageRange) String() string {
	var sb strings.Builder
	sb.WriteString(pr.FileHandleName)
	if pr.BeginPage != 0 {
		fmt.Fprintf(&sb, "%d", pr.BeginPage)
		if pr.BeginPage != pr.EndPage {
			if pr.EndPage != 0 {
				fmt.Fprintf(&sb, "-%d", pr.EndPage)
			} else {
				fmt.Fprint(&sb, "-end")
			}
		}
	}
	if pr.Qualifier != "" {
		fmt.Fprintf(&sb, "%s", pr.Qualifier)
	}
	if pr.Rotation != "" {
		fmt.Fprintf(&sb, "%s", pr.Rotation)
	}
	return sb.String()
}

func pageRangesToStrings(pageRanges []PageRange) []string {
	sl := make([]string, len(pageRanges))
	for i, pr := range pageRanges {
		sl[i] = pr.String()
	}
	return sl
}
