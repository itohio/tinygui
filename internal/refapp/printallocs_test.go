package refapp

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Every heap allocation site TinyGo leaves in the library is in a function
// reviewed as setup-only (constructors, options, growth guards):
// allocs_allowlist.txt lists each function with its number of sites, so a
// new site — even inside an already listed function — fails the test until
// it is reviewed and the count updated (verify TG-1-postcommit R8). Needs
// `tinygo` on PATH (skipped otherwise); REFAPP_ALLOCS_DUMP=1 prints the
// current sites in the allowlist's format. Allocations inside runtime
// helpers (string concatenation, conversions) are reported at the runtime,
// not here: the -gc=leaking tests catch those.
func TestLibraryAllocatesOnlyInSetup(t *testing.T) {
	if runtime.Compiler != "gc" {
		t.Skip("runs tinygo itself; run with go test")
	}
	tinygo, err := exec.LookPath("tinygo")
	if err != nil {
		t.Skip("tinygo not on PATH")
	}
	root, _ := filepath.Abs("../..")
	cmd := exec.Command(tinygo, "build", "-o", filepath.Join(t.TempDir(), "refsoak"),
		"-print-allocs=github.com/itohio/tinygui", "./internal/refapp/cmd/refsoak")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tinygo build: %v\n%s", err, out)
	}
	funcs := map[string][]fn{} // per file
	count := map[string]int{}  // sites per "file function"
	seenPos := map[string]bool{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		i := strings.Index(line, ": object allocated on the heap")
		if i < 0 {
			continue
		}
		pos := line[:i] // path:line:col
		parts := strings.Split(pos, ":")
		if len(parts) < 3 {
			continue
		}
		path := strings.Join(parts[:len(parts)-2], ":")
		rel, err := filepath.Rel(root, path)
		if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "internal/refapp") {
			continue // not the library
		}
		ln, _ := strconv.Atoi(parts[len(parts)-2])
		if _, ok := funcs[rel]; !ok {
			funcs[rel] = parseFuncs(t, path)
		}
		if seenPos[pos] {
			continue
		}
		seenPos[pos] = true
		count[rel+" "+enclosing(funcs[rel], ln)]++
	}
	var keys []string
	for k := range count {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if os.Getenv("REFAPP_ALLOCS_DUMP") != "" {
		for _, k := range keys {
			t.Logf("%s %d", k, count[k])
		}
	}
	allowed := readAllowlist(t) // after the dump: it works with a broken list too
	for _, k := range keys {
		switch want, ok := allowed[k]; {
		case !ok:
			t.Errorf("heap allocation in %s (%d sites): not in allocs_allowlist.txt (remove it, or review it as setup-only and list it)", k, count[k])
		case count[k] > want:
			t.Errorf("%s: %d heap sites, %d reviewed: a new site (remove it, or review it and update the count)", k, count[k], want)
		case count[k] < want:
			t.Errorf("%s: %d heap sites, the allowlist says %d: update the count", k, count[k], want)
		}
	}
	for k := range allowed {
		if count[k] == 0 {
			t.Errorf("allocs_allowlist.txt lists %s, which has no heap site any more: remove it", k)
		}
	}
}

// go test runs the TinyGo checks too when tinygo is on PATH: the allocation
// tests under -gc=leaking (the boards' kind of allocator; plain tinygo test
// counts as well, through runtime.mallocs), so a TinyGo-only allocation
// fails a plain `go test ./internal/refapp` (verify TG-1-postcommit R8).
func TestTinyGoCountsNoAllocations(t *testing.T) {
	if runtime.Compiler != "gc" {
		t.Skip("this is the TinyGo run itself")
	}
	tinygo, err := exec.LookPath("tinygo")
	if err != nil {
		t.Skip("tinygo not on PATH")
	}
	cmd := exec.Command(tinygo, "test", "-gc=leaking", ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("tinygo test -gc=leaking ./internal/refapp: %v\n%s", err, out)
	}
}

func readAllowlist(t *testing.T) map[string]int {
	f, err := os.Open("allocs_allowlist.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m := map[string]int{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		n, err := strconv.Atoi(f[len(f)-1])
		if len(f) != 3 || err != nil || n < 1 {
			t.Fatalf("allocs_allowlist.txt: %q: want <file> <function> <sites>", sc.Text())
		}
		m[f[0]+" "+f[1]] = n
	}
	return m
}

type fn struct {
	name       string
	start, end int
}

// parseFuncs lists the file's functions and methods (Type.Method) by line.
func parseFuncs(t *testing.T, path string) []fn {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var fns []fn
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		if fd.Recv != nil && len(fd.Recv.List) > 0 {
			name = recvName(fd.Recv.List[0].Type) + "." + name
		}
		fns = append(fns, fn{name, fset.Position(fd.Pos()).Line, fset.Position(fd.End()).Line})
	}
	return fns
}

func recvName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.StarExpr:
		return recvName(x.X)
	case *ast.IndexExpr:
		return recvName(x.X)
	case *ast.IndexListExpr:
		return recvName(x.X)
	case *ast.Ident:
		return x.Name
	}
	return "?"
}

func enclosing(fns []fn, line int) string {
	for _, f := range fns {
		if line >= f.start && line <= f.end {
			return f.name
		}
	}
	return "<package level>"
}
