//nolint:all
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
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/helpers"
	"github.com/mirrorru/cruds/struct_info"
)

type srcSpec struct {
	path    string
	pattern string
}

type genField struct {
	Name         string
	AccessExpr   string
	Path         []string
	SQLName      string
	GoType       string
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
	srcs      []srcSpec
	dest      string
	pkg       string
	noGenStr  bool
	builds    []string
	genTable  bool
	genJoiner bool
}

type pkgTypeEntry struct {
	pkgName    string
	pkgPath    string
	typeName   string
	structType *ast.StructType
	file       *ast.File
}

type joinFieldInfo struct {
	Name          string
	AccessExpr    string
	TypeName      string
	PkgName       string
	SubPkgName    string
	SubPkgPath    string
	IsPointer     bool
	IsFrom        bool
	IsPK          bool
	SortPriority  int
	JoinMode      string
	Alias         string
	RefAliasMap   map[string]string
	Fields        []genField
	SQLName       string
	HasSQLName    bool
	ComputedAlias string
	Index         int
}

type joinTypeInfo struct {
	Name       string
	PkgPath    string
	PkgName    string
	Fields     []joinFieldInfo
	PkgImports map[string]string
}

func main() {
	var srcFlags multiFlag
	var buildFlags multiFlag
	var dest string
	var pkg string
	var noGenStr bool
	var tableFlag bool
	var joinerFlag bool
	var help bool

	flag.Var(&srcFlags, "src", "Source specification (path:pattern), can be repeated")
	flag.Var(&buildFlags, "build", "Build tag to add to generated file (can be repeated, joined with ||)")
	flag.StringVar(&dest, "dest", "", "Destination directory (required)")
	flag.StringVar(&pkg, "pkg", "", "Package name (default: dest directory name)")
	flag.BoolVar(&noGenStr, "no-genstr", false, "Don't add //go:generate comment")
	flag.BoolVar(&tableFlag, "table", false, "Generate Table* typed tables (default true if neither -table nor -joiner specified)")
	flag.BoolVar(&joinerFlag, "joiner", false, "Generate Joiner* typed joiners")
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

	hasTable := false
	hasJoiner := false
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "table":
			hasTable = true
		case "joiner":
			hasJoiner = true
		}
	})

	if !hasTable && !hasJoiner {
		tableFlag = true
	}

	cfg := genConfig{
		srcs:      srcs,
		dest:      dest,
		pkg:       pkg,
		noGenStr:  noGenStr,
		builds:    []string(buildFlags),
		genTable:  tableFlag,
		genJoiner: joinerFlag,
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("crudsgen - Code generator for typed table and joiner implementations")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  crudsgen -src=<path:pattern> [-src=...] -dest=<dir> [-pkg=<name>] [-no-genstr] [-build=<build-tag>] [-table] [-joiner]")
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
	fmt.Println("  -build      Build tag to add to generated file (can be repeated, joined with ||)")
	fmt.Println("  -table      Generate Table* typed tables (default if neither -table nor -joiner specified)")
	fmt.Println("  -joiner     Generate Joiner* typed joiners")
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
	allJoinTypes := make([]joinTypeInfo, 0)

	var globalRegistry map[string]*pkgTypeEntry
	var globalPkgImports map[string]string
	if cfg.genJoiner {
		var err error
		globalRegistry, globalPkgImports, err = buildGlobalRegistry(cfg.srcs)
		if err != nil {
			return fmt.Errorf("failed to build global type registry: %w", err)
		}
	}

	for _, src := range cfg.srcs {
		if cfg.genTable {
			types, err := findTypes(src.path, src.pattern)
			if err != nil {
				return fmt.Errorf("failed to find types in %s: %w", src.path, err)
			}
			allTypes = append(allTypes, types...)
		}

		if cfg.genJoiner {
			jtypes, err := findJoinTypes(src.path, src.pattern, globalRegistry)
			if err != nil {
				return fmt.Errorf("failed to find join types in %s: %w", src.path, err)
			}
			for i := range jtypes {
				jtypes[i].PkgImports = globalPkgImports
			}
			allJoinTypes = append(allJoinTypes, jtypes...)
		}
	}

	if len(allTypes) == 0 && len(allJoinTypes) == 0 {
		fmt.Println("No matching types found")
		return nil
	}

	for _, t := range allTypes {
		if err := generateFile(cfg, t); err != nil {
			return fmt.Errorf("failed to generate %s: %w", t.Name, err)
		}
		fmt.Printf("Generated Table%s for %s.%s\n", t.Name, t.PkgName, t.Name)
	}

	for _, jt := range allJoinTypes {
		if err := generateJoinerFile(cfg, jt); err != nil {
			return fmt.Errorf("failed to generate joiner %s: %w", jt.Name, err)
		}
		fmt.Printf("Generated Joiner%s for %s.%s\n", jt.Name, jt.PkgName, jt.Name)
	}

	return nil
}

