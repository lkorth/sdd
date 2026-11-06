package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lkorth/sdd/internal/record"
)

// The canvas follows the JSON Canvas format of Obsidian: https://jsoncanvas.org/spec/1.0/
// The json-canvas skill in https://github.com/kepano/obsidian-skills gives the conventions.
// Each ID is 16 hex digits, as Obsidian writes them. A hash of the location gives each ID, so
// that the same records give the same canvas.
type canvas struct {
	Nodes []canvasNode `json:"nodes"`
	Edges []canvasEdge `json:"edges"`
}

type canvasNode struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Color   string `json:"color,omitempty"`
	File    string `json:"file,omitempty"`
	Subpath string `json:"subpath,omitempty"`
	Text    string `json:"text,omitempty"`
	Label   string `json:"label,omitempty"`
}

type canvasEdge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	FromSide string `json:"fromSide"`
	ToNode   string `json:"toNode"`
	ToSide   string `json:"toSide"`
	Color    string `json:"color,omitempty"`
}

const (
	cardWidth  = 480
	cardHeight = 320
	fileHeight = 180
	gapX       = 60
	gapY       = 200
	padding    = 40
	grid       = 20
	missingRed = "1"
)

// featureColors are the preset canvas colors for features. Red is for missing and retired
// parents. L0 records have no color.
var featureColors = []string{"2", "3", "4", "5", "6"}

// canvasLayout places each record below its parent, so that the canvas shows the parent tree.
type canvasLayout struct {
	c        *record.Corpus
	vault    string
	colors   map[string]string // feature path to color
	children map[string][]*record.Record
	nodeOf   map[string]string // record ID to the ID of its node
	placed   map[*record.Record]bool
	out      canvas
}

// canvasName is the canvas that delete and the hooks refresh, in the sdd directory.
const canvasName = "records.canvas"

// canvasResult reports a canvas file. Status is updated or unchanged for --out, and current or
// stale for --check.
type canvasResult struct {
	File   string `json:"file"`
	Status string `json:"status"`
}

func cmdCanvas(c *record.Corpus, o *options, _ []string, out *printer) int {
	switch {
	case o.canvasOut != "" && o.canvasCheck != "":
		out.fail("use only one of --out and --check")
		return 2
	case o.canvasOut != "":
		status, err := writeCanvas(c, o.canvasOut, o.vault)
		if err != nil {
			out.err = err
			return 2
		}
		out.canvasResult(canvasResult{out.path(o.canvasOut), status})
		return 0
	case o.canvasCheck != "":
		vault := o.vault
		if vault == "" {
			vault = vaultOf(c, o.canvasCheck)
		}
		data, err := buildCanvas(c, vault)
		if err != nil {
			out.err = err
			return 2
		}
		if !sameCanvas(o.canvasCheck, data) {
			out.canvasResult(canvasResult{out.path(o.canvasCheck), "stale"})
			return 1
		}
		out.canvasResult(canvasResult{out.path(o.canvasCheck), "current"})
		return 0
	}
	vault := o.vault
	if vault == "" {
		vault = filepath.Dir(c.Dir)
	}
	data, err := buildCanvas(c, vault)
	if err != nil {
		out.err = err
		return 2
	}
	out.printf("%s", data)
	return 0
}

func (p *printer) canvasResult(r canvasResult) {
	if p.json {
		p.writeJSON(r)
		return
	}
	if r.Status == "stale" {
		p.printf("%s  stale. To update it, run: sdd canvas --out %s\n", r.File, r.File)
		return
	}
	p.printf("%s  %s\n", r.File, r.Status)
}

// writeCanvas writes the canvas to path if its content changed, and returns updated or
// unchanged. With no vault, it keeps the vault of the canvas at path, or uses the directory that
// holds the sdd directory.
func writeCanvas(c *record.Corpus, path, vault string) (string, error) {
	if vault == "" {
		vault = vaultOf(c, path)
	}
	data, err := buildCanvas(c, vault)
	if err != nil {
		return "", err
	}
	if sameCanvas(path, data) {
		return "unchanged", nil
	}
	return "updated", os.WriteFile(path, data, 0o644)
}

func buildCanvas(c *record.Corpus, vault string) ([]byte, error) {
	vault, err := filepath.Abs(vault)
	if err != nil {
		return nil, err
	}
	l := &canvasLayout{c: c, vault: vault, colors: map[string]string{},
		children: map[string][]*record.Record{}, nodeOf: map[string]string{},
		placed: map[*record.Record]bool{}, out: canvas{Nodes: []canvasNode{}, Edges: []canvasEdge{}}}
	if err := l.build(); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(l.out, "", "\t")
	return append(data, '\n'), err
}

// sameCanvas reports whether the file at path holds the same JSON as data. Obsidian writes a
// canvas in its own format, so the test compares the values, not the bytes.
func sameCanvas(path string, data []byte) bool {
	old, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var a, b any
	return json.Unmarshal(old, &a) == nil && json.Unmarshal(data, &b) == nil && reflect.DeepEqual(a, b)
}

// vaultOf finds the vault of the canvas at path: the directory that, joined with the path of a
// file node, gives the path of a feature file. With no match, it is the directory that holds the
// sdd directory.
func vaultOf(c *record.Corpus, path string) string {
	fallback := filepath.Dir(c.Dir)
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	var old struct {
		Nodes []struct{ Type, File string }
	}
	if json.Unmarshal(data, &old) != nil {
		return fallback
	}
	for _, n := range old.Nodes {
		if n.Type != "file" || n.File == "" {
			continue
		}
		file := string(filepath.Separator) + filepath.FromSlash(n.File)
		for _, f := range c.Features {
			if abs, err := filepath.Abs(f.Path); err == nil && strings.HasSuffix(abs, file) {
				return strings.TrimSuffix(abs, file)
			}
		}
	}
	return fallback
}

