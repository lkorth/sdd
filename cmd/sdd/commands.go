package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lkorth/sdd/internal/record"
)

var featureCode = regexp.MustCompile(`^[A-Z]{2,4}$`)

// lookup returns the records with an ID, or its retired entry, or a missing item.
func lookup(c *record.Corpus, id string) []item {
	if records, ok := c.ByID[id]; ok {
		items := make([]item, len(records))
		for i, r := range records {
			items[i] = item{rec: r}
		}
		return items
	}
	if retired, ok := c.Retired[id]; ok {
		return []item{{retired: retired}}
	}
	return []item{{missing: id}}
}

func cmdShow(c *record.Corpus, o *options, args []string, out *printer) int {
	var items []item
	for _, id := range args {
		items = append(items, lookup(c, id)...)
	}
	out.items(items, o.detail("normal"))
	return exitFor(items)
}

func cmdAncestors(c *record.Corpus, o *options, args []string, out *printer) int {
	start := lookup(c, args[0])[0]
	if start.rec == nil {
		out.items([]item{start}, "short")
		return 1
	}
	var items []item
	seen := map[string]bool{args[0]: true}
	for r := start.rec; r.Parent() != "" && !seen[r.Parent()]; {
		seen[r.Parent()] = true
		found := lookup(c, r.Parent())[0]
		items = append(items, found)
		if found.rec == nil {
			break
		}
		r = found.rec
	}
	out.items(items, o.detail("normal"))
	return exitFor(items)
}

func cmdChildren(c *record.Corpus, o *options, args []string, out *printer) int {
	start := lookup(c, args[0])[0]
	if start.rec == nil {
		out.items([]item{start}, "short")
		return 1
	}
	var items []item
	seen := map[string]bool{args[0]: true}
	var walk func(id string, depth int)
	walk = func(id string, depth int) {
		for _, child := range c.Children(id) {
			if seen[child.ID] {
				continue
			}
			seen[child.ID] = true
			it := item{rec: child}
			if o.all {
				it.depth = depth
			}
			items = append(items, it)
			if o.all {
				walk(child.ID, depth+1)
			}
		}
	}
	walk(args[0], 1)
	out.items(items, o.detail("short"))
	return 0
}

func cmdList(c *record.Corpus, o *options, _ []string, out *printer) int {
	var items []item
	for _, r := range c.Records {
		feature := r.Feature()
		if feature == "" {
			feature = "L0"
		}
		if (o.level == "" || r.Level() == o.level) &&
			(o.feature == "" || feature == o.feature) &&
			(o.status == "" || r.Status() == o.status) {
			items = append(items, item{rec: r})
		}
	}
	out.items(items, o.detail("short"))
	return 0
}

type searchHit struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	File  string `json:"file"`
	Line  int    `json:"line"`
	Match string `json:"match"`
}

func cmdSearch(c *record.Corpus, _ *options, args []string, out *printer) int {
	pattern, err := regexp.Compile("(?i)" + args[0])
	if err != nil {
		out.err = err
		return 2
	}
	hits := []searchHit{}
	for _, r := range c.Records {
		check := func(line int, text string) {
			if pattern.MatchString(text) {
				hits = append(hits, searchHit{r.ID, r.Title, out.path(r.Path), line, strings.TrimSpace(text)})
			}
		}
		check(r.Line, r.ID+": "+r.Title)
		for n, text := range r.Body {
			check(r.Line+1+n, text)
		}
	}
	if out.json {
		out.writeJSON(hits)
	} else {
		for _, h := range hits {
			out.printf("%s  %s:%d  %s\n", h.ID, h.File, h.Line, h.Match)
		}
	}
	if len(hits) == 0 {
		return 1
	}
	return 0
}

type featureRow struct {
	Code    string         `json:"code"`
	Name    string         `json:"name"`
	File    string         `json:"file"`
	Records map[string]int `json:"records"`
}