func buildGlobalRegistry(specs []srcSpec) (map[string]*pkgTypeEntry, map[string]string, error) {
	registry := make(map[string]*pkgTypeEntry)
	pkgImports := make(map[string]string)

	processed := make(map[string]bool)

	for _, spec := range specs {
		absDir, err := filepath.Abs(spec.path)
		if err != nil {
			return nil, nil, err
		}
		processed[absDir] = true
	}

	modRoot := filepath.Dir(mustFindGoMod("."))
	modName := findModuleName(modRoot)
	if modName == "" {
		return nil, nil, errors.New("failed to find module name")
	}

	type dirToProcess struct {
		dir  string
		from string
	}

	dirs := make([]dirToProcess, 0)
	for _, spec := range specs {
		absDir, _ := filepath.Abs(spec.path)
		dirs = append(dirs, dirToProcess{dir: absDir, from: "src"})
	}

	for len(dirs) > 0 {
		dp := dirs[0]
		dirs = dirs[1:]

		modulePath, err := findModulePath(dp.dir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find module path for %s: %w", dp.dir, err)
		}

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dp.dir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}

		for pkgName, pkg := range pkgs {
			pkgPath := modulePath
			modDir := filepath.Dir(mustFindGoMod(dp.dir))
			if dp.dir != modDir {
				relPath, relErr := filepath.Rel(modDir, dp.dir)
				if relErr != nil {
					return nil, nil, relErr
				}
				if relPath != "." {
					pkgPath = modulePath + "/" + filepath.ToSlash(relPath)
				}
			}

			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					var importPkgName string
					if imp.Name != nil {
						importPkgName = imp.Name.Name
					} else {
						idx := strings.LastIndex(importPath, "/")
						if idx >= 0 {
							importPkgName = importPath[idx+1:]
						} else {
							importPkgName = importPath
						}
					}
					if importPkgName == "_" || importPkgName == "." || importPkgName == "" {
						continue
					}
					if _, exists := pkgImports[importPkgName]; !exists {
						pkgImports[importPkgName] = importPath
					}

					if strings.HasPrefix(importPath, modName+"/") || importPath == modName {
						relPath := strings.TrimPrefix(importPath, modName)
						relPath = strings.TrimPrefix(relPath, "/")
						depDir := filepath.Join(modRoot, filepath.FromSlash(relPath))
						if !processed[depDir] {
							processed[depDir] = true
							dirs = append(dirs, dirToProcess{dir: depDir, from: "import"})
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
						structType, isStruct := typeSpec.Type.(*ast.StructType)
						if !isStruct {
							continue
						}
						key := registryKey(pkgName, typeSpec.Name.Name)
						if _, exists := registry[key]; !exists {
							registry[key] = &pkgTypeEntry{
								pkgName:    pkgName,
								pkgPath:    pkgPath,
								typeName:   typeSpec.Name.Name,
								structType: structType,
								file:       file,
							}
						}
					}
				}
			}
		}
	}

	return registry, pkgImports, nil
}

