package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// sdd runs the command in the fixture project and returns stdout, stderr and the exit code.
func sdd(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	t.Chdir(filepath.Join("..", "..", "testdata", "project"))
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func expect(t *testing.T, args []string, wantOut string, wantCode int) {
	t.Helper()
	out, errOut, code := sdd(t, args...)
	if out != wantOut || code != wantCode {
		t.Errorf("sdd %s\n got (exit %d):\n%s\nwant (exit %d):\n%s\nstderr: %s",
			strings.Join(args, " "), code, out, wantCode, wantOut, errOut)
	}
}

func TestShowPrintsTheObligation(t *testing.T) {
	expect(t, []string{"show", "L2_PER_004"}, `L2_PER_004: Acknowledgement is durable before deletion
sdd/features/persistence.md:11-25  parent L1_LNK_001
The system SHALL write an upload acknowledgement to durable storage before deleting the
acknowledged record.
`, 0)
}

func TestShowShortWithFlagAfterTheID(t *testing.T) {
	expect(t, []string{"show", "L2_LNK_002", "--short"},
		"L2_LNK_002  sdd/features/link.md:45-58  The reconnect delay is bounded  (tbr)\n", 0)
}

func TestShowFullPrintsTheWholeRecord(t *testing.T) {
	out, _, code := sdd(t, "show", "--full", "L0_002")
	if code != 0 || !strings.HasSuffix(out, "- THEN the timeout ends at its configured duration\n") ||
		!strings.Contains(out, "That jump breaks each timeout that spans it.") {
		t.Errorf("exit %d, output:\n%s", code, out)
	}
}

func TestShowRetiredAndMissing(t *testing.T) {
	expect(t, []string{"show", "L2_PER_005", "L1_LNK_009"}, `L2_PER_005  sdd/retired.jsonl:2  retired: Merged into L2_PER_004
---
L1_LNK_009  missing: no record and no retired entry
`, 1)
}

func TestAncestorsNearestFirst(t *testing.T) {
	expect(t, []string{"ancestors", "--short", "L2_PER_004"}, `L1_LNK_001  sdd/features/link.md:11-25  An interrupted upload completes without duplicate delivery
L0_001  sdd/system.md:5-18  Recorded data reaches the archive
`, 0)
}

func TestChildrenAllIsATree(t *testing.T) {
	expect(t, []string{"children", "L0_001", "--all"}, `L1_LNK_001  sdd/features/link.md:11-25  An interrupted upload completes without duplicate delivery
  L2_LNK_001  sdd/features/link.md:29-43  An interrupted upload resumes after the last acknowledgement
  L2_LNK_002  sdd/features/link.md:45-58  The reconnect delay is bounded  (tbr)
  L2_PER_004  sdd/features/persistence.md:11-25  Acknowledgement is durable before deletion
`, 0)
}

func TestListFilters(t *testing.T) {
	expect(t, []string{"list", "--feature", "L0"}, `L0_001  sdd/system.md:5-18  Recorded data reaches the archive
L0_002  sdd/system.md:20-35  No feature reads the wall clock
`, 0)
	expect(t, []string{"list", "--status", "tbr", "--level", "L1"}, "", 0)
}

func TestSearch(t *testing.T) {
	expect(t, []string{"search", "TBR_1 seconds"}, `L2_LNK_002  sdd/features/link.md:52  The link SHALL try to reconnect within TBR_1 seconds of a drop.
L2_LNK_002  sdd/features/link.md:58  - THEN the link tries to reconnect within TBR_1 seconds
`, 0)
	expect(t, []string{"search", "commentary"}, "", 1)
}

func TestFeatures(t *testing.T) {
	expect(t, []string{"features"}, `L0  sdd/system.md  System  (2 L0)
LNK  sdd/features/link.md  Link  (1 L1, 2 L2)
PER  sdd/features/persistence.md  Persistence  (1 L2)
`, 0)
}

func TestNextCountsTheLedger(t *testing.T) {
	expect(t, []string{"next", "L2", "PER"}, "L2_PER_006\n", 0)
	expect(t, []string{"next", "L0"}, "L0_003\n", 0)
	expect(t, []string{"next", "L1", "NEW"}, "L1_NEW_001\n", 0)
	expect(t, []string{"next", "L2"}, "", 2)
}

func TestJSON(t *testing.T) {
	out, _, code := sdd(t, "show", "L2_LNK_001", "L2_PER_003", "--full", "--json")
	var items []map[string]any
	if err := json.Unmarshal([]byte(out), &items); err != nil || code != 0 || len(items) != 2 {
		t.Fatalf("exit %d, err %v, output:\n%s", code, err, out)
	}
	if items[0]["rationale"] != "The acknowledgement is the only fact both ends agree on." ||
		items[0]["end_line"] != 43.0 || items[0]["feature"] != "LNK" {
		t.Errorf("record = %v", items[0])
	}
	if items[1]["retired"] != true || items[1]["reason"] != "Replaced by a shorter timeout" {
		t.Errorf("retired = %v", items[1])
	}
}

func TestEmptyJSONIsAnArray(t *testing.T) {
	expect(t, []string{"list", "--status", "tbr", "--level", "L1", "--json"}, "[]\n", 0)
}

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"show"}, {"nope"}, {"list", "--short", "--full"}, {"list", "--level", "L3"}} {
		if _, _, code := sdd(t, args...); code != 2 {
			t.Errorf("sdd %s: exit %d, want 2", strings.Join(args, " "), code)
		}
	}
}

