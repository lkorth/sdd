package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lkorth/sdd/internal/record"
)

// item is one answer: a record, a retired entry, or an ID that is neither.
type item struct {
	rec     *record.Record
	retired *record.Retired
	missing string
	depth   int // 0 unless the answer is a tree
}

type printer struct {
	w      io.Writer
	errW   io.Writer
	json   bool
	pretty bool
	cwd    string
	err    error
}

func newPrinter(w, errW io.Writer, asJSON bool) *printer {
	cwd, _ := os.Getwd()
	pretty := false
	if f, ok := w.(*os.File); ok {
		if info, err := f.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			pretty = true
		}
	}
	return &printer{w: w, errW: errW, json: asJSON, pretty: pretty, cwd: cwd}
}

// fail prints a message to stderr. The caller returns the exit code.
func (p *printer) fail(format string, args ...any) {
	fmt.Fprintf(p.errW, "sdd: "+format+"\n", args...)
}

func (p *printer) printf(format string, args ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, format, args...)
	}
}

func (p *printer) writeJSON(v any) {
	enc := json.NewEncoder(p.w)
	enc.SetEscapeHTML(false)
	if p.pretty {
		enc.SetIndent("", "  ")
	}
	if p.err == nil {
		p.err = enc.Encode(v)
	}
}

// path returns a path relative to the working directory when it is below it.
func (p *printer) path(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if rel, err := filepath.Rel(p.cwd, abs); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return abs
}

func (p *printer) where(it item) string {
	if it.retired != nil {
		return fmt.Sprintf("%s:%d", p.path(it.retired.Path), it.retired.Line)
	}
	return fmt.Sprintf("%s:%d-%d", p.path(it.rec.Path), it.rec.Line, it.rec.EndLine)
}

func (p *printer) items(items []item, detail string) {
	if p.json {
		out := make([]jsonItem, len(items))
		for i, it := range items {
			out[i] = p.toJSON(it, detail)
		}
		p.writeJSON(out)
		return
	}
	if len(items) == 0 {
		return
	}
	texts := make([]string, len(items))
	for i, it := range items {
		texts[i] = p.toText(it, detail)
	}
	separator := "\n---\n"
	if detail == "short" {
		separator = "\n"
	}
	p.printf("%s\n", strings.Join(texts, separator))
}

func (p *printer) toText(it item, detail string) string {
	indent := ""
	if it.depth > 1 {
		indent = strings.Repeat("  ", it.depth-1)
	}
	switch {
	case it.missing != "":
		return indent + it.missing + "  missing: no record and no retired entry"
	case it.retired != nil:
		text := indent + it.retired.ID + "  " + p.where(it) + "  retired"
		if it.retired.Reason != "" {
			text += ": " + it.retired.Reason
		}
		return text
	}
	r := it.rec
	if detail == "short" {
		text := indent + r.ID + "  " + p.where(it) + "  " + r.Title
		if r.Status() == "tbr" {
			text += "  (tbr)"
		}
		return text
	}
	meta := []string{p.where(it)}
	for _, f := range r.Fields {
		meta = append(meta, f.Key+" "+f.Value)
	}
	head := r.ID + ": " + r.Title + "\n" + strings.Join(meta, "  ")
	if detail == "full" {
		return head + "\n" + strings.Trim(strings.Join(r.Body, "\n"), "\n")
	}
	if len(r.Obligation) == 0 {
		return head
	}
	return head + "\n" + strings.Join(r.Obligation, "\n")
}

type jsonScenario struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

type jsonItem struct {
	ID         string         `json:"id"`
	Missing    bool           `json:"missing,omitempty"`
	Retired    bool           `json:"retired,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	Date       string         `json:"date,omitempty"`
	PR         int            `json:"pr,omitempty"`
	Title      string         `json:"title,omitempty"`
	File       string         `json:"file,omitempty"`
	Line       int            `json:"line,omitempty"`
	EndLine    int            `json:"end_line,omitempty"`
	Depth      int            `json:"depth,omitempty"`
	Level      string         `json:"level,omitempty"`
	Feature    string         `json:"feature,omitempty"`
	Parent     string         `json:"parent,omitempty"`
	Status     string         `json:"status,omitempty"`
	Obligation string         `json:"obligation,omitempty"`
	Rationale  string         `json:"rationale,omitempty"`
	Scenarios  []jsonScenario `json:"scenarios,omitempty"`
}

func (p *printer) toJSON(it item, detail string) jsonItem {
	switch {
	case it.missing != "":
		return jsonItem{ID: it.missing, Missing: true}
	case it.retired != nil:
		r := it.retired
		return jsonItem{ID: r.ID, Retired: true, Reason: r.Reason, Date: r.Date, PR: r.PR,
			File: p.path(r.Path), Line: r.Line}
	}
	r := it.rec
	out := jsonItem{ID: r.ID, Title: r.Title, File: p.path(r.Path), Line: r.Line,
		EndLine: r.EndLine, Depth: it.depth}
	if detail == "short" {
		if r.Status() != "active" {
			out.Status = r.Status()
		}
		return out
	}
	out.Level, out.Feature, out.Status = r.Level(), r.Feature(), r.Status()
	out.Parent, _ = r.Field("parent")
	out.Obligation = r.ObligationText()
	if detail == "full" {
		out.Rationale = r.RationaleText()
		for _, s := range r.Scenarios {
			out.Scenarios = append(out.Scenarios, jsonScenario{s.Name, s.Steps})
		}
	}
	return out
}