func findModuleName(modRoot string) string {
	content, err := os.ReadFile(filepath.Join(modRoot, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
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

		localReg := make(map[string]*pkgTypeEntry, len(allStructs))
		for name, st := range allStructs {
			localReg[registryKey(m.pkgName, name)] = &pkgTypeEntry{
				pkgName:    m.pkgName,
				pkgPath:    m.pkgPath,
				typeName:   name,
				structType: st,
				file:       m.file,
			}
		}

		fields := parseStructFields(structType, localReg, m.pkgName)
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

func findJoinTypes(dir, pattern string, registry map[string]*pkgTypeEntry) ([]joinTypeInfo, error) {
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

	result := make([]joinTypeInfo, 0, len(matches))
	for _, m := range matches {
		structType, ok := m.spec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		jFields, isJoin := parseJoinStructFields(structType, registry, m.pkgName)
		if !isJoin || len(jFields) == 0 {
			continue
		}

		result = append(result, joinTypeInfo{
			Name:    m.typeName,
			PkgPath: m.pkgPath,
			PkgName: m.pkgName,
			Fields:  jFields,
		})
	}

	return result, nil
}

func parseJoinStructFields(structType *ast.StructType, registry map[string]*pkgTypeEntry, localPkgName string) ([]joinFieldInfo, bool) {
	var result []joinFieldInfo
	hasJoinFields := false

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			if entry, ok := registry[registryKey(localPkgName, getTypeName(field.Type))]; ok {
				subFields, subIsJoin := parseJoinStructFields(entry.structType, registry, entry.pkgName)
				if subIsJoin {
					result = append(result, subFields...)
					hasJoinFields = true
				}
			}
			continue
		}

		if !field.Names[0].IsExported() {
			continue
		}

		var tagStr string
		if field.Tag != nil {
			tagStr = field.Tag.Value
			if len(tagStr) >= 2 && tagStr[0] == '`' && tagStr[len(tagStr)-1] == '`' {
				tagStr = tagStr[1 : len(tagStr)-1]
			}
			tagStr = extractTblTag(tagStr)
		}

		if tagStr == "" {
			continue
		}

		joinFlags, processable := cruds.ParseJoinTableFlags(tagStr)
		if !processable {
			continue
		}

		isPtr := false
		if _, isStar := field.Type.(*ast.StarExpr); isStar {
			isPtr = true
		}

		typeName, subPkgName := getQualifiedTypeName(field.Type, localPkgName)
		entry, ok := registry[registryKey(subPkgName, typeName)]
		if !ok {
			continue
		}

		hasJoinFields = true

		subFields := parseStructFields(entry.structType, registry, entry.pkgName)
		sqlName, hasSQLName := findSQLName(entry, registry)
		if !hasSQLName {
			sqlName = helpers.ToSnakeCase(typeName)
		}

		sortPriority := 0
		if joinFlags.Sort != "" {
			sortPriority, _ = strconv.Atoi(joinFlags.Sort)
		}

		refAliasMap := make(map[string]string)
		if joinFlags.Map != "" {
			for _, m := range strings.Split(joinFlags.Map, ",") {
				kv := strings.SplitN(m, ":", 2)
				if len(kv) == 2 {
					refAliasMap[kv[0]] = kv[1]
				}
			}
		}

		fieldName := field.Names[0].Name
		result = append(result, joinFieldInfo{
			Name:         fieldName,
			AccessExpr:   fieldName,
			TypeName:     typeName,
			PkgName:      subPkgName,
			SubPkgName:   subPkgName,
			SubPkgPath:   entry.pkgPath,
			IsPointer:    isPtr,
			IsFrom:       joinFlags.IsFrom,
			IsPK:         joinFlags.IsPK,
			SortPriority: sortPriority,
			JoinMode:     joinFlags.Join,
			Alias:        joinFlags.Alias,
			RefAliasMap:  refAliasMap,
			Fields:       subFields,
			SQLName:      sqlName,
			HasSQLName:   hasSQLName,
		})
	}

	if !hasJoinFields {
		return nil, false
	}

	return result, true
}

func findSQLName(entry *pkgTypeEntry, registry map[string]*pkgTypeEntry) (string, bool) {
	if name, ok := extractSQLName(entry.file, entry.typeName); ok {
		return name, true
	}

	if name, ok := extractSQLNameEmbd(entry.file, entry.structType, registry, entry.pkgName); ok {
		return name, true
	}

	return "", false
}

func extractSQLNameEmbd(file *ast.File, structType *ast.StructType, registry map[string]*pkgTypeEntry, localPkgName string) (string, bool) {
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			typeName, pkgName := getQualifiedTypeName(field.Type, localPkgName)
			if entry, ok := registry[registryKey(pkgName, typeName)]; ok {
				if name, ok2 := extractSQLName(file, typeName); ok2 {
					return name, true
				}
				if name, ok2 := extractSQLNameEmbd(entry.file, entry.structType, registry, entry.pkgName); ok2 {
					return name, true
				}
			}
			continue
		}
	}
	return "", false
}

func parseStructFields(structType *ast.StructType, registry map[string]*pkgTypeEntry, pkgName string) []genField {
	var result []genField
	for _, field := range structType.Fields.List {
		fields := collectFieldInfo(field, registry, struct_info.FieldTagFlags{}, nil, pkgName)
		result = append(result, fields...)
	}
	return result
}

