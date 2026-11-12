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

// Every heap allocation TinyGo leaves in the library is in a function
// reviewed as setup-only (constructors, options, first-draw buffer sizing,
// …): allocs_allowlist.txt. A new site fails the test until it is removed
// or reviewed and listed. Needs `tinygo` on PATH (skipped otherwise);
// REFAPP_ALLOCS_DUMP=1 prints the current sites in the allowlist's format.
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
	allowed := readAllowlist(t)
	funcs := map[string][]fn{} // per file
	var sites []string
	seen := map[string]bool{}
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
		key := rel + " " + enclosing(funcs[rel], ln)
		if seen[key] {
			continue
		}
		seen[key] = true
		sites = append(sites, key)
	}
	sort.Strings(sites)
	if os.Getenv("REFAPP_ALLOCS_DUMP") != "" {
		for _, s := range sites {
			t.Log(s)
		}
	}
	for _, s := range sites {
		if !allowed[s] {
			t.Errorf("heap allocation in %s: not in allocs_allowlist.txt (remove it, or review it as setup-only and list it)", s)
		}
	}
}

func readAllowlist(t *testing.T) map[string]bool {
	f, err := os.Open("allocs_allowlist.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		if f := strings.Fields(line); len(f) == 2 {
			m[f[0]+" "+f[1]] = true
		}
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
