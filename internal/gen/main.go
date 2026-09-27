// Command gen generates partially-applied wrappers for every testify assertion.
//
// It reads the testify version this module already depends on, so the output
// cannot drift from the library being wrapped.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type param struct {
	name string
	typ  string // rendered, already qualified
}

type fn struct {
	name    string
	doc     string
	params  []param // excluding t and msgAndArgs
	hole    int
	generic bool   // hole is interface{} -> wrapper is generic in T
	ret     string // rendered return type: GenericAssertionFunc[T] or GenericAssertionFunc[<concrete>]
}

func main() {
	out := "require"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	dir := testifyDir()
	fns := collect(filepath.Join(dir, "assert"))

	if err := os.WriteFile(filepath.Join(out, "generated.go"), renderFile(fns, false), 0o644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "assertions_generated.go"), renderFile(fns, true), 0o644); err != nil {
		log.Fatal(err)
	}
	var g int
	for _, f := range fns {
		if f.generic {
			g++
		}
	}
	fmt.Printf("generated %d checks (%d generic, %d concrete) and %d Assertions[T] methods\n",
		len(fns), g, len(fns)-g, g)
}

func testifyDir() string {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/stretchr/testify").Output()
	if err != nil {
		log.Fatalf("locating testify: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func collect(assertDir string) []fn {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, assertDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var out []fn
	for _, p := range pkgs {
		for _, file := range p.Files {
			for _, decl := range file.Decls {
				d, ok := decl.(*ast.FuncDecl)
				if !ok || d.Recv != nil || !d.Name.IsExported() {
					continue
				}
				if strings.HasSuffix(d.Name.Name, "f") || skip[d.Name.Name] {
					continue
				}
				f, ok := convert(fset, d)
				if !ok {
					continue
				}
				out = append(out, f)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

func convert(fset *token.FileSet, d *ast.FuncDecl) (fn, bool) {
	ps := d.Type.Params
	if ps == nil || len(ps.List) == 0 {
		return fn{}, false
	}
	if id, ok := ps.List[0].Type.(*ast.Ident); !ok || id.Name != "TestingT" {
		return fn{}, false
	}
	if d.Type.Results == nil || len(d.Type.Results.List) != 1 {
		return fn{}, false
	}
	if id, ok := d.Type.Results.List[0].Type.(*ast.Ident); !ok || id.Name != "bool" {
		return fn{}, false
	}

	var params []param
	for _, fl := range ps.List[1:] {
		typ := render(fset, fl.Type)
		for _, n := range fl.Names {
			if n.Name == "msgAndArgs" {
				continue
			}
			params = append(params, param{name: n.Name, typ: typ})
		}
	}
	if len(params) == 0 {
		return fn{}, false
	}

	hole := holeIndex(d.Name.Name, params)
	ht := params[hole].typ
	generic := ht == "interface{}" || ht == "any"
	ret := "GenericAssertionFunc[" + ht + "]"
	if generic {
		ret = "GenericAssertionFunc[T]"
	}
	return fn{
		name:    d.Name.Name,
		doc:     doc(d),
		params:  params,
		hole:    hole,
		generic: generic,
		ret:     ret,
	}, true
}

func holeIndex(name string, params []param) int {
	if i, ok := holeOverride[name]; ok {
		if i < 0 {
			return len(params) + i
		}
		return i
	}
	for i, p := range params {
		if p.name == "actual" {
			return i
		}
	}
	for i, p := range params {
		if p.name == "object" {
			return i
		}
	}
	return 0
}

// render prints a type expression and qualifies bare exported identifiers,
// which in the assert package mean assert-package types (PanicTestFunc, ...).
func render(fset *token.FileSet, e ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, e); err != nil {
		log.Fatal(err)
	}
	s := buf.String()
	if id, ok := e.(*ast.Ident); ok && ast.IsExported(id.Name) {
		return "assert." + s
	}
	return s
}

func doc(d *ast.FuncDecl) string {
	if d.Doc == nil {
		return ""
	}
	// Keep the leading prose paragraph. Everything after the first blank
	// comment line is a testify usage example written against assert.X(t, ...),
	// which does not apply to a partially applied check.
	var lines []string
	for _, c := range d.Doc.List {
		t := strings.TrimPrefix(c.Text, "//")
		if strings.TrimSpace(t) == "" {
			break
		}
		lines = append(lines, "//"+strings.ReplaceAll(t, "assert.", ""))
	}
	return strings.Join(lines, "\n")
}