// sddCopy runs the command in a fresh copy of the fixture project, for commands that write.
func sddCopy(t *testing.T, args ...string) (dir, stdout, stderr string, code int) {
	t.Helper()
	dir = t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(filepath.Join("..", "..", "testdata", "project"))); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return dir, out.String(), errOut.String(), code
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestDeleteRemovesTheRecordAndRetiresTheID(t *testing.T) {
	dir, out, errOut, code := sddCopy(t, "delete", "L2_LNK_002", "--reason", "Folded into L2_LNK_001", "--pr", "7", "--date", "2026-10-13")
	want := "L2_LNK_002  sdd/features/link.md:45-58  deleted\nL2_LNK_002  sdd/retired.jsonl:3  retired: Folded into L2_LNK_001\n"
	if out != want || code != 0 {
		t.Fatalf("exit %d, stdout:\n%s\nstderr: %s", code, out, errOut)
	}
	link := readFile(t, filepath.Join(dir, "sdd", "features", "link.md"))
	if strings.Contains(link, "L2_LNK_002") || !strings.Contains(link, "  holds samples 1 to 40 once\n\n## Notes\n") {
		t.Errorf("link.md after delete:\n%s", link)
	}
	ledger := readFile(t, filepath.Join(dir, "sdd", "retired.jsonl"))
	if !strings.HasSuffix(ledger, "\n"+`{"id": "L2_LNK_002", "reason": "Folded into L2_LNK_001", "date": "2026-10-13", "pr": 7}`+"\n") {
		t.Errorf("ledger:\n%s", ledger)
	}
	var stdout bytes.Buffer
	if run([]string{"show", "L2_LNK_002", "--short"}, &stdout, io.Discard); stdout.String() != "L2_LNK_002  sdd/retired.jsonl:3  retired: Folded into L2_LNK_001\n" {
		t.Errorf("show after delete: %q", stdout.String())
	}
}

func TestDeleteTheLastRecordOfAFile(t *testing.T) {
	dir, _, errOut, code := sddCopy(t, "delete", "L2_PER_004", "--date", "2026-10-13")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if got := readFile(t, filepath.Join(dir, "sdd", "features", "persistence.md")); !strings.HasSuffix(got, "Durable storage on the device.\n\n## L2\n") {
		t.Errorf("persistence.md after delete:\n%q", got)
	}
}

func TestDeleteRefusesAndChangesNothing(t *testing.T) {
	original, err := filepath.Abs(filepath.Join("..", "..", "testdata", "project", "sdd"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ id, msg string }{
		{"L1_LNK_001", "L1_LNK_001 has dependents"},
		{"L2_PER_003", "already retired"},
		{"L2_ZZ_001", "no record"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			dir, _, errOut, code := sddCopy(t, "delete", tc.id)
			if code != 1 || !strings.Contains(errOut, tc.msg) {
				t.Errorf("exit %d, stderr %q", code, errOut)
			}
			for _, name := range []string{"retired.jsonl", "features/link.md", "system.md"} {
				if readFile(t, filepath.Join(dir, "sdd", name)) != readFile(t, filepath.Join(original, name)) {
					t.Errorf("changed %s", name)
				}
			}
		})
	}
}

