package record

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Corpus is every record, feature and retired ID of one project.
type Corpus struct {
	Dir      string
	Records  []*Record
	Features []*Feature
	Retired  map[string]*Retired
	ByID     map[string][]*Record // more than one record per ID is a duplicate
}

// Feature is one file of records. system.md is the feature with the code L0.
type Feature struct {
	Code    string
	Name    string
	Path    string
	Records []*Record
}

// Retired is one line of retired.jsonl. Line is 1-based.
type Retired struct {
	ID     string
	Path   string
	Line   int
	Reason string
	Date   string
	PR     int
}

// Load reads system.md, features/*.md and retired.jsonl from an sdd directory.
func Load(dir string) (*Corpus, error) {
	c := &Corpus{Dir: dir, Retired: map[string]*Retired{}, ByID: map[string][]*Record{}}

	system := filepath.Join(dir, "system.md")
	if _, err := os.Stat(system); err == nil {
		if err := c.addFile(system, "L0", "System"); err != nil {
			return nil, err
		}
	}
	paths, err := filepath.Glob(filepath.Join(dir, "features", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := c.addFile(path, "?", strings.TrimSuffix(filepath.Base(path), ".md")); err != nil {
			return nil, err
		}
	}
	if err := c.loadLedger(filepath.Join(dir, "retired.jsonl")); err != nil {
		return nil, err
	}
	for _, r := range c.Records {
		c.ByID[r.ID] = append(c.ByID[r.ID], r)
	}
	return c, nil
}

func (c *Corpus) addFile(path, defaultCode, defaultName string) error {
	code, name, records, err := ParseFile(path)
	if err != nil {
		return err
	}
	if defaultCode == "L0" || code == "" {
		code = defaultCode
	}
	if name == "" {
		name = defaultName
	}
	c.Features = append(c.Features, &Feature{Code: code, Name: name, Path: path, Records: records})
	c.Records = append(c.Records, records...)
	return nil
}

func (c *Corpus) loadLedger(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for n := 1; scanner.Scan(); n++ {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var entry struct {
			ID     string `json:"id"`
			Reason string `json:"reason"`
			Date   string `json:"date"`
			PR     int    `json:"pr"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return fmt.Errorf("%s:%d: %w", path, n, err)
		}
		c.Retired[entry.ID] = &Retired{ID: entry.ID, Path: path, Line: n,
			Reason: entry.Reason, Date: entry.Date, PR: entry.PR}
	}
	return scanner.Err()
}

// Children returns the records whose parent is id, in file order.
func (c *Corpus) Children(id string) []*Record {
	var children []*Record
	for _, r := range c.Records {
		if r.Parent() == id {
			children = append(children, r)
		}
	}
	return children
}

// FindDir returns the sdd directory. root can name the sdd directory or the project that holds
// it. With no root, FindDir uses $SDD_ROOT, and then searches up from the working directory.
func FindDir(root string) (string, error) {
	if root == "" {
		root = os.Getenv("SDD_ROOT")
	}
	if root != "" {
		if isSDDDir(root) || filepath.Base(root) == "sdd" {
			return root, nil
		}
		return filepath.Join(root, "sdd"), nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if filepath.Base(dir) == "sdd" && isSDDDir(dir) {
			return dir, nil
		}
		if candidate := filepath.Join(dir, "sdd"); isSDDDir(candidate) {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no sdd/ directory here or above. Use --root")
		}
		dir = parent
	}
}

func isSDDDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "system.md")); err == nil {
		return true
	}
	info, err := os.Stat(filepath.Join(dir, "features"))
	return err == nil && info.IsDir()
}
