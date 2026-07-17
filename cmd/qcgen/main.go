package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/mirrorru/crudquick/helpers"
	"github.com/mirrorru/crudquick/struct_info"
)

type srcSpec struct {
	path    string
	pattern string
}

type fieldTagFlags struct {
	IsPK        bool
	ReadOnly    bool
	AutoGen     bool
	Embed       bool
	ForceUpdate bool
	ForceInsert bool
	SkipReading bool
	ColName     string
	Prefix      string
	Ref         string
	Sort        string
}

func (f *fieldTagFlags) merge(parent fieldTagFlags) {
	f.IsPK = f.IsPK || parent.IsPK
	f.ReadOnly = f.ReadOnly || parent.ReadOnly
	f.AutoGen = f.AutoGen || parent.AutoGen
	f.ForceUpdate = f.ForceUpdate || parent.ForceUpdate
	f.ForceInsert = f.ForceInsert || parent.ForceInsert
	f.SkipReading = f.SkipReading || parent.SkipReading
	f.Prefix = parent.Prefix + f.Prefix
	if f.Sort == "" {
		f.Sort = parent.Sort
	}
}

func (f *fieldTagFlags) canInsert() bool {
	return f.ForceInsert || !f.ReadOnly && !f.AutoGen
}

func (f *fieldTagFlags) canUpdate() bool {
	return f.ForceUpdate || !f.ReadOnly && !f.IsPK
}

func (f *fieldTagFlags) canSelect() bool {
	return !f.SkipReading
}

func parseFieldTag(tag string) (result fieldTagFlags, ok bool) {

	keys := strings.Split(tag, struct_info.KeysSeparator)
	for _, key := range keys {
		switch {
		case key == struct_info.KeyPK:
			result.IsPK = true
		case key == struct_info.KeyRO:
			result.ReadOnly = true
		case key == struct_info.KeyAuto:
			result.AutoGen = true
		case key == struct_info.KeyEmbed:
			result.Embed = true
		case strings.HasPrefix(key, struct_info.KeyOmit):
			return result, false
		case key == struct_info.KeyInsert:
			result.ForceInsert = true
		case key == struct_info.KeyUpdate:
			result.ForceUpdate = true
		case key == struct_info.KeyHide:
			result.SkipReading = true
		case strings.HasPrefix(key, struct_info.KeyColName):
			result.ColName = key[len(struct_info.KeyColName):]
		case strings.HasPrefix(key, struct_info.KeyPrefix):
			result.Prefix = key[len(struct_info.KeyPrefix):]
		case strings.HasPrefix(key, struct_info.KeyRef):
			result.Ref = key[len(struct_info.KeyRef):]
		case strings.HasPrefix(key, struct_info.KeySort):
			result.Sort = key[len(struct_info.KeySort):]
		}
	}
	return result, true
}

type genField struct {
	Name         string
	AccessExpr   string
	Path         []string
	SQLName      string
	IsPK         bool
	CanSelect    bool
	CanInsert    bool
	CanUpdate    bool
	SortPos      int
	SortBackward bool
	RefTable     string
	RefField     string
}