func cmdFeatures(c *record.Corpus, _ *options, _ []string, out *printer) int {
	rows := []featureRow{}
	for _, f := range c.Features {
		counts := map[string]int{}
		for _, r := range f.Records {
			counts[r.Level()]++
		}
		rows = append(rows, featureRow{f.Code, f.Name, out.path(f.Path), counts})
	}
	if out.json {
		out.writeJSON(rows)
		return 0
	}
	for _, row := range rows {
		levels := make([]string, 0, len(row.Records))
		for level := range row.Records {
			levels = append(levels, level)
		}
		sort.Strings(levels)
		parts := make([]string, len(levels))
		for i, level := range levels {
			parts[i] = fmt.Sprintf("%d %s", row.Records[level], level)
		}
		counts := strings.Join(parts, ", ")
		if counts == "" {
			counts = "no records"
		}
		out.printf("%s  %s  %s  (%s)\n", row.Code, row.File, row.Name, counts)
	}
	return 0
}

func cmdNext(c *record.Corpus, _ *options, args []string, out *printer) int {
	level, code := args[0], ""
	if len(args) > 1 {
		code = args[1]
	}
	var prefix string
	switch {
	case level == "L0" && code == "":
		prefix = "L0_"
	case (level == "L1" || level == "L2") && featureCode.MatchString(code):
		prefix = level + "_" + code + "_"
	default:
		out.err = fmt.Errorf("next takes L0, or L1 or L2 with a feature code: next L2 LNK")
		return 2
	}
	highest := 0
	check := func(id string) {
		if strings.HasPrefix(id, prefix) && record.ValidID.MatchString(id) {
			if n, err := strconv.Atoi(id[len(prefix):]); err == nil && n > highest {
				highest = n
			}
		}
	}
	for id := range c.ByID {
		check(id)
	}
	for id := range c.Retired {
		check(id)
	}
	next := fmt.Sprintf("%s%03d", prefix, highest+1)
	if out.json {
		out.writeJSON(map[string]string{"id": next})
	} else {
		out.printf("%s\n", next)
	}
	return 0
}

type deleteResult struct {
	ID      string   `json:"id"`
	Deleted jsonItem `json:"deleted"`
	Retired jsonItem `json:"retired"`
}

func cmdDelete(c *record.Corpus, o *options, args []string, out *printer) int {
	id := args[0]
	records := c.ByID[id]
	switch {
	case len(records) == 0 && c.Retired[id] != nil:
		out.fail("%s is already retired", id)
		return 1
	case len(records) == 0:
		out.fail("%s: no record with this ID", id)
		return 1
	case len(records) > 1:
		out.fail("%s is the ID of %d records. Give one of them a new ID first", id, len(records))
		return 1
	}
	if dependents := c.Dependents(id); len(dependents) > 0 {
		lines := make([]string, len(dependents))
		for i, r := range dependents {
			lines[i] = out.toText(item{rec: r}, "short")
		}
		out.fail("%s has dependents. Change or delete them first:\n%s", id, strings.Join(lines, "\n"))
		return 1
	}

	r := records[0]
	deleted := item{rec: r}
	where := out.where(deleted)
	entry := record.Retired{ID: id, Reason: o.reason, Date: o.date, PR: o.pr}
	if entry.Date == "" {
		entry.Date = time.Now().Format("2006-01-02")
	}
	path, line, err := record.AppendRetired(c.Dir, entry)
	if err != nil {
		out.err = err
		return 2
	}
	if err := record.RemoveRecord(r); err != nil {
		out.err = err
		return 2
	}
	entry.Path, entry.Line = path, line
	retired := item{retired: &entry}

	if out.json {
		out.writeJSON(deleteResult{id, out.toJSON(deleted, "short"), out.toJSON(retired, "")})
		return 0
	}
	out.printf("%s  %s  deleted\n%s\n", id, where, out.toText(retired, ""))
	return 0
}

func exitFor(items []item) int {
	for _, it := range items {
		if it.rec == nil && it.retired == nil {
			return 1
		}
	}
	return 0
}