// recordText returns the text of a small record, for projects that a test writes.
func recordText(id, parent string) string {
	return "### " + id + ": Title of " + id + "\n\n```yaml\nparent: " + parent + "\n```\n\n" +
		"The system SHALL do " + id + ".\n\nScenario: " + id + "\n- WHEN it runs\n- THEN it does " + id + "\n"
}

// writeProject writes files below a new project directory and returns the directory.
func writeProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		path := filepath.Join(dir, "sdd", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// sddIn runs the command with dir as the working directory.
func sddIn(t *testing.T, dir string, args ...string) (string, string, int) {
	t.Helper()
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func TestDeleteRefusesADuplicateID(t *testing.T) {
	system := recordText("L0_001", "self") + "\n" + recordText("L0_001", "self")
	dir := writeProject(t, map[string]string{"system.md": system})
	_, errOut, code := sddIn(t, dir, "delete", "L0_001")
	if code != 1 || !strings.Contains(errOut, "is the ID of 2 records") {
		t.Errorf("exit %d, stderr %q", code, errOut)
	}
	if got := readFile(t, filepath.Join(dir, "sdd", "system.md")); got != system {
		t.Errorf("system.md changed:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "sdd", "retired.jsonl")); !os.IsNotExist(err) {
		t.Errorf("the refusal wrote a ledger: %v", err)
	}
}

func TestDeleteKeepsCRLF(t *testing.T) {
	system := strings.ReplaceAll("# System\n\n"+recordText("L0_001", "self")+"\n"+recordText("L0_002", "self"), "\n", "\r\n")
	dir := writeProject(t, map[string]string{"system.md": system})
	if _, errOut, code := sddIn(t, dir, "delete", "L0_001"); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	want := strings.ReplaceAll("# System\n\n"+recordText("L0_002", "self"), "\n", "\r\n")
	if got := readFile(t, filepath.Join(dir, "sdd", "system.md")); got != want {
		t.Errorf("system.md = %q, want %q", got, want)
	}
}

func TestDeleteAddsALineToALedgerWithNoFinalNewline(t *testing.T) {
	dir := writeProject(t, map[string]string{
		"system.md":     recordText("L0_001", "self") + "\n" + recordText("L0_002", "self"),
		"retired.jsonl": `{"id": "L0_003"}`,
	})
	out, errOut, code := sddIn(t, dir, "delete", "L0_002", "--date", "2026-10-14")
	if code != 0 || !strings.Contains(out, "sdd/retired.jsonl:2  retired\n") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	want := `{"id": "L0_003"}` + "\n" + `{"id": "L0_002", "date": "2026-10-14"}` + "\n"
	if got := readFile(t, filepath.Join(dir, "sdd", "retired.jsonl")); got != want {
		t.Errorf("ledger = %q, want %q", got, want)
	}
}

func TestDeleteCreatesTheLedger(t *testing.T) {
	dir := writeProject(t, map[string]string{"system.md": recordText("L0_001", "self")})
	if _, errOut, code := sddIn(t, dir, "delete", "L0_001", "--reason", "Not needed", "--date", "2026-10-14"); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	want := `{"id": "L0_001", "reason": "Not needed", "date": "2026-10-14"}` + "\n"
	if got := readFile(t, filepath.Join(dir, "sdd", "retired.jsonl")); got != want {
		t.Errorf("ledger = %q, want %q", got, want)
	}
	if got := readFile(t, filepath.Join(dir, "sdd", "system.md")); got != "" {
		t.Errorf("system.md = %q, want an empty file", got)
	}
}

func TestBadInputExitsWith2AndChangesNothing(t *testing.T) {
	system := recordText("L0_001", "self")
	for _, tc := range []struct {
		name, ledger, msg string
		args              []string
	}{
		{"date", "", "--date is YYYY-MM-DD", []string{"delete", "L0_001", "--date", "14-10-2026"}},
		{"pr", "", "--pr is a positive number", []string{"delete", "L0_001", "--pr", "-3"}},
		{"pattern", "", "error parsing regexp", []string{"search", "(unclosed"}},
		{"ledger", "{\"id\": \"L0_002\"}\nnot json\n", "retired.jsonl:2:", []string{"delete", "L0_001"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"system.md": system}
			if tc.ledger != "" {
				files["retired.jsonl"] = tc.ledger
			}
			dir := writeProject(t, files)
			_, errOut, code := sddIn(t, dir, tc.args...)
			if code != 2 || !strings.Contains(errOut, tc.msg) {
				t.Errorf("exit %d, stderr %q, want exit 2 and %q", code, errOut, tc.msg)
			}
			if got := readFile(t, filepath.Join(dir, "sdd", "system.md")); got != system {
				t.Errorf("system.md changed:\n%s", got)
			}
			ledger := filepath.Join(dir, "sdd", "retired.jsonl")
			if tc.ledger == "" {
				if _, err := os.Stat(ledger); !os.IsNotExist(err) {
					t.Errorf("wrote a ledger: %v", err)
				}
			} else if got := readFile(t, ledger); got != tc.ledger {
				t.Errorf("ledger changed: %q", got)
			}
		})
	}
}

func TestNoSDDDirectoryExitsWith2(t *testing.T) {
	t.Setenv("SDD_ROOT", "")
	_, errOut, code := sddIn(t, t.TempDir(), "list")
	if code != 2 || !strings.Contains(errOut, "no sdd/ directory") {
		t.Errorf("exit %d, stderr %q", code, errOut)
	}
}

func TestAncestorsStopsAtAMissingParent(t *testing.T) {
	dir := writeProject(t, map[string]string{
		"system.md":        recordText("L0_001", "self"),
		"features/link.md": "---\nfeature: LNK\n---\n\n## L2\n\n" + recordText("L2_LNK_001", "L1_LNK_001") + "\n" + recordText("L2_LNK_002", "L2_LNK_001"),
	})
	out, _, code := sddIn(t, dir, "ancestors", "--short", "L2_LNK_002")
	want := "L2_LNK_001  sdd/features/link.md:7-17  Title of L2_LNK_001\nL1_LNK_001  missing: no record and no retired entry\n"
	if out != want || code != 1 {
		t.Errorf("exit %d, output:\n%s\nwant (exit 1):\n%s", code, out, want)
	}
}

func TestAncestorsStopsInAParentCycle(t *testing.T) {
	dir := writeProject(t, map[string]string{
		"features/link.md": "---\nfeature: LNK\n---\n\n## L1\n\n" + recordText("L1_LNK_001", "L1_LNK_002") + "\n" + recordText("L1_LNK_002", "L1_LNK_001"),
	})
	t.Chdir(dir)
	done := make(chan struct{})
	var stdout bytes.Buffer
	var code int
	go func() {
		defer close(done)
		code = run([]string{"ancestors", "--short", "L1_LNK_001"}, &stdout, io.Discard)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ancestors did not stop in a parent cycle")
	}
	want := "L1_LNK_002  sdd/features/link.md:19-29  Title of L1_LNK_002\n"
	if out := stdout.String(); out != want || code != 0 {
		t.Errorf("exit %d, output:\n%s\nwant:\n%s", code, out, want)
	}
}

func TestRootFromOutsideTheProject(t *testing.T) {
	project, err := filepath.Abs(filepath.Join("..", "..", "testdata", "project"))
	if err != nil {
		t.Fatal(err)
	}
	want := "L0_001  " + filepath.Join(project, "sdd", "system.md") + ":5-18  Recorded data reaches the archive\n"
	elsewhere := t.TempDir()

	t.Run("flag", func(t *testing.T) {
		t.Setenv("SDD_ROOT", "")
		if out, errOut, code := sddIn(t, elsewhere, "show", "--short", "--root", project, "L0_001"); out != want || code != 0 {
			t.Errorf("exit %d, stdout %q, stderr %q, want %q", code, out, errOut, want)
		}
	})
	t.Run("environment", func(t *testing.T) {
		t.Setenv("SDD_ROOT", filepath.Join(project, "sdd"))
		if out, errOut, code := sddIn(t, elsewhere, "show", "--short", "L0_001"); out != want || code != 0 {
			t.Errorf("exit %d, stdout %q, stderr %q, want %q", code, out, errOut, want)
		}
	})
}