func (l *canvasLayout) build() error {
	n := 0
	for _, f := range l.c.Features {
		if f.Code != "L0" {
			l.colors[f.Path] = featureColors[n%len(featureColors)]
			n++
		}
	}
	for _, f := range l.c.Features {
		if _, err := l.vaultPath(f.Path); err != nil {
			return err
		}
	}
	for _, r := range l.c.Records {
		if _, ok := l.nodeOf[r.ID]; !ok {
			l.nodeOf[r.ID] = l.nodeID(r)
		}
		if p := r.Parent(); p != "" {
			l.children[p] = append(l.children[p], r)
		}
	}

	l.addFiles()

	// Roots are L0 records, then the missing and retired parents, then a record in a parent
	// cycle, which no root reaches.
	x := 0
	for _, r := range l.c.Records {
		if r.Parent() == "" && !l.placed[r] {
			x = l.placeTree(r, x, 0) + gapX
		}
	}
	for _, r := range l.c.Records {
		p := r.Parent()
		if p == "" || len(l.c.ByID[p]) > 0 {
			continue
		}
		if _, done := l.nodeOf[p]; done {
			continue
		}
		x = l.placeMissing(p, x) + gapX
	}
	for _, r := range l.c.Records {
		if !l.placed[r] {
			x = l.placeTree(r, x, 0) + gapX
		}
	}

	for _, r := range l.c.Records {
		if p := r.Parent(); p != "" && l.placed[r] {
			l.out.Edges = append(l.out.Edges, canvasEdge{ID: hexID("edge " + l.nodeID(r)), FromNode: l.nodeID(r),
				FromSide: "top", ToNode: l.nodeOf[p], ToSide: "bottom", Color: l.colors[r.Path]})
		}
	}
	return nil
}

// addFiles puts one node for each feature file in a group to the left of the tree. A node for
// a whole file shows the file, so that a reader can open it from the canvas.
func (l *canvasLayout) addFiles() {
	x := -cardWidth - 4*gapX
	group := canvasNode{ID: hexID("group files"), Type: "group", X: x - padding, Y: -padding,
		Width: cardWidth + 2*padding, Height: len(l.c.Features)*(fileHeight+gapX) - gapX + 2*padding,
		Label: "Files"}
	l.out.Nodes = append(l.out.Nodes, group)
	for i, f := range l.c.Features {
		file, _ := l.vaultPath(f.Path)
		l.out.Nodes = append(l.out.Nodes, canvasNode{ID: hexID("file " + file), Type: "file", X: x,
			Y: i * (fileHeight + gapX), Width: cardWidth, Height: fileHeight, Color: l.colors[f.Path],
			File: file})
	}
}

// placeTree places r and its descendants with their left edge at x, and returns the right edge.
func (l *canvasLayout) placeTree(r *record.Record, x, depth int) int {
	l.placed[r] = true
	var kids []*record.Record
	if l.nodeOf[r.ID] == l.nodeID(r) {
		for _, child := range l.children[r.ID] {
			if !l.placed[child] {
				l.placed[child] = true
				kids = append(kids, child)
			}
		}
	}
	right := l.placeRow(kids, x, depth+1)
	left, mid := x, x+cardWidth
	if len(kids) > 0 {
		mid = snap((x + right) / 2)
		left = mid - cardWidth/2
	} else {
		right = mid
	}
	file, _ := l.vaultPath(r.Path)
	l.out.Nodes = append(l.out.Nodes, canvasNode{ID: l.nodeID(r), Type: "file", X: left,
		Y: depth * (cardHeight + gapY), Width: cardWidth, Height: cardHeight, Color: l.colors[r.Path],
		File: file, Subpath: "#" + r.ID + ": " + r.Title})
	return max(right, left+cardWidth)
}

// placeRow places records side by side from x, one level below depth, and returns the right
// edge.
func (l *canvasLayout) placeRow(records []*record.Record, x, depth int) int {
	right := x
	for i, r := range records {
		if i > 0 {
			x = right + gapX
		}
		right = l.placeTree(r, x, depth)
	}
	return right
}

// placeMissing places a text node for a parent that is not a record, with the records that
// name it below it.
func (l *canvasLayout) placeMissing(id string, x int) int {
	text := "**" + id + "**\n\nmissing: no record and no retired entry"
	if retired := l.c.Retired[id]; retired != nil {
		text = "**" + id + "**\n\nretired"
		if retired.Reason != "" {
			text += ": " + retired.Reason
		}
	}
	node := hexID("missing " + id)
	l.nodeOf[id] = node
	var kids []*record.Record
	for _, child := range l.children[id] {
		if !l.placed[child] {
			l.placed[child] = true
			kids = append(kids, child)
		}
	}
	right := l.placeRow(kids, x, 1)
	left := snap((x+right)/2) - cardWidth/2
	l.out.Nodes = append(l.out.Nodes, canvasNode{ID: node, Type: "text", X: left, Y: 0,
		Width: cardWidth, Height: fileHeight, Color: missingRed, Text: text})
	return max(right, left+cardWidth)
}

func (l *canvasLayout) vaultPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(l.vault, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s is not in the vault %s. Use --vault", path, l.vault)
	}
	return filepath.ToSlash(rel), nil
}

// nodeID is unique when two records have the same ID. build checks each path first.
func (l *canvasLayout) nodeID(r *record.Record) string {
	file, _ := l.vaultPath(r.Path)
	return hexID(fmt.Sprintf("record %s %s:%d", r.ID, file, r.Line))
}

func hexID(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}

// snap moves a position down to the grid.
func snap(n int) int {
	return n - ((n%grid)+grid)%grid
}