type typeInfo struct {
	Name      string
	PkgPath   string
	PkgName   string
	HasSQLNam bool
	SQLName   string
	Fields    []genField
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
	if err := os.MkdirAll(cfg.dest, 0750); err != nil {
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

	type matchInfo struct {
		typeName string
		pkgName  string
		pkgPath  string
		file     *ast.File
		spec     *ast.TypeSpec
	}

	var matches []matchInfo
	allStructs := make(map[string]*ast.StructType)

	for pkgName, pkg := range pkgs {
		pkgPath := modulePath
		if absDir != filepath.Dir(mustFindGoMod(absDir)) {
			relPath, relErr := filepath.Rel(filepath.Dir(mustFindGoMod(absDir)), absDir)
			if relErr != nil {
				return nil, relErr
			}
			if relPath != "." {
				pkgPath = modulePath + "/" + filepath.ToSlash(relPath)
			}
		}

		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, isGenDecl := decl.(*ast.GenDecl)
				if !isGenDecl || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if structType, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
						allStructs[typeSpec.Name.Name] = structType
					}
				}
			}

			for _, decl := range file.Decls {
				genDecl, isGenDecl := decl.(*ast.GenDecl)
				if !isGenDecl || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, isStruct := typeSpec.Type.(*ast.StructType); !isStruct {
						continue
					}

					typeName := typeSpec.Name.Name
					matched, matchErr := filepath.Match(pattern, typeName)
					if matchErr != nil {
						return nil, matchErr
					}

					if matched {
						matches = append(matches, matchInfo{
							typeName: typeName,
							pkgName:  pkgName,
							pkgPath:  pkgPath,
							file:     file,
							spec:     typeSpec,
						})
					}
				}
			}
		}
	}

	result := make([]typeInfo, 0, len(matches))
	for _, m := range matches {
		structType, ok := m.spec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		fields := parseStructFields(structType, allStructs)
		sqlName, hasSQLName := extractSQLName(m.file, m.typeName)
		if !hasSQLName {
			sqlName = helpers.ToSnakeCase(m.typeName)
		}

		result = append(result, typeInfo{
			Name:      m.typeName,
			PkgPath:   m.pkgPath,
			PkgName:   m.pkgName,
			HasSQLNam: hasSQLName,
			SQLName:   sqlName,
			Fields:    fields,
		})
	}

	return result, nil
}

func parseStructFields(structType *ast.StructType, allStructs map[string]*ast.StructType) []genField {
	var result []genField
	for _, field := range structType.Fields.List {
		fields := collectFieldInfo(field, allStructs, fieldTagFlags{}, nil)
		result = append(result, fields...)
	}
	return result
}

func collectFieldInfo(field *ast.Field, allStructs map[string]*ast.StructType, parentFlags fieldTagFlags, parentPath []string) []genField {
	if len(field.Names) > 0 && !field.Names[0].IsExported() {
		return nil
	}

	var tagStr string
	if field.Tag != nil {
		tagStr = field.Tag.Value
		if len(tagStr) >= 2 && tagStr[0] == '`' && tagStr[len(tagStr)-1] == '`' {
			tagStr = tagStr[1 : len(tagStr)-1]
		}
		tagStr = extractTblTag(tagStr)
	}

	flags, processable := parseFieldTag(tagStr)
	if !processable {
		return nil
	}
	flags.merge(parentFlags)

	isEmbedded := false
	var embeddedTypeName string

	if len(field.Names) == 0 {
		isEmbedded = true
		embeddedTypeName = getTypeName(field.Type)
	} else if flags.Embed || flags.Prefix != "" {
		embeddedTypeName = getTypeName(field.Type)
		if embeddedTypeName != "" {
			isEmbedded = true
		}
	}

	if isEmbedded && embeddedTypeName != "" {
		if embeddedStruct, ok := allStructs[embeddedTypeName]; ok {
			var result []genField
			fieldName := embeddedTypeName
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}
			newPath := make([]string, len(parentPath)+1)
			copy(newPath, parentPath)
			newPath[len(parentPath)] = fieldName

			for _, subField := range embeddedStruct.Fields.List {
				subFields := collectFieldInfo(subField, allStructs, flags, newPath)
				result = append(result, subFields...)
			}
			return result
		}
	}

	if len(field.Names) == 0 {
		return nil
	}

	fieldName := field.Names[0].Name
	path := make([]string, len(parentPath)+1)
	copy(path, parentPath)
	path[len(parentPath)] = fieldName

	sqlName := flags.ColName
	if sqlName == "" {
		sqlName = helpers.ToSnakeCase(fieldName)
	}
	sqlName = flags.Prefix + sqlName

	sortParts := strings.SplitN(flags.Sort, ":", 2)
	sortPos := 0
	sortBackward := false
	if len(sortParts) > 0 && sortParts[0] != "" {
		sortPos, _ = strconv.Atoi(sortParts[0])
	}
	if sortPos > 1 && len(sortParts) > 1 {
		sortBackward = strings.ToLower(sortParts[1]) == "desc"
	}

	refTable := ""
	refField := ""
	refParts := strings.SplitN(flags.Ref, ",", 2)
	if len(refParts) == 2 {
		refTable, refField = refParts[0], refParts[1]
	}

	accessExpr := strings.Join(path, ".")

	return []genField{{
		Name:         fieldName,
		AccessExpr:   accessExpr,
		Path:         path,
		SQLName:      sqlName,
		IsPK:         flags.IsPK,
		CanSelect:    flags.canSelect(),
		CanInsert:    flags.canInsert(),
		CanUpdate:    flags.canUpdate(),
		SortPos:      sortPos,
		SortBackward: sortBackward,
		RefTable:     refTable,
		RefField:     refField,
	}}
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getTypeName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

