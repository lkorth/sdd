// Command sdd queries sdd requirement records without reading whole files.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/lkorth/sdd/internal/record"
)

const usage = `sdd queries sdd requirement records without reading whole files. Each answer names the
file and the line range of each record, so you can read more on demand.

Usage: sdd <command> [flags] [arguments]

Commands:
  show ID...          print records by ID. Default: the obligation
  ancestors ID        print the parent chain, nearest first. Default: the obligations
  children ID         print the records whose parent is ID. Default: --short
  list                print records, with filters. Default: --short
  search PATTERN      print each record line that matches a regular expression, ignoring case
  features            print each feature with its file and its record counts
  next LEVEL [CODE]   print the next free ID for a level and a feature
  delete ID           delete a record and add its ID to the ledger of retired IDs

Detail flags, for show, ancestors, children and list:
  --short             one line for each record: ID, location and title
  --normal            the location, the fields and the obligation paragraph
  --full              the whole record, with rationale and scenarios

Flags for each command:
  --json              print JSON, not text
  --root DIR          the sdd/ directory, or the project that holds it.
                      Default: $SDD_ROOT, then a search up from the working directory

Run "sdd <command> -h" for the flags of one command.
`

type options struct {
	json                bool
	root                string
	short, normal, full bool
	all                 bool
	level, feature      string
	status              string
	reason, date        string
	pr                  int
}

func (o *options) detail(fallback string) string {
	switch {
	case o.short:
		return "short"
	case o.full:
		return "full"
	case o.normal:
		return "normal"
	}
	return fallback
}

type command struct {
	args    string
	help    string
	detail  bool
	minArgs int
	maxArgs int // -1 for no limit
	run     func(c *record.Corpus, o *options, args []string, out *printer) int
}

var commands = map[string]command{
	"show":      {"ID...", "Print records by ID. Default: the obligation.", true, 1, -1, cmdShow},
	"ancestors": {"ID", "Print the parent chain of a record, nearest first. Default: the obligations.", true, 1, 1, cmdAncestors},
	"children":  {"ID", "Print the records whose parent is ID. Default: --short.", true, 1, 1, cmdChildren},
	"list":      {"", "Print records, with filters. Default: --short.", true, 0, 0, cmdList},
	"search":    {"PATTERN", "Print each record line that matches a regular expression, ignoring case.", false, 1, 1, cmdSearch},
	"features":  {"", "Print each feature with its file and its record counts.", false, 0, 0, cmdFeatures},
	"next":      {"LEVEL [CODE]", "Print the next free ID for a level and a feature. L0 takes no code.", false, 1, 2, cmdNext},
	"delete":    {"ID", "Delete a record and add its ID to the ledger of retired IDs. Refuses if another record names the ID as its parent, in its obligation or in a scenario.", false, 1, 1, cmdDelete},
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	name := args[0]
	cmd, ok := commands[name]
	if !ok {
		fmt.Fprintf(stderr, "sdd: unknown command %q\n\n%s", name, usage)
		return 2
	}

	var o options
	fs := flag.NewFlagSet("sdd "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.BoolVar(&o.json, "json", false, "print JSON, not text")
	fs.StringVar(&o.root, "root", "", "the sdd/ `directory`, or the project that holds it")
	if cmd.detail {
		fs.BoolVar(&o.short, "short", false, "one line for each record: ID, location and title")
		fs.BoolVar(&o.normal, "normal", false, "the location, the fields and the obligation paragraph")
		fs.BoolVar(&o.full, "full", false, "the whole record, with rationale and scenarios")
	}
	switch name {
	case "children":
		fs.BoolVar(&o.all, "all", false, "all descendants, not only children")
	case "list":
		fs.StringVar(&o.level, "level", "", "only this `level`: L0, L1 or L2")
		fs.StringVar(&o.feature, "feature", "", "only this feature `code`, or L0 for system.md")
		fs.StringVar(&o.status, "status", "", "only this `status`: active or tbr")
	case "delete":
		fs.StringVar(&o.reason, "reason", "", "why the record is gone, for the ledger")
		fs.StringVar(&o.date, "date", "", "the `date` for the ledger, as YYYY-MM-DD. Default: today")
		fs.IntVar(&o.pr, "pr", 0, "the pull request `number` for the ledger")
	}
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: sdd %s [flags] %s\n\n%s\n\nFlags:\n", name, cmd.args, cmd.help)
		fs.PrintDefaults()
	}

	flags, positional := splitArgs(args[1:], fs)
	if err := fs.Parse(flags); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	positional = append(positional, fs.Args()...)
	if len(positional) < cmd.minArgs || (cmd.maxArgs >= 0 && len(positional) > cmd.maxArgs) {
		fmt.Fprintf(stderr, "Usage: sdd %s [flags] %s\n", name, cmd.args)
		return 2
	}
	if countTrue(o.short, o.normal, o.full) > 1 {
		fmt.Fprintln(stderr, "sdd: use only one of --short, --normal and --full")
		return 2
	}
	if msg := checkFilters(&o); msg != "" {
		fmt.Fprintln(stderr, "sdd: "+msg)
		return 2
	}

	dir, err := record.FindDir(o.root)
	if err != nil {
		fmt.Fprintln(stderr, "sdd: "+err.Error())
		return 2
	}
	corpus, err := record.Load(dir)
	if err != nil {
		fmt.Fprintln(stderr, "sdd: "+err.Error())
		return 2
	}
	out := newPrinter(stdout, stderr, o.json)
	code := cmd.run(corpus, &o, positional, out)
	if out.err != nil {
		fmt.Fprintln(stderr, "sdd: "+out.err.Error())
		return 2
	}
	return code
}

// splitArgs separates flags from positional arguments, so that flags can come after them:
// "sdd show L2_LNK_001 --full".
func splitArgs(args []string, fs *flag.FlagSet) (flags, positional []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return flags, append(positional, args[i+1:]...)
		}
		if len(arg) < 2 || arg[0] != '-' {
			positional = append(positional, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if strings.Contains(name, "=") {
			continue
		}
		if f := fs.Lookup(name); f != nil && !isBoolFlag(f) && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return flags, positional
}

func isBoolFlag(f *flag.Flag) bool {
	b, ok := f.Value.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}

var isoDate = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func countTrue(values ...bool) int {
	n := 0
	for _, v := range values {
		if v {
			n++
		}
	}
	return n
}

func checkFilters(o *options) string {
	if o.level != "" && o.level != "L0" && o.level != "L1" && o.level != "L2" {
		return "--level is L0, L1 or L2"
	}
	if o.status != "" && o.status != "active" && o.status != "tbr" {
		return "--status is active or tbr"
	}
	if o.date != "" && !isoDate.MatchString(o.date) {
		return "--date is YYYY-MM-DD"
	}
	if o.pr < 0 {
		return "--pr is a positive number"
	}
	return ""
}
