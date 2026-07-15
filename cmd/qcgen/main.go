package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type srcSpec struct {
	path    string
	pattern string
}

type typeInfo struct {
	Name      string
	PkgPath   string
	PkgName   string
	HasSQLNam bool
}

type genConfig struct {
	srcs     []srcSpec
	dest     string
	pkg      string
	noGenStr bool
}

func main() {
	var srcFlags multiFlag
	var dest string
	var pkg string
	var noGenStr bool
	var help bool

	flag.Var(&srcFlags, "src", "Source specification (path:pattern), can be repeated")
	flag.StringVar(&dest, "dest", "", "Destination directory (required)")
	flag.StringVar(&pkg, "pkg", "", "Package name (default: dest directory name)")
	flag.BoolVar(&noGenStr, "no-genstr", false, "Don't add //go:generate comment")
	flag.BoolVar(&help, "help", false, "Show help")

	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if len(srcFlags) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one -src flag is required")
		printHelp()
		os.Exit(1)
	}

	if dest == "" {
		fmt.Fprintln(os.Stderr, "Error: -dest flag is required")
		printHelp()
		os.Exit(1)
	}

	srcs := make([]srcSpec, 0, len(srcFlags))
	for _, s := range srcFlags {
		spec := parseSrcSpec(s)
		srcs = append(srcs, spec)
	}

	if pkg == "" {
		pkg = filepath.Base(dest)
	}

	cfg := genConfig{
		srcs:     srcs,
		dest:     dest,
		pkg:      pkg,
		noGenStr: noGenStr,
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("qcgen - Code generator for typed table implementations")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  qcgen -src=<path:pattern> [-src=...] -dest=<dir> [-pkg=<name>] [-no-genstr]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -src        Source specification (can be repeated)")
	fmt.Println("              Format: path:pattern (both optional)")
	fmt.Println("              path    - directory containing Go files (default: .)")
	fmt.Println("              pattern - glob pattern for type names (default: *)")
	fmt.Println("              Examples:")
	fmt.Println("                -src=../model:*Row")
	fmt.Println("                -src=../model:")
	fmt.Println("                -src=:*Row")
	fmt.Println("                -src=")
	fmt.Println("  -dest       Destination directory (required)")
	fmt.Println("  -pkg        Package name (default: dest directory name)")
	fmt.Println("  -no-genstr  Don't add //go:generate comment")
	fmt.Println("  -help       Show this help")
}

type multiFlag []string

func (f *multiFlag) String() string {
	return strings.Join(*f, ", ")
}

func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func parseSrcSpec(s string) srcSpec {
	if s == "" {
		return srcSpec{path: ".", pattern: "*"}
	}

	parts := strings.SplitN(s, ":", 2)
	spec := srcSpec{}

	if len(parts) == 1 {
		if parts[0] == "" {
			spec.path = "."
			spec.pattern = "*"
		} else {
			spec.path = parts[0]
			spec.pattern = "*"
		}
	} else {
		if parts[0] == "" {
			spec.path = "."
		} else {
			spec.path = parts[0]
		}
		if parts[1] == "" {
			spec.pattern = "*"
		} else {
			spec.pattern = parts[1]
		}
	}

	return spec
}

func run(cfg genConfig) error {
	if err := os.MkdirAll(cfg.dest, 0755); err != nil {
		return fmt.Errorf("failed to create dest directory: %w", err)
	}

	allTypes := make([]typeInfo, 0)

	for _, src := range cfg.srcs {
		types, err := findTypes(src.path, src.pattern)
		if err != nil {
			return fmt.Errorf("failed to find types in %s: %w", src.path, err)
		}
		allTypes = append(allTypes, types...)
	}

	if len(allTypes) == 0 {
		fmt.Println("No matching types found")
		return nil
	}

	for _, t := range allTypes {
		if err := generateFile(cfg, t); err != nil {
			return fmt.Errorf("failed to generate %s: %w", t.Name, err)
		}
		fmt.Printf("Generated Table%s for %s.%s\n", t.Name, t.PkgName, t.Name)
	}

	return nil
}

func findTypes(dir, pattern string) ([]typeInfo, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	modulePath, err := findModulePath(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find module path: %w", err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, absDir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	types := make([]typeInfo, 0)

	for pkgName, pkg := range pkgs {
		pkgPath := modulePath
		if absDir != filepath.Dir(mustFindGoMod(absDir)) {
			relPath, err := filepath.Rel(filepath.Dir(mustFindGoMod(absDir)), absDir)
			if err != nil {
				return nil, err
			}
			if relPath != "." {
				pkgPath = modulePath + "/" + filepath.ToSlash(relPath)
			}
		}

		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					if _, ok := typeSpec.Type.(*ast.StructType); !ok {
						continue
					}

					typeName := typeSpec.Name.Name
					matched, err := filepath.Match(pattern, typeName)
					if err != nil {
						return nil, err
					}

					if matched {
						hasSQLName := checkSQLNameMethod(file, typeName)
						types = append(types, typeInfo{
							Name:      typeName,
							PkgPath:   pkgPath,
							PkgName:   pkgName,
							HasSQLNam: hasSQLName,
						})
					}
				}
			}
		}
	}

	return types, nil
}

func mustFindGoMod(dir string) string {
	path, err := findGoMod(dir)
	if err != nil {
		panic(err)
	}
	return path
}