func extractTblTag(tag string) string {
	st := reflect.StructTag(tag)
	value, _ := st.Lookup("tbl")
	return value
}

func extractSQLName(file *ast.File, typeName string) (string, bool) {
	for _, decl := range file.Decls {
		funcDecl, isFuncDecl := decl.(*ast.FuncDecl)
		if !isFuncDecl || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		recv := funcDecl.Recv.List[0]
		var recvType string

		switch t := recv.Type.(type) {
		case *ast.Ident:
			recvType = t.Name
		case *ast.StarExpr:
			if ident, isIdent := t.X.(*ast.Ident); isIdent {
				recvType = ident.Name
			}
		}

		if recvType == typeName && funcDecl.Name.Name == "SQLName" {
			if funcDecl.Body != nil {
				for _, stmt := range funcDecl.Body.List {
					if retStmt, isRetStmt := stmt.(*ast.ReturnStmt); isRetStmt {
						if len(retStmt.Results) > 0 {
							if lit, isLit := retStmt.Results[0].(*ast.BasicLit); isLit && lit.Kind == token.STRING {
								value := lit.Value
								if len(value) >= 2 {
									return value[1 : len(value)-1], true
								}
							}
						}
					}
				}
			}
			return "", true
		}
	}

	return "", false
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
			return "", errors.New("go.mod not found")
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

	return "", errors.New("module directive not found in go.mod")
}