func collectFieldInfo(field *ast.Field, registry map[string]*pkgTypeEntry, parentFlags struct_info.FieldTagFlags, parentPath []string, pkgName string) []genField {
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

	flags, processable := struct_info.ParseFieldTag(tagStr)
	if !processable {
		return nil
	}
	flags.Merge(parentFlags)

	isEmbedded := false
	var embeddedTypeName string

	if len(field.Names) == 0 {
		isEmbedded = true
		embeddedTypeName, _ = getQualifiedTypeName(field.Type, pkgName)
	} else if flags.Embed || flags.Prefix != "" {
		embeddedTypeName, _ = getQualifiedTypeName(field.Type, pkgName)
		if embeddedTypeName != "" {
			isEmbedded = true
		}
	}

	if isEmbedded && embeddedTypeName != "" {
		eTypeName, ePkgName := getQualifiedTypeName(field.Type, pkgName)
		if entry, ok := registry[registryKey(ePkgName, eTypeName)]; ok {
			embeddedStruct := entry.structType
			embeddedPkgName := entry.pkgName
			var result []genField
			fieldName := embeddedTypeName
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}
			newPath := make([]string, len(parentPath)+1)
			copy(newPath, parentPath)
			newPath[len(parentPath)] = fieldName

			for _, subField := range embeddedStruct.Fields.List {
				subFields := collectFieldInfo(subField, registry, flags, newPath, embeddedPkgName)
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
	if sortPos > 0 && len(sortParts) > 1 {
		sortBackward = strings.ToLower(sortParts[1]) == "desc"
	}

	refTable := ""
	refField := ""
	refParts := strings.SplitN(flags.Ref, ":", 2)
	if len(refParts) == 2 {
		refTable, refField = refParts[0], refParts[1]
	}

	accessExpr := strings.Join(path, ".")

	goType := astTypeToGoType(field.Type, pkgName)

	return []genField{{
		Name:         fieldName,
		AccessExpr:   accessExpr,
		Path:         path,
		SQLName:      sqlName,
		GoType:       goType,
		IsPK:         flags.IsPK,
		CanSelect:    flags.CanSelect(),
		CanInsert:    flags.CanInsert(),
		CanUpdate:    flags.CanUpdate(),
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

func identName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func getQualifiedTypeName(expr ast.Expr, localPkgName string) (typeName string, pkgName string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, localPkgName
	case *ast.StarExpr:
		return getQualifiedTypeName(t.X, localPkgName)
	case *ast.SelectorExpr:
		return t.Sel.Name, identName(t.X)
	}
	return "", ""
}

func registryKey(pkgName, typeName string) string {
	return pkgName + "." + typeName
}

var builtinTypes = map[string]bool{
	"bool":       true,
	"byte":       true,
	"complex64":  true,
	"complex128": true,
	"error":      true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"rune":       true,
	"string":     true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"any":        true,
}

func astTypeToGoType(expr ast.Expr, pkgName string) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if !builtinTypes[t.Name] && pkgName != "" {
			return pkgName + "." + t.Name
		}
		return t.Name
	case *ast.StarExpr:
		return "*" + astTypeToGoType(t.X, pkgName)
	case *ast.SelectorExpr:
		return exprToGoString(t)
	case *ast.ArrayType:
		return "[]" + astTypeToGoType(t.Elt, pkgName)
	case *ast.MapType:
		return "map[" + astTypeToGoType(t.Key, pkgName) + "]" + astTypeToGoType(t.Value, pkgName)
	default:
		return "any"
	}
}

func exprToGoString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToGoString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToGoString(t.X)
	default:
		return "any"
	}
}

