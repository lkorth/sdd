package record

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func loadFixture(t *testing.T) *Corpus {
	t.Helper()
	c, err := Load(filepath.Join("..", "..", "testdata", "project", "sdd"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func get(t *testing.T, c *Corpus, id string) *Record {
	t.Helper()
	records := c.ByID[id]
	if len(records) != 1 {
		t.Fatalf("%s: got %d records, want 1", id, len(records))
	}
	return records[0]
}

func TestLoadFindsEachRecordInFileOrder(t *testing.T) {
	c := loadFixture(t)
	var ids []string
	for _, r := range c.Records {
		ids = append(ids, r.ID)
	}
	want := []string{"L0_001", "L0_002", "L1_LNK_001", "L2_LNK_001", "L2_LNK_002", "L2_PER_004"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}
}

func TestFeaturesComeFromFrontmatter(t *testing.T) {
	c := loadFixture(t)
	var got [][2]string
	for _, f := range c.Features {
		got = append(got, [2]string{f.Code, f.Name})
	}
	want := [][2]string{{"L0", "System"}, {"LNK", "Link"}, {"PER", "Persistence"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("features = %v, want %v", got, want)
	}
}

func TestObligationIncludesDefinitionsBeforeTheShall(t *testing.T) {
	r := get(t, loadFixture(t), "L0_002")
	want := "A monotonic clock is a clock that never goes back. Each feature SHALL measure time with a monotonic clock."
	if got := r.ObligationText(); got != want {
		t.Errorf("obligation = %q, want %q", got, want)
	}
}

func TestRationaleRunsToTheFirstScenario(t *testing.T) {
	r := get(t, loadFixture(t), "L0_002")
	want := "The wall clock jumps when the device synchronizes it.\n\nThat jump breaks each timeout that spans it."
	if got := r.RationaleText(); got != want {
		t.Errorf("rationale = %q, want %q", got, want)
	}
}

func TestScenarioStepsJoinWrappedLines(t *testing.T) {
	r := get(t, loadFixture(t), "L2_LNK_001")
	want := []Scenario{{
		Name: "Resume after acknowledgement 40",
		Steps: []string{
			"WHEN the last acknowledgement names sample 40",
			"THEN the next upload starts at sample 41, and the archive holds samples 1 to 40 once",
		},
	}}
	if !reflect.DeepEqual(r.Scenarios, want) {
		t.Errorf("scenarios = %#v, want %#v", r.Scenarios, want)
	}
}

func TestRecordEndsAtTheNextHeading(t *testing.T) {
	r := get(t, loadFixture(t), "L2_LNK_002")
	if r.Line != 45 || r.EndLine != 58 {
		t.Errorf("lines = %d-%d, want 45-58", r.Line, r.EndLine)
	}
	last := r.Body[len(r.Body)-1]
	if last != "- THEN the link tries to reconnect within TBR_1 seconds" {
		t.Errorf("last body line = %q, so the record ran into the commentary", last)
	}
}

func TestFields(t *testing.T) {
	c := loadFixture(t)
	if r := get(t, c, "L2_LNK_002"); r.Status() != "tbr" || r.Parent() != "L1_LNK_001" {
		t.Errorf("L2_LNK_002 status, parent = %q, %q", r.Status(), r.Parent())
	}
	if r := get(t, c, "L0_001"); r.Status() != "active" || r.Parent() != "" {
		t.Errorf("L0_001 status, parent = %q, %q", r.Status(), r.Parent())
	}
}

func TestChildrenCrossFeatures(t *testing.T) {
	var ids []string
	for _, r := range loadFixture(t).Children("L1_LNK_001") {
		ids = append(ids, r.ID)
	}
	want := []string{"L2_LNK_001", "L2_LNK_002", "L2_PER_004"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("children = %v, want %v", ids, want)
	}
}

func TestLedger(t *testing.T) {
	r := loadFixture(t).Retired["L2_PER_005"]
	if r == nil {
		t.Fatal("L2_PER_005 is not in the ledger")
	}
	want := Retired{ID: "L2_PER_005", Path: r.Path, Line: 2, Reason: "Merged into L2_PER_004", Date: "2026-10-01", PR: 12}
	if *r != want {
		t.Errorf("retired = %+v, want %+v", *r, want)
	}
}

func TestFindDir(t *testing.T) {
	project, err := filepath.Abs(filepath.Join("..", "..", "testdata", "project"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(project, "sdd")
	for _, root := range []string{project, want} {
		if got, err := FindDir(root); err != nil || got != want {
			t.Errorf("FindDir(%q) = %q, %v, want %q", root, got, err, want)
		}
	}
	t.Chdir(filepath.Join(want, "features"))
	if got, err := FindDir(""); err != nil || got != want {
		t.Errorf("FindDir from features/ = %q, %v, want %q", got, err, want)
	}
}

func TestDependentsIgnoreTheRationale(t *testing.T) {
	dir := t.TempDir()
	system := `### L0_001: One

` + "```yaml\nparent: self\n```" + `

The system SHALL do one thing.

Scenario: One
- WHEN one
- THEN one

### L0_002: Two

` + "```yaml\nparent: self\n```" + `

The system SHALL do two things.

Rationale: This replaces part of L0_001.

Scenario: Two
- WHEN two
- THEN two

### L0_003: Three

` + "```yaml\nparent: self\n```" + `

The system SHALL do three things.

Scenario: Three after L0_001
- WHEN three
- THEN three
`
	if err := os.WriteFile(filepath.Join(dir, "system.md"), []byte(system), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, r := range c.Dependents("L0_001") {
		ids = append(ids, r.ID)
	}
	if !reflect.DeepEqual(ids, []string{"L0_003"}) {
		t.Errorf("dependents = %v, want [L0_003]", ids)
	}
}