func generateFile(cfg genConfig, t typeInfo) error {
	fileName := strings.ToLower("table_" + t.Name + ".go")
	filePath := filepath.Join(cfg.dest, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Pre-compute filtered lists for template
	var pkIndices, insertIndices, updateIndices, selectIndices, sortIndices, refIndices []int
	var selectableFields, insertableFields, updateableFields, pkFields []genField

	for idx, field := range t.Fields {
		if field.IsPK {
			pkIndices = append(pkIndices, idx)
			pkFields = append(pkFields, field)
		}
		if field.CanInsert {
			insertIndices = append(insertIndices, idx)
			insertableFields = append(insertableFields, field)
		}
		if field.CanUpdate {
			updateIndices = append(updateIndices, idx)
			updateableFields = append(updateableFields, field)
		}
		if field.CanSelect {
			selectIndices = append(selectIndices, idx)
			selectableFields = append(selectableFields, field)
		}
		if field.SortPos != 0 {
			sortIndices = append(sortIndices, idx)
		}
		if field.RefField != "" {
			refIndices = append(refIndices, idx)
		}
	}

	data := struct {
		GoGenerate       string
		Package          string
		TypeInfo         typeInfo
		PKIndices        []int
		InsertIndices    []int
		UpdateIndices    []int
		SelectIndices    []int
		SortIndices      []int
		RefIndices       []int
		SelectableFields []genField
		InsertableFields []genField
		UpdateableFields []genField
		PKFields         []genField
	}{
		GoGenerate:       "",
		Package:          cfg.pkg,
		TypeInfo:         t,
		PKIndices:        pkIndices,
		InsertIndices:    insertIndices,
		UpdateIndices:    updateIndices,
		SelectIndices:    selectIndices,
		SortIndices:      sortIndices,
		RefIndices:       refIndices,
		SelectableFields: selectableFields,
		InsertableFields: insertableFields,
		UpdateableFields: updateableFields,
		PKFields:         pkFields,
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
	"strings"

	"github.com/mirrorru/crudquick/contracts"
	"github.com/mirrorru/crudquick/defs"
	"github.com/mirrorru/crudquick/dialect"
	"github.com/mirrorru/crudquick/filter"
	"github.com/mirrorru/crudquick/struct_info"

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
	tableInfo := struct_info.TableInfo{
		SQLName: "{{.TypeInfo.SQLName}}",
		Fields: struct_info.TableFields{
{{- range $idx, $field := .TypeInfo.Fields}}
			{Path: []string{ {{- range $i, $p := $field.Path}}{{if $i}}, {{end}}"{{$p}}"{{end}} }, SQLName: "{{$field.SQLName}}"{{if $field.IsPK}}, IsPK: true{{end}}{{if $field.CanSelect}}, CanSelect: true{{end}}{{if $field.CanInsert}}, CanInsert: true{{end}}{{if $field.CanUpdate}}, CanUpdate: true{{end}}{{if ne $field.SortPos 0}}, SortPos: {{$field.SortPos}}{{end}}{{if $field.SortBackward}}, SortBackward: true{{end}}{{if $field.RefTable}}, RefTable: "{{$field.RefTable}}", RefField: "{{$field.RefField}}"{{end}}},
{{- end}}
		},
		FieldNameIdx:  map[string]int{ {{- range $idx, $field := .TypeInfo.Fields}}{{if $idx}}, {{end}}"{{$field.SQLName}}": {{$idx}}{{end}} },
		PKIdxList:     []int{ {{- range $idx, $v := .PKIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		InsertIdxList: []int{ {{- range $idx, $v := .InsertIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		UpdateIdxList: []int{ {{- range $idx, $v := .UpdateIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		SelectIdxList: []int{ {{- range $idx, $v := .SelectIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		SortIdxList:   []int{ {{- range $idx, $v := .SortIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		RefIdxList:    []int{ {{- range $idx, $v := .RefIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
	}

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

func (t *Table{{.TypeInfo.Name}}) scanRefs(buf *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) []any {
	return []any{ {{- range $idx, $field := .SelectableFields}}{{if $idx}}, {{end}}&buf.{{$field.AccessExpr}}{{end}} }
}

func (t *Table{{.TypeInfo.Name}}) insertArgs(row *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) []any {
	return []any{ {{- range $idx, $field := .InsertableFields}}{{if $idx}}, {{end}}row.{{$field.AccessExpr}}{{end}} }
}

func (t *Table{{.TypeInfo.Name}}) updateArgs(row *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) []any {
	return []any{ {{- range $idx, $field := .UpdateableFields}}{{if $idx}}, {{end}}row.{{$field.AccessExpr}}{{end}}{{range $idx, $field := .PKFields}}, row.{{$field.AccessExpr}}{{end}} }
}

func (t *Table{{.TypeInfo.Name}}) Ins(ctx context.Context, tx contracts.TxProcessor, row *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) (*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, contracts.Result, error) {
	args := t.insertArgs(row)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Insert, args)
		return row, sqlResult, err
	}
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.scanRefs(buf)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Insert, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table{{.TypeInfo.Name}}) Upd(ctx context.Context, tx contracts.TxProcessor, row *{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}) (*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, contracts.Result, error) {
	args := t.updateArgs(row)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Update, args)
		return row, sqlResult, err
	}
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.scanRefs(buf)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Update, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table{{.TypeInfo.Name}}) One(ctx context.Context, tx contracts.TxProcessor, struct_info.Keys ...any) (*{{.TypeInfo.PkgName}}.{{.TypeInfo.Name}}, error) {
	buf := new({{.TypeInfo.PkgName}}.{{.TypeInfo.Name}})
	refs := t.scanRefs(buf)
	err := tx.QueryRowContext(ctx, t.sqlTexts.GetOne, struct_info.Keys...).Scan(refs...)

	return buf, err
}

func (t *Table{{.TypeInfo.Name}}) Del(ctx context.Context, tx contracts.TxProcessor, struct_info.Keys ...any) (contracts.Result, error) {
	return tx.ExecContext(ctx, t.sqlTexts.Delete, struct_info.Keys...)
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
	refs := t.scanRefs(buf)

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
