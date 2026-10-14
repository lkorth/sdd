package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Dependents returns each other record that names id as its parent, in its obligation, or in
// a scenario. A retired ID cannot appear in those places. A rationale can name it.
func (c *Corpus) Dependents(id string) []*Record {
	word := regexp.MustCompile(`\b` + regexp.QuoteMeta(id) + `\b`)
	var out []*Record
	for _, r := range c.Records {
		if r.ID == id {
			continue
		}
		if r.Parent() == id || word.MatchString(strings.Join(r.Obligation, "\n")) || namedInScenarios(r, word) {
			out = append(out, r)
		}
	}
	return out
}

func namedInScenarios(r *Record, word *regexp.Regexp) bool {
	for _, s := range r.Scenarios {
		if word.MatchString(s.Name) || word.MatchString(strings.Join(s.Steps, "\n")) {
			return true
		}
	}
	return false
}

// RemoveRecord deletes the lines of a record from its file, with the blank lines that follow
// it, so that the text around it keeps its spacing.
func RemoveRecord(r *Record) error {
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return err
	}
	newline := "\n"
	if strings.Contains(string(data), "\r\n") {
		newline = "\r\n"
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}

	start, end := r.Line-1, r.EndLine
	for end < len(lines) && strings.TrimSpace(lines[end]) == "" {
		end++
	}
	kept := append(append([]string(nil), lines[:start]...), lines[end:]...)
	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	if len(kept) == 0 {
		return writeFile(r.Path, "")
	}
	return writeFile(r.Path, strings.Join(kept, newline)+newline)
}

// AppendRetired adds one line to the ledger and returns its 1-based line number.
func AppendRetired(dir string, entry Retired) (string, int, error) {
	path := filepath.Join(dir, "retired.jsonl")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", 0, err
	}
	text := string(data)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	// Write the keys in a fixed order, with a space after each separator, as in records.md.
	fields := []string{`"id": ` + quote(entry.ID)}
	if entry.Reason != "" {
		fields = append(fields, `"reason": `+quote(entry.Reason))
	}
	if entry.Date != "" {
		fields = append(fields, `"date": `+quote(entry.Date))
	}
	if entry.PR != 0 {
		fields = append(fields, `"pr": `+strconv.Itoa(entry.PR))
	}
	text += "{" + strings.Join(fields, ", ") + "}\n"
	if err := writeFile(path, text); err != nil {
		return "", 0, err
	}
	return path, strings.Count(text, "\n"), nil
}

func quote(s string) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s) // a string always encodes
	return strings.TrimSuffix(b.String(), "\n")
}

// writeFile replaces a file through a temporary file, so that a failed write leaves the old
// file whole.
func writeFile(path, text string) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".sdd-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(text); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
