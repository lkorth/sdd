// Package record parses sdd requirement records, following the grammar in records.md.
package record

import (
	"os"
	"regexp"
	"strings"
)

const idPattern = `L0_\d{3}|L[12]_[A-Z]{2,4}_\d{3}`

var (
	// ValidID matches a whole record ID.
	ValidID = regexp.MustCompile(`^(?:` + idPattern + `)$`)

	recordHeading = regexp.MustCompile(`^### (` + idPattern + `): (.+?)\s*$`)
	anyHeading    = regexp.MustCompile(`^#{1,6}\s`)
	fence         = regexp.MustCompile("^(```+|~~~+)")
	fieldLine     = regexp.MustCompile(`^([a-z_]+):\s*(.*?)\s*$`)
	stepLine      = regexp.MustCompile(`^- (WHEN|THEN|AND)\b\s*(.*)$`)
)

// Record is one requirement. Line and EndLine are 1-based and inclusive.
type Record struct {
	ID         string
	Title      string
	Path       string
	Line       int
	EndLine    int
	Fields     []Field
	Obligation []string
	Rationale  []string
	Scenarios  []Scenario
	Body       []string // every line after the heading, as written
}

// Field is one key: value line of the fields block, in the order written.
type Field struct {
	Key, Value string
}

// Scenario holds a name and its steps, each as "WHEN ...", "THEN ..." or "AND ...".
type Scenario struct {
	Name  string
	Steps []string
}

// Level returns L0, L1 or L2.
func (r *Record) Level() string { return r.ID[:2] }

// Feature returns the feature code, or "" for an L0 record.
func (r *Record) Feature() string {
	if r.Level() == "L0" {
		return ""
	}
	return strings.Split(r.ID, "_")[1]
}

// Field returns the value of a field and whether the record has it.
func (r *Record) Field(key string) (string, bool) {
	for _, f := range r.Fields {
		if f.Key == key {
			return f.Value, true
		}
	}
	return "", false
}

// Parent returns the parent ID, or "" for an L0 record or a record with no parent.
func (r *Record) Parent() string {
	parent, _ := r.Field("parent")
	if parent == "self" {
		return ""
	}
	return parent
}

// Status returns the status field, or "active" when the record has none.
func (r *Record) Status() string {
	if status, ok := r.Field("status"); ok {
		return status
	}
	return "active"
}

// ObligationText returns the obligation as one unwrapped paragraph.
func (r *Record) ObligationText() string { return Unwrap(r.Obligation) }

// RationaleText returns the rationale without its label, one line per paragraph.
func (r *Record) RationaleText() string {
	return strings.TrimSpace(strings.TrimPrefix(Unwrap(r.Rationale), "Rationale:"))
}

// ParseFile returns the feature code from the frontmatter, the title of the file, and its
// records. A record runs from its heading to the next heading of any level, or to the end of
// the file.
func ParseFile(path string) (code, name string, records []*Record, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}

	start := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for n := 1; n < len(lines); n++ {
			if strings.TrimSpace(lines[n]) == "---" {
				start = n + 1
				break
			}
			if m := fieldLine.FindStringSubmatch(lines[n]); m != nil && m[1] == "feature" {
				code = m[2]
			}
		}
	}

	current := -1
	var heading []string
	closeRecord := func(end int) {
		body := append([]string(nil), lines[current+1:end]...)
		for len(body) > 0 && strings.TrimSpace(body[len(body)-1]) == "" {
			body = body[:len(body)-1]
		}
		records = append(records, parseRecord(heading[1], heading[2], path, current+1, body))
	}

	inFence := false
	for n := start; n < len(lines); n++ {
		line := lines[n]
		if fence.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence || !anyHeading.MatchString(line) {
			continue
		}
		if current >= 0 {
			closeRecord(n)
			current = -1
		}
		if m := recordHeading.FindStringSubmatch(line); m != nil {
			current, heading = n, m
		} else if name == "" && strings.HasPrefix(line, "# ") {
			name = strings.TrimSpace(line[2:])
		}
	}
	if current >= 0 {
		closeRecord(len(lines))
	}
	return code, name, records, nil
}

// parseRecord splits a record body into its pieces. The fields block comes first, the
// obligation is the next paragraph, and the rationale runs to the first Scenario: line.
func parseRecord(id, title, path string, line int, body []string) *Record {
	r := &Record{ID: id, Title: title, Path: path, Line: line, EndLine: line + len(body), Body: body}
	n := 0
	skipBlank := func() {
		for n < len(body) && strings.TrimSpace(body[n]) == "" {
			n++
		}
	}

	skipBlank()
	if n < len(body) && fence.MatchString(body[n]) {
		for n++; n < len(body) && !fence.MatchString(body[n]); n++ {
			if m := fieldLine.FindStringSubmatch(body[n]); m != nil {
				r.Fields = append(r.Fields, Field{m[1], m[2]})
			}
		}
		n++
	}

	skipBlank()
	if n < len(body) && !strings.HasPrefix(body[n], "Rationale:") && !strings.HasPrefix(body[n], "Scenario:") {
		for ; n < len(body) && strings.TrimSpace(body[n]) != ""; n++ {
			r.Obligation = append(r.Obligation, body[n])
		}
	}

	for ; n < len(body); n++ {
		text := body[n]
		switch {
		case strings.HasPrefix(text, "Scenario:"):
			r.Scenarios = append(r.Scenarios, Scenario{Name: strings.TrimSpace(strings.TrimPrefix(text, "Scenario:"))})
		case len(r.Scenarios) > 0:
			s := &r.Scenarios[len(r.Scenarios)-1]
			if m := stepLine.FindStringSubmatch(text); m != nil {
				s.Steps = append(s.Steps, strings.TrimSpace(m[1]+" "+m[2]))
			} else if strings.TrimSpace(text) != "" && len(s.Steps) > 0 {
				s.Steps[len(s.Steps)-1] += " " + strings.TrimSpace(text)
			}
		case strings.HasPrefix(text, "Rationale:") || len(r.Rationale) > 0:
			r.Rationale = append(r.Rationale, text)
		}
	}
	for len(r.Rationale) > 0 && strings.TrimSpace(r.Rationale[len(r.Rationale)-1]) == "" {
		r.Rationale = r.Rationale[:len(r.Rationale)-1]
	}
	return r
}

// Unwrap joins wrapped lines into one line for each paragraph, with a blank line between
// paragraphs.
func Unwrap(lines []string) string {
	var paragraphs, current []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			current = append(current, strings.TrimSpace(line))
		} else if len(current) > 0 {
			paragraphs = append(paragraphs, strings.Join(current, " "))
			current = nil
		}
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, strings.Join(current, " "))
	}
	return strings.Join(paragraphs, "\n\n")
}