func findGoMod(dir string) (string, error) {
	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return modPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func findModulePath(srcDir string) (string, error) {
	modPath, err := findGoMod(srcDir)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(modPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("module directive not found in go.mod")
}

func checkSQLNameMethod(file *ast.File, typeName string) bool {
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		recv := funcDecl.Recv.List[0]
		var recvType string

		switch t := recv.Type.(type) {
		case *ast.Ident:
			recvType = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recvType = ident.Name
			}
		}

		if recvType == typeName && funcDecl.Name.Name == "SQLName" {
			return true
		}
	}

	return false
}

func generateFile(cfg genConfig, t typeInfo) error {
	fileName := strings.ToLower("table_" + t.Name + ".go")
	filePath := filepath.Join(cfg.dest, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		GoGenerate string
		Package    string
		TypeInfo   typeInfo
	}{
		Package:  cfg.pkg,
		TypeInfo: t,
	}

	if !cfg.noGenStr {
		args := fmt.Sprintf("-src=%s:%s -dest=%s -pkg=%s", 
			filepath.Dir(mustFindGoMod(".")), "*", cfg.dest, cfg.pkg)
		data.GoGenerate = "//go:generate qcgen " + args
	}

	tmpl, err := template.New("table").Parse(tableTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

const tableTemplate = `{{if .GoGenerate}}{{.GoGenerate}}
{{end}}// Code generated by qcgen; DO NOT EDIT.

package {{.Package}}

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"quick-crud/contracts"
	"quick-crud/defs"
	"quick-crud/dialect"
	"quick-crud/filter"
	"quick-crud/struct_info"

	"github.com/mirrorru/dot"

	"{{.TypeInfo.PkgPath}}"
)

var _ contracts.TypedTable[{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}] = (*Table{{.TypeInfo.Name}})(nil)

type Table{{.TypeInfo.Name}} struct {
	dialect   dialect.SQLDialect
	tableInfo struct_info.TableInfo
	sqlTexts  struct_info.SqlTexts
}

type table{{.TypeInfo.Name}}Internals struct {
	TableInfo struct_info.TableInfo
	SqlTexts  struct_info.SqlTexts
}

func NewTable{{.TypeInfo.Name}}Val(d dialect.SQLDialect) Table{{.TypeInfo.Name}} {
	tableInfo := dot.MustMake(struct_info.GetTableInfo(reflect.TypeFor[{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}]()))

	return Table{{.TypeInfo.Name}}{
		dialect:   d,
		tableInfo: tableInfo,
		sqlTexts:  struct_info.SqlBuilderVal.SQLTexts(d, &tableInfo),
	}
}

func NewTable{{.TypeInfo.Name}}(d dialect.SQLDialect) *Table{{.TypeInfo.Name}} {
	return new(NewTable{{.TypeInfo.Name}}Val(d))
}

func (t *Table{{.TypeInfo.Name}}) Internals() table{{.TypeInfo.Name}}Internals {
	return table{{.TypeInfo.Name}}Internals{
		TableInfo: t.tableInfo,
		SqlTexts:  t.sqlTexts,
	}
}

func (t *Table{{.TypeInfo.Name}}) Ins(ctx context.Context, tx contracts.TxProcessor, row *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) (*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, contracts.Result, error) {
	args := t.tableInfo.Fields.ExtractArgs(row, t.tableInfo.InsertIdxList)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Insert, args)
		return row, sqlResult, err
	}
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Insert, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table{{.TypeInfo.Name}}) Upd(ctx context.Context, tx contracts.TxProcessor, row *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) (*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, contracts.Result, error) {
	args := t.tableInfo.Fields.ExtractArgs(row, t.tableInfo.UpdateIdxList)
	args = append(args,
		t.tableInfo.Fields.ExtractArgs(row, t.tableInfo.PKIdxList)...,
	)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Update, args)
		return row, sqlResult, err
	}
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Update, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table{{.TypeInfo.Name}}) One(ctx context.Context, tx contracts.TxProcessor, keys ...any) (*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, error) {
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)
	err := tx.QueryRowContext(ctx, t.sqlTexts.GetOne, keys...).Scan(refs...)

	return buf, err
}

func (t *Table{{.TypeInfo.Name}}) Del(ctx context.Context, tx contracts.TxProcessor, keys ...any) (contracts.Result, error) {
	return tx.ExecContext(ctx, t.sqlTexts.Delete, keys...)
}

func (t *Table{{.TypeInfo.Name}}) Many(ctx context.Context, tx contracts.TxProcessor, filter *filter.Filter) (result []*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, err error) {
	var (
		query strings.Builder
		args  []any
		where string
	)
	query.WriteString(t.sqlTexts.ListStart)

	if filter != nil {
		if filter.Range != nil {
			var argIdx int
			where, args, err = filter.Range.Build(t.tableInfo.Fields, t.dialect, &argIdx)
			if err != nil {
				return nil, err
			}
			query.WriteString(defs.SQLWhere)
			query.WriteString(where)
		}
	}
	query.WriteString(t.sqlTexts.SortPart)
	if filter != nil {
		query.WriteString(t.dialect.OffsetAndLimit(filter.Offset, filter.Limit))
	}
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)

	rows, err := tx.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	for rows.Next() {
		if err = rows.Scan(refs...); err != nil {
			return nil, err
		}
		rec := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
		*rec = *buf
		result = append(result, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
`