func extractTblTag(tag string) string {
	st := reflect.StructTag(tag)
	value, _ := st.Lookup(struct_info.TagName)
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

func buildTagLine(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "//go:build " + strings.Join(tags, " || ")
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

	needSrcPkgTable := t.PkgPath != moduleOfDest(cfg)
	qualTableType := t.PkgName + "." + t.Name
	if !needSrcPkgTable {
		qualTableType = t.Name
	}

	data := struct {
		BuildTag         string
		GoGenerate       string
		Package          string
		TypeInfo         typeInfo
		QualTypeName     string
		NeedSrcPkg       bool
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
		BuildTag:         buildTagLine(cfg.builds),
		GoGenerate:       "",
		Package:          cfg.pkg,
		TypeInfo:         t,
		QualTypeName:     qualTableType,
		NeedSrcPkg:       needSrcPkgTable,
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
		data.GoGenerate = "//go:generate crudsgen " + args
	}

	tmpl, err := template.New("table").Parse(tableTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

const tableTemplate = `{{if .BuildTag}}{{.BuildTag}}
{{end}}{{if .GoGenerate}}{{.GoGenerate}}
{{end}}// Code generated by crudsgen; DO NOT EDIT.

package {{.Package}}

import (
	"context"
	"errors"
	"strings"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/defs"
	"github.com/mirrorru/cruds/dialect"
	"github.com/mirrorru/cruds/struct_info"
{{if .NeedSrcPkg}}
	"{{.TypeInfo.PkgPath}}"
{{end}}
)

var _ cruds.TypedTable[{{.QualTypeName}}] = (*Table{{.TypeInfo.Name}})(nil)

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

func (t *Table{{.TypeInfo.Name}}) scanRefs(buf *{{.QualTypeName}}) []any {
	return []any{ {{- range $idx, $field := .SelectableFields}}{{if $idx}}, {{end}}&buf.{{$field.AccessExpr}}{{end}}}
}

func (t *Table{{.TypeInfo.Name}}) insertArgs(row *{{.QualTypeName}}) []any {
	return []any{ {{- range $idx, $field := .InsertableFields}}{{if $idx}}, {{end}}row.{{$field.AccessExpr}}{{end}}}
}

func (t *Table{{.TypeInfo.Name}}) updateArgs(row *{{.QualTypeName}}) []any {
	return []any{ {{- range $idx, $field := .UpdateableFields}}{{if $idx}}, {{end}}row.{{$field.AccessExpr}}{{end}}{{range $idx, $field := .PKFields}}, row.{{$field.AccessExpr}}{{end}}}
}

func (t *Table{{.TypeInfo.Name}}) Ins(ctx context.Context, tx cruds.TxProcessor, row *{{.QualTypeName}}) (*{{.QualTypeName}}, cruds.Result, error) {
	args := t.insertArgs(row)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Insert, args)
		return row, sqlResult, err
	}
	buf := new({{.QualTypeName}})
	refs := t.scanRefs(buf)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Insert, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table{{.TypeInfo.Name}}) Upd(ctx context.Context, tx cruds.TxProcessor, row *{{.QualTypeName}}) (*{{.QualTypeName}}, cruds.Result, error) {
	args := t.updateArgs(row)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Update, args)
		return row, sqlResult, err
	}
	buf := new({{.QualTypeName}})
	refs := t.scanRefs(buf)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Update, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table{{.TypeInfo.Name}}) One(ctx context.Context, tx cruds.TxProcessor, keys ...any) (*{{.QualTypeName}}, error) {
	buf := new({{.QualTypeName}})
	refs := t.scanRefs(buf)
	err := tx.QueryRowContext(ctx, t.sqlTexts.GetOne, keys...).Scan(refs...)

	return buf, err
}

func (t *Table{{.TypeInfo.Name}}) Del(ctx context.Context, tx cruds.TxProcessor, keys ...any) (cruds.Result, error) {
	return tx.ExecContext(ctx, t.sqlTexts.Delete, keys...)
}

func (t *Table{{.TypeInfo.Name}}) Many(ctx context.Context, tx cruds.TxProcessor, filter *cruds.Filter) (result []*{{.QualTypeName}}, err error) {
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
	buf := new({{.QualTypeName}})
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
		rec := new({{.QualTypeName}})
		*rec = *buf
		result = append(result, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
`

type subTableTmpl struct {
	joinFieldInfo
	SelectableFields []genField
	SelectIndices    []int
	PKFields         []genField
	PKIndices        []int
	SortFields       []genField
	SortIndices      []int
	RefFields        []genField
	RefIndices       []int
	QualSubTypeName  string
}

func (st subTableTmpl) SelectIdxOfField(f genField) int {
	for i, sf := range st.SelectableFields {
		if sf.AccessExpr == f.AccessExpr {
			return i
		}
	}
	return -1
}

type allFieldTmpl struct {
	Alias     string
	FieldName string
}

type extraImportTmpl struct {
	PkgName string
	PkgPath string
}

type pointerVarTmpl struct {
	VarName string
	Fields  []genField
}

func generateJoinerFile(cfg genConfig, jt joinTypeInfo) error {
	fileName := strings.ToLower("joiner_" + jt.Name + ".go")
	filePath := filepath.Join(cfg.dest, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	aliasCnt := 0
	aliasUsed := map[string]bool{}
	for i := range jt.Fields {
		jt.Fields[i].Index = i
		if jt.Fields[i].Alias != "" {
			aliasUsed[jt.Fields[i].Alias] = true
		}
	}
	for i := range jt.Fields {
		if jt.Fields[i].Alias == "" {
			for {
				aliasCnt++
				jt.Fields[i].ComputedAlias = fmt.Sprintf("T%d", aliasCnt)
				if !aliasUsed[jt.Fields[i].ComputedAlias] {
					break
				}
			}
		} else {
			jt.Fields[i].ComputedAlias = jt.Fields[i].Alias
		}
	}

	var subTables []subTableTmpl
	needSrcPkg := jt.PkgPath != moduleOfDest(cfg)
	destMod := moduleOfDest(cfg)

	extraImports := map[string]string{}
	for _, jf := range jt.Fields {
		if jf.SubPkgPath != "" && jf.SubPkgPath != destMod {
			extraImports[jf.SubPkgPath] = jf.SubPkgName
		}
		for _, f := range jf.Fields {
			extractPkgRefs(f.GoType, jt.PkgImports, extraImports)
		}
	}

	for _, jf := range jt.Fields {
		st := subTableTmpl{joinFieldInfo: jf}
		for fi, f := range jf.Fields {
			if f.CanSelect {
				st.SelectableFields = append(st.SelectableFields, f)
				st.SelectIndices = append(st.SelectIndices, fi)
			}
			if f.IsPK {
				st.PKFields = append(st.PKFields, f)
				st.PKIndices = append(st.PKIndices, fi)
			}
			if f.SortPos != 0 {
				st.SortFields = append(st.SortFields, f)
				st.SortIndices = append(st.SortIndices, fi)
			}
			if f.RefField != "" {
				st.RefFields = append(st.RefFields, f)
				st.RefIndices = append(st.RefIndices, fi)
			}
		}

		subNeedPkg := jf.SubPkgPath != destMod
		if subNeedPkg {
			st.QualSubTypeName = jf.SubPkgName + "." + st.TypeName
		} else {
			st.QualSubTypeName = st.TypeName
			pkgPrefix := jf.SubPkgName + "."
			for i := range st.Fields {
				st.Fields[i].GoType = strings.TrimPrefix(st.Fields[i].GoType, pkgPrefix)
			}
			for i := range st.SelectableFields {
				st.SelectableFields[i].GoType = strings.TrimPrefix(st.SelectableFields[i].GoType, pkgPrefix)
			}
		}
		subTables = append(subTables, st)
	}

	hasFrom := false
	for i := range subTables {
		if subTables[i].IsFrom {
			hasFrom = true
			break
		}
	}
	if !hasFrom && len(subTables) > 0 {
		subTables[0].IsFrom = true
	}

	var allFields []allFieldTmpl
	for _, st := range subTables {
		for _, f := range st.SelectableFields {
			allFields = append(allFields, allFieldTmpl{
				Alias:     st.ComputedAlias,
				FieldName: f.SQLName,
			})
		}
	}

	var pointerVars []pointerVarTmpl
	for _, st := range subTables {
		if !st.IsPointer {
			continue
		}
		pointerVars = append(pointerVars, pointerVarTmpl{
			VarName: "_pr" + st.ComputedAlias,
			Fields:  st.SelectableFields,
		})
	}

	goGenStr := ""
	if !cfg.noGenStr {
		args := fmt.Sprintf("-src=%s:%s -dest=%s -pkg=%s -joiner",
			filepath.Dir(mustFindGoMod(".")), "*", cfg.dest, cfg.pkg)
		goGenStr = "//go:generate crudsgen " + args
	}

	qualType := jt.PkgName + "." + jt.Name
	if !needSrcPkg {
		qualType = jt.Name
	}

	data := struct {
		BuildTag     string
		GoGenerate   string
		Package      string
		TypeInfo     joinTypeInfo
		SubTables    []subTableTmpl
		AllFields    []allFieldTmpl
		PointerVars  []pointerVarTmpl
		NeedSrcPkg   bool
		QualTypeName string
		ExtraImports []extraImportTmpl
	}{
		BuildTag:     buildTagLine(cfg.builds),
		GoGenerate:   goGenStr,
		Package:      cfg.pkg,
		TypeInfo:     jt,
		SubTables:    subTables,
		AllFields:    allFields,
		PointerVars:  pointerVars,
		NeedSrcPkg:   needSrcPkg,
		QualTypeName: qualType,
		ExtraImports: make([]extraImportTmpl, 0),
	}

	for pkgPath, pkgName := range extraImports {
		data.ExtraImports = append(data.ExtraImports, extraImportTmpl{PkgName: pkgName, PkgPath: pkgPath})
	}

	tmpl, err := template.New("joiner").Funcs(template.FuncMap{
		"selIdxOf": func(fields []genField, f genField) int {
			for i, sf := range fields {
				if sf.AccessExpr == f.AccessExpr {
					return i
				}
			}
			return -1
		},
		"joinModeName": func(mode string) string {
			switch strings.ToLower(mode) {
			case "left":
				return "cruds.LeftJoin"
			case "right":
				return "cruds.RightJoin"
			case "outer":
				return "cruds.OuterJoin"
			case "cross":
				return "cruds.CrossJoin"
			case "inner", "":
				return "cruds.InnerJoin"
			default:
				return "cruds.InnerJoin"
			}
		},
	}).Parse(joinerTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

func extractPkgRefs(goType string, pkgImports map[string]string, extraImports map[string]string) {
	parts := strings.Split(strings.TrimPrefix(goType, "*"), ".")
	if len(parts) < 2 {
		return
	}
	pkgName := parts[0]
	if pkgImports == nil {
		return
	}
	if pkgPath, ok := pkgImports[pkgName]; ok {
		if _, already := extraImports[pkgPath]; !already {
			extraImports[pkgPath] = pkgName
		}
	}
}

func moduleOfDest(cfg genConfig) string {
	absDest, err := filepath.Abs(cfg.dest)
	if err != nil {
		return ""
	}
	modPath := filepath.Dir(mustFindGoMod(absDest))
	modRel, err := filepath.Rel(modPath, absDest)
	if err != nil {
		return ""
	}
	modContent, err := os.ReadFile(filepath.Join(modPath, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(modContent), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			if modRel == "." {
				return modName
			}
			return modName + "/" + filepath.ToSlash(modRel)
		}
	}
	return ""
}

func sortOrderIdx(subTables []subTableTmpl) []int {
	type idxSort struct {
		idx      int
		priority int
	}
	var items []idxSort
	for i, st := range subTables {
		if st.SortPriority != 0 {
			items = append(items, idxSort{idx: i, priority: st.SortPriority})
		}
	}
	sort.Slice(items, func(a, b int) bool {
		return items[a].priority < items[b].priority
	})
	result := make([]int, len(items))
	for i, item := range items {
		result[i] = item.idx
	}
	return result
}

const joinerTemplate = `{{if .BuildTag}}{{.BuildTag}}
{{end}}{{if .GoGenerate}}{{.GoGenerate}}
{{end}}// Code generated by crudsgen; DO NOT EDIT.
//nolint:lll
package {{.Package}}

import (
	"context"
	"errors"
	"strings"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/defs"
	"github.com/mirrorru/cruds/dialect"
	"github.com/mirrorru/cruds/struct_info"
{{if .NeedSrcPkg}}
	"{{.TypeInfo.PkgPath}}"
{{end}}{{range .ExtraImports}}
	{{.PkgName}} "{{.PkgPath}}"
{{end}}
)

var _ cruds.TypedJoiner[{{.QualTypeName}}] = (*Joiner{{.TypeInfo.Name}})(nil)

type Joiner{{.TypeInfo.Name}} struct {
	dialect    dialect.SQLDialect
	allFields  struct_info.TableFields
	getOneSQL  string
	getManySQL string
	sortSQL    string

}

func NewJoiner{{.TypeInfo.Name}}Val(d dialect.SQLDialect) Joiner{{.TypeInfo.Name}} {
{{- range .SubTables}}
	ti{{.Index}} := &struct_info.TableInfo{
		SQLName: "{{.SQLName}}",
		Fields: struct_info.TableFields{
{{- range .Fields}}
			{Path: []string{ {{- range $pi, $p := .Path}}{{if $pi}}, {{end}}"{{$p}}"{{end}} }, SQLName: "{{.SQLName}}"{{if .IsPK}}, IsPK: true{{end}}{{if .CanSelect}}, CanSelect: true{{end}}{{if .CanInsert}}, CanInsert: true{{end}}{{if .CanUpdate}}, CanUpdate: true{{end}}{{if ne .SortPos 0}}, SortPos: {{.SortPos}}{{end}}{{if .SortBackward}}, SortBackward: true{{end}}{{		if .RefTable}}, RefTable: "{{.RefTable}}", RefField: "{{.RefField}}"{{end}}},
{{- end}}
		},
		FieldNameIdx:  map[string]int{ {{- range $idx, $f := .Fields}}{{if $idx}}, {{end}}"{{$f.SQLName}}": {{$idx}}{{end}} },
		PKIdxList:     []int{ {{- range $idx, $v := .PKIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		InsertIdxList: make([]int, 0),
		UpdateIdxList: make([]int, 0),
		SelectIdxList: []int{ {{- range $idx, $v := .SelectIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		SortIdxList:   []int{ {{- range $idx, $v := .SortIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} },
		RefIdxList:    {{if .RefIndices}}[]int{ {{- range $idx, $v := .RefIndices}}{{if $idx}}, {{end}}{{$v}}{{end}} }{{else}}make([]int, 0){{end}},
	}
{{end}}
	joinTables := cruds.JoinTables{
{{- range .SubTables}}
		{
			TableInfo:    ti{{.Index}},
			IsPointer:    {{.IsPointer}},
			IsFrom:       {{.IsFrom}},
			IsPK:         {{.IsPK}},
			SortPriority: {{.SortPriority}},
			JoinModeVal:  {{joinModeName .JoinMode}},
			Alias:        "{{.Alias}}",
			RefAliasMap:  {{if .RefAliasMap}}map[string]string{ {{- range $k, $v := .RefAliasMap}}"{{$k}}": "{{$v}}",{{end}} }{{else}}nil{{end}},
			Index:        []int{ {{.Index}} },
		},
{{- end}}
	}
	jBase, err := cruds.MakeJoinerBase(joinTables, d)
	if err != nil {
		panic(err)
	}
	return Joiner{{.TypeInfo.Name}}{
		dialect:    d,
		allFields:  jBase.AllFields(),
		getOneSQL:  jBase.OneSQL(),
		getManySQL: jBase.ManySQL(),
		sortSQL:    jBase.SortSQL(),
	}
}

func NewJoiner{{.TypeInfo.Name}}(d dialect.SQLDialect) *Joiner{{.TypeInfo.Name}} {
	v := NewJoiner{{.TypeInfo.Name}}Val(d)
	return &v
}

func (j *Joiner{{.TypeInfo.Name}}) makeRefs(buf *{{.QualTypeName}}) ([]any, func()) {
{{- range $pv := .PointerVars}}
{{- range $f := $pv.Fields}}
	var {{$pv.VarName}}_{{$f.AccessExpr}} *{{$f.GoType}}
{{- end}}
{{- end}}
	return []any{
{{- range $st := .SubTables}}
{{- if $st.IsPointer}}
{{- range $f := $st.SelectableFields}}
		&_pr{{$st.ComputedAlias}}_{{$f.AccessExpr}},
{{- end}}
{{- else}}
{{- $access := $st.AccessExpr}}
{{- range $f := $st.SelectableFields}}
		&buf.{{$access}}.{{$f.AccessExpr}},
{{- end}}
{{- end}}
{{- end}}
	}, func() {
{{- range $st := .SubTables}}
{{- if $st.IsPointer}}
		buf.{{$st.AccessExpr}} = nil
		if _pr{{$st.ComputedAlias}}_{{(index $st.SelectableFields 0).AccessExpr}} != nil {
			buf.{{$st.AccessExpr}} = new({{$st.QualSubTypeName}})
{{- range $f := $st.SelectableFields}}
			buf.{{$st.AccessExpr}}.{{$f.AccessExpr}} = *_pr{{$st.ComputedAlias}}_{{$f.AccessExpr}}
{{- end}}
		}
{{- end}}
{{- end}}
	}
}

func (j *Joiner{{.TypeInfo.Name}}) One(ctx context.Context, tx cruds.TxProcessor, keys ...any) (*{{.QualTypeName}}, error) {
	buf := new({{.QualTypeName}})
	refs, apply := j.makeRefs(buf)
	err := tx.QueryRowContext(ctx, j.getOneSQL, keys...).Scan(refs...)
	apply()
	return buf, err
}

func (j *Joiner{{.TypeInfo.Name}}) Many(ctx context.Context, tx cruds.TxProcessor, filter *cruds.Filter) (result []*{{.QualTypeName}}, err error) {
	query := j.getManySQL
	var args []any
	if filter != nil {
		if filter.Range != nil {
			var sb strings.Builder
			sb.WriteString(j.getManySQL)
			sb.WriteString(defs.SQLWhere)
			var argIdx int
			where, whereArgs, werr := filter.Range.Build(j.allFields, j.dialect, &argIdx)
			if werr != nil {
				return nil, werr
			}
			sb.WriteString(where)
			query = sb.String()
			args = whereArgs
		}
		if filter.Offset > 0 || filter.Limit > 0 {
			query += j.dialect.OffsetAndLimit(filter.Offset, filter.Limit)
		}
	}
	query += j.sortSQL

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	buf := new({{.QualTypeName}})
	refs, apply := j.makeRefs(buf)
	for rows.Next() {
		if err = rows.Scan(refs...); err != nil {
			return nil, err
		}
		apply()
		rec := new({{.QualTypeName}})
		*rec = *buf
		result = append(result, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
`
