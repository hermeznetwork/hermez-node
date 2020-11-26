package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// "github.com/davecgh/go-spew/spew"
	"golang.org/x/tools/go/ast/astutil"
)

func wrapErr(c *astutil.Cursor, e ast.Expr) {
	wrapped := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "tracerr"},
			Sel: &ast.Ident{Name: "Wrap"},
		},
		// LParen:
		Args: []ast.Expr{
			e,
		},
		// Ellipsis:
		// Rparen:
	}
	c.Replace(wrapped)
}

func unwrapErr(c *astutil.Cursor, e ast.Expr) {
	unwrapped := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "tracerr"},
			Sel: &ast.Ident{Name: "Unwrap"},
		},
		// LParen:
		Args: []ast.Expr{
			e,
		},
		// Ellipsis:
		// Rparen:
	}
	c.Replace(unwrapped)
}

// isLastResultReturn returns true if the cursor is in the last element of a
// return statement
func isLastResultReturn(c *astutil.Cursor) bool {
	node := c.Node()
	parent := c.Parent()
	if p, ok := parent.(*ast.ReturnStmt); ok {
		if p.Results[len(p.Results)-1] == node {
			return true
		}
	}
	return false
}

// isBinCmpWithNotNil returns true if the cursor is in a binary comparison with
// something that is not nil
func isBoolCmpWithNotNil(c *astutil.Cursor) bool {
	node := c.Node()
	parent := c.Parent()
	if p, ok := parent.(*ast.BinaryExpr); ok {
		if p.X == node {
			if y, ok := p.Y.(*ast.Ident); ok {
				if y.Name == "nil" {
					return false
				}
			}
			if p.Op == token.EQL || p.Op == token.NEQ {
				return true
			}
		}
	}
	return false
}

func isAssertRequireNotNil(c *astutil.Cursor) bool {
	node := c.Node()
	parent := c.Parent()
	if p, ok := parent.(*ast.CallExpr); ok {
		if isCallExpr(p, "assert", "Equal") ||
			isCallExpr(p, "require", "Equal") ||
			isCallExpr(p, "assert", "NotEqual") ||
			isCallExpr(p, "require", "NotEqual") {
			if a1, ok := p.Args[1].(*ast.Ident); ok {
				if a1.Name == "nil" {
					return false
				}
			}
			if p.Args[2] == node {
				return true
			}
		}
	}
	return false
}

// isCallExpr returns true if the CallExpr matches `pkgName.fnName`
func isCallExpr(n *ast.CallExpr, pkgName, fnName string) bool {
	fun := n.Fun
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		if x, ok := f.X.(*ast.Ident); ok {
			if x.Name == pkgName && f.Sel.Name == fnName {
				return true
			}
		}
	}
	return false
}

type instrument struct {
	updated bool
}

func (i *instrument) post(c *astutil.Cursor) bool {
	node := c.Node()
	switch n := node.(type) {
	case *ast.Ident:
		if strings.HasPrefix(strings.ToLower(n.Name), "err") {
			if isLastResultReturn(c) {
				wrapErr(c, n)
				i.updated = true
			} else if isBoolCmpWithNotNil(c) {
				unwrapErr(c, n)
				i.updated = true
			} else if isAssertRequireNotNil(c) {
				unwrapErr(c, n)
				i.updated = true
			}
		}
	case *ast.CallExpr:
		if isCallExpr(n, "fmt", "Errorf") ||
			isCallExpr(n, "errors", "New") {
			if isLastResultReturn(c) {
				wrapErr(c, n)
				i.updated = true
			}
		}
	}
	return true
}

func instrumentSrc(filename string, oldSource []byte) ([]byte, error) {
	fset := token.NewFileSet()
	oldAST, err := parser.ParseFile(fset, filename, oldSource, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", filename, err)
	}
	inst := instrument{}
	newAST := astutil.Apply(oldAST, nil, inst.post)
	if inst.updated {
		astutil.AddImport(fset, oldAST, "github.com/ztrue/tracerr")
	}

	buf := &bytes.Buffer{}
	err = format.Node(buf, fset, newAST)
	if err != nil {
		return nil, fmt.Errorf("error formatting new code: %w", err)
	}
	// fmt.Println()
	// spew.Dump(oldAST)
	return buf.Bytes(), nil
}

func main() {
	if err := filepath.Walk("../..",
		func(filename string, info os.FileInfo, err error) error {
			if strings.HasPrefix(filename, "../../scripts") {
				return nil
			}
			if !strings.HasSuffix(filename, ".go") {
				return nil
			}
			// fmt.Println(filename)
			oldSource, err := ioutil.ReadFile(filename) //nolint:gosec
			if err != nil {
				return fmt.Errorf("couldn't read from %s: %v", filename, err)
			}
			newSource, err := instrumentSrc(filename, oldSource)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(filename, newSource, info.Mode()); err != nil {
				return err
			}
			// fmt.Print(string(newSource))
			return nil
		},
	); err != nil {
		panic(err)
	}
}
