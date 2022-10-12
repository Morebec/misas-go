package gogenerator

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/morebec/misas-go/spectool/builtin"
	"github.com/morebec/misas-go/spectool/maputils"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// GoMod represents a go.mod file
type GoMod struct {
	Name string
	Path string
}

// GoPackage represents a Go Package.
type GoPackage struct {
	ModFile        GoMod
	Parent         *GoPackage
	Name           string
	FilePath       string
	SubPackages    []*GoPackage
	generatedFiles map[string]*GeneratedGoFile
}

// Root Returns the root node this GoPackage is part of.
// If the node is the root node, the node itself is returned.
func (p *GoPackage) Root() *GoPackage {
	if p.Parent == nil {
		return p
	}
	return p.Parent.Root()
}

// FindPackageForPath returns the GoPackage that can correspond to a certain path starting from this Package and going down.
func (p *GoPackage) FindPackageForPath(path string) *GoPackage {
	return p.SearchTreePreorderTraversal(func(node *GoPackage) bool {
		return node.FilePath == filepath.Dir(path)
	})
}

// SearchTreePreorderTraversal performs a using preorder traversal search on this package tree.
func (p *GoPackage) SearchTreePreorderTraversal(selectFunc func(node *GoPackage) bool) *GoPackage {
	if selectFunc(p) {
		return p
	}

	for _, c := range p.SubPackages {
		if found := c.SearchTreePreorderTraversal(selectFunc); found != nil {
			return found
		}
	}

	return nil
}

// ImportPath Returns the import path of this Package.
func (p *GoPackage) ImportPath() string {
	if p.Parent == nil {
		return p.ModFile.Name + "/" + p.Name
	}

	return p.Parent.ImportPath() + "/" + p.Name
}

// GeneratedFiles Returns the files of this package.
func (p *GoPackage) GeneratedFiles() []*GeneratedGoFile {
	return maputils.MapValues(p.generatedFiles)
}

// GeneratedFilesRecursive returns the files of this package and its subpackages and their subpackages...
func (p *GoPackage) GeneratedFilesRecursive() []*GeneratedGoFile {
	files := p.GeneratedFiles()
	for _, sub := range p.SubPackages {
		files = append(files, sub.GeneratedFilesRecursive()...)
	}
	return files
}

// GeneratedTypes Returns the go types that were used in this package.
func (p *GoPackage) GeneratedTypes() []GoType {
	var types []GoType
	for _, f := range p.generatedFiles {
		types = append(types, f.GeneratedTypes()...)
	}

	return types
}

// GeneratedTypesRecursive returns the go types that were generated in this package and its subpackages and their subpackages...
func (p *GoPackage) GeneratedTypesRecursive() []GoType {
	types := p.GeneratedTypes()
	for _, sub := range p.SubPackages {
		types = append(types, sub.GeneratedTypesRecursive()...)
	}
	return types
}

// AddFile Adds a file to this package. If the file already exists, it is overwritten by the provided object.
func (p *GoPackage) AddFile(f *GeneratedGoFile) {
	if p.generatedFiles == nil {
		p.generatedFiles = map[string]*GeneratedGoFile{}
	}
	p.generatedFiles[f.Path] = f
}

// FindGeneratedFileAtPath Returns a GeneratedGoFile at a given path in this package's descending hierarchy.
func (p *GoPackage) FindGeneratedFileAtPath(path string) *GeneratedGoFile {
	for _, f := range p.GeneratedFilesRecursive() {
		if f.Path == path {
			return f
		}
	}

	return nil
}

// BuildGoPackageTree builds a GoPackageTree from a GoMod file.
func BuildGoPackageTree(goMod GoMod) (*GoPackage, error) {
	root, err := BuildGoPackage(filepath.Dir(goMod.Path), nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not build go package tree")
	}

	root.ModFile = goMod

	return root, nil
}

// BuildGoPackage builds a GoPackage from a certain path as a child of a provide parent GoPackage.
func BuildGoPackage(path string, parent *GoPackage) (*GoPackage, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file info at \"%s\"", path)
	}

	if !fileInfo.IsDir() {
		return nil, errors.Errorf("cannot load non-directory at \"%s\"", path)
	}

	var modFile GoMod
	if parent != nil {
		modFile = parent.ModFile
	}

	node := &GoPackage{
		ModFile:     modFile,
		Parent:      parent,
		Name:        fileInfo.Name(),
		FilePath:    path,
		SubPackages: nil,
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading directory at \"%s\"", path)
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		nPath, err := filepath.Abs(path + "/" + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "failed loading node at \"%s\"", path)
		}
		if n, err := BuildGoPackage(nPath, node); err != nil {
			return nil, errors.Wrapf(err, "failed loading node at \"%s\"", path)
		} else {
			node.SubPackages = append(node.SubPackages, n)
		}
	}

	return node, nil
}

// FindGoMod finds the go.mod file using the system Spec's location.
func FindGoMod(systemSpec spec.Spec) (GoMod, error) {
	if systemSpec.Type != processing.SystemSpecType {
		return GoMod{}, spec.UnexpectedSpecTypeError(systemSpec.Type, processing.SystemSpecType)
	}

	goModPath := filepath.Dir(systemSpec.Source.Location) + "/go.mod"
	f, err := os.OpenFile(goModPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return GoMod{}, errors.Wrapf(err, "failed reading go.mod file %s", goModPath)
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text() // GET the line string
		if strings.HasPrefix(line, "module ") {
			abs, err := filepath.Abs(goModPath)
			if err != nil {
				return GoMod{}, errors.Wrapf(err, "failed reading go.mod file %s", goModPath)
			}

			return GoMod{
				Name: strings.ReplaceAll(line, "module ", ""),
				Path: abs,
			}, nil
		}
	}
	if err := sc.Err(); err != nil {
		return GoMod{}, errors.Wrapf(err, "failed reading go.mod file %s", goModPath)
	}

	return GoMod{}, errors.Errorf("invalid go.mod file: no module directive could be found at \"%s\"", goModPath)
}

// GoType represents a type in Go.
type GoType struct {
	TypeName         string
	InternalTypeName typesystem.DataType
	ImportPath       string
}

func NewGoType(typeName string, internalTypeName typesystem.DataType, importPath string) GoType {
	return GoType{TypeName: typeName, InternalTypeName: internalTypeName, ImportPath: importPath}
}

// GeneratedGoFile represents a file with generated code.
type GeneratedGoFile struct {
	Package  *GoPackage
	Path     string
	Snippets []GoSnippet
}

// AddSnippet adds a snippet to this file.
func (f *GeneratedGoFile) AddSnippet(snippet GoSnippet) {
	f.Snippets = append(f.Snippets, snippet)
}

// GeneratedTypes Returns the types generated in this file.
func (f *GeneratedGoFile) GeneratedTypes() []GoType {
	var types []GoType
	for _, s := range f.Snippets {
		types = append(types, s.GeneratedTypes...)
	}
	return types
}

// TypesUsed returns the list of GoTypes used by this file.
func (f *GeneratedGoFile) TypesUsed() []GoType {
	used := map[string]GoType{}

	for _, s := range f.Snippets {
		for _, ut := range s.TypesUsed {
			used[ut.TypeName] = ut
		}
	}

	return maputils.MapValues(used)
}

// GoSnippet represents a piece of generated code.
type GoSnippet struct {
	Code string
	// List of Go types that are generated by this snippet.
	GeneratedTypes []GoType
	// Returns a list of the types that are used by this snippet.
	TypesUsed     []GoType
	StaticImports []string
}

// GoSnippetGenerationContext represents template to generate a GoSnippet.
type GoSnippetGenerationContext struct {
	ParentContext *GoProcessingContext

	// Name of the template for debugging purposes.
	TemplateName string

	TemplateCode string

	TemplateData any

	// List of Go types generated by the snippet.
	GeneratedTypes []GoType

	// List of the types that are used by this template. This field will be automatically populated when calling the ResolveGoType function.
	TypesUsed []GoType

	StaticImports []string
}

func NewGoSnippetGenerationContext(
	parentContext *GoProcessingContext,
	templateName string,
	templateCode string,
	templateData any,
	generatedTypes []GoType,
	staticImports []string,
) *GoSnippetGenerationContext {
	return &GoSnippetGenerationContext{
		ParentContext:  parentContext,
		TemplateName:   templateName,
		TemplateCode:   templateCode,
		TemplateData:   templateData,
		GeneratedTypes: generatedTypes,
		StaticImports:  staticImports,
	}
}

func (ctx *GoSnippetGenerationContext) PackageTree() *GoPackage {
	return ctx.ParentContext.PackageTree
}

type GoProcessingContext struct {
	ParentContext *processing.Context
	PackageTree   *GoPackage
}

func (ctx *GoProcessingContext) Specs() spec.Group {
	return ctx.ParentContext.Specs
}

// GenerateCodeForSpec generates some go code for a given spec using a template and some data.
// It adds the resulting file and snippets to the GoProcessingContext.
func GenerateCodeForSpec(ctx *GoSnippetGenerationContext, spec spec.Spec) error {
	// Find target package
	pkg := ctx.PackageTree().FindPackageForPath(spec.Source.Location)
	if pkg == nil {
		return errors.Errorf("failed generating code for %s %s, could not find a suitable package", spec.Type, spec.TypeName)
	}

	// Determine file to write the snippet to
	fileName := "generated.go"
	if spec.Annotations.HasKey("gen:go:fileName") {
		fileName = spec.Annotations.GetOrDefault("gen:go:fileName", fileName).(string)
	} else if commandAggregateName := extractAggregateName(spec.TypeName); commandAggregateName != "" {
		fileName = commandAggregateName + "_" + fileName
	}
	filePath := pkg.FilePath + "/" + fileName

	file := pkg.FindGeneratedFileAtPath(filePath)
	if file == nil {
		file = &GeneratedGoFile{
			Package:  pkg,
			Path:     filePath,
			Snippets: nil,
		}
		pkg.AddFile(file)
	}

	// Generate snippet
	snippet, err := GenerateSnippet(ctx)
	if err != nil {
		return err
	}

	file.AddSnippet(snippet)

	return nil
}

// GenerateSnippet generates a GoSnippet from a GoSnippetGenerationContext.
func GenerateSnippet(ctx *GoSnippetGenerationContext) (GoSnippet, error) {

	acronyms := map[string]struct{}{
		"URL":  {},
		"ID":   {},
		"HTTP": {},
	}

	t := template.New("template " + ctx.TemplateName).Funcs(map[string]any{

		// converts a string so that it adheres to the exported type naming scheme of go.
		// This can be useful for type names, struct field names, and constants.
		"AsExportedGoName": func(value string) string {
			upper := strcase.ToCamel(value)
			re := regexp.MustCompile(`[A-Z][^A-Z]*`)
			matches := re.FindAllString(upper, -1)
			final := ""
			for _, element := range matches {
				upperElem := strings.ToUpper(element)
				if _, found := acronyms[upperElem]; found {
					final += upperElem
				} else {
					final += element
				}
			}
			return final
		},

		// converts a string so that it adheres to the non exported type naming scheme of go.
		// This can be useful for type names, struct field names, and constants.
		"AsUnexportedGoName": func(value string) string {
			if value == "id" {
				return "ID"
			}
			return strcase.ToLowerCamel(value)
		},

		// Converts a DataType to a GoType.
		"AsResolvedGoType": func(t typesystem.DataType) string {
			rgt, err := ResolveGoType(ctx, t)
			if err != nil {
				panic(err)
			}
			return rgt.TypeName
		},
		"AsJsonAnnotation": func(fieldName string) string {

			if fieldName != "id" {
				fieldName = strcase.ToLowerCamel(fieldName)
			}

			return fmt.Sprintf("`json:\"%s\"`", fieldName)
		},
	})

	t, err := t.Parse(ctx.TemplateCode)
	if err != nil {
		return GoSnippet{}, errors.Wrap(err, "failed generating go code")
	}

	b := bytes.Buffer{}
	if err := t.Execute(&b, ctx.TemplateData); err != nil {
		return GoSnippet{}, errors.Wrap(err, "failed generating go code")
	}

	return GoSnippet{
		Code:           b.String(),
		GeneratedTypes: ctx.GeneratedTypes,
		TypesUsed:      ctx.TypesUsed,
		StaticImports:  ctx.StaticImports,
	}, nil
}

// ResolveGoType Resolves a GoType from an internal Type in a GoSnippetGenerationContext.
func ResolveGoType(ctx *GoSnippetGenerationContext, t typesystem.DataType) (GoType, error) {
	doResolve := func() (GoType, error) {
		switch t {
		case typesystem.Null:
			return NewGoType("nil", typesystem.Null, ""), nil
		case typesystem.Identifier:
			return NewGoType("string", typesystem.Identifier, ""), nil
		case typesystem.String:
			return NewGoType("string", typesystem.String, ""), nil
		case typesystem.Bool:
			return NewGoType("bool", typesystem.Bool, ""), nil
		case typesystem.Int:
			return NewGoType("int64", typesystem.Int, ""), nil
		case typesystem.Float:
			return NewGoType("float64", typesystem.Float, ""), nil
		case typesystem.Date:
			return NewGoType("time.Time", typesystem.Date, "time"), nil
		case typesystem.DateTime:
			return NewGoType("time.Time", typesystem.DateTime, "time"), nil
		case typesystem.Duration:
			return NewGoType("time.Duration", typesystem.Duration, "time"), nil
		case typesystem.Char:
			return NewGoType("rune", typesystem.Char, ""), nil
		}

		if t.IsContainer() {
			resolved, err := ResolveGoType(ctx, t.BaseType())
			if err != nil {
				return GoType{}, errors.Wrapf(err, "failed resolving container type %s", t)
			}

			if t.IsMap() {
				keyType, err := ResolveGoType(ctx, t.ContainerInfo().KeyType)
				if err != nil {
					return GoType{}, errors.Wrapf(err, "failed resolving key of container type %s", t)
				}
				resolved.TypeName = fmt.Sprintf("map[%s]%s", keyType.TypeName, resolved.TypeName)
				return resolved, nil
			}

			resolved.TypeName = fmt.Sprintf("[]%s", resolved.TypeName)
			return resolved, nil
		}

		// User defined type.
		for _, gt := range ctx.PackageTree().GeneratedTypesRecursive() {
			if gt.InternalTypeName == t {
				return gt, nil
			}
		}

		return GoType{}, errors.Errorf("Could not resolve a Go type for \"%s\"", t)
	}
	resolved, err := doResolve()
	if err != nil {
		return GoType{}, err
	}

	ctx.TypesUsed = append(ctx.TypesUsed, resolved)
	return resolved, nil
}

// RenderGeneratedFile Renders and Formats a Generated File as a string
func RenderGeneratedFile(f GeneratedGoFile) (string, error) {
	// Resolve imports
	var imports []string
	for _, ut := range f.TypesUsed() {
		if ut.ImportPath != "" {
			imports = append(imports, ut.ImportPath)
		}
	}

	// Generate Header of file.
	header := fmt.Sprintf("// IMPORTANT: This file was auto-generated by the morebec/spectool program. Do not edit manually. \n\n")
	header += fmt.Sprintf("package %s\n\n", f.Package.Name)
	if imports != nil {
		header += fmt.Sprintf("import (\n\"%s\"\n)\n", strings.Join(imports, "\"\n\n\""))
	}

	code := "\n"
	for _, snip := range f.Snippets {
		code = snip.Code + code
	}

	source, err := FormatGoSource([]byte(header + code))
	if err != nil {
		return "", err
	}

	return string(source), nil
}

// FormatGoSource Formats Go Source Code.
func FormatGoSource(content []byte) ([]byte, error) {

	formattedContent, err := format.Source(content)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed formatting go source")
	}

	return formattedContent, nil
}

// extractAggregateName extracts the name of an aggregate for a SpecificationTypeName of a Command/Query/Payload.
// E.g. website.add -> website.
func extractAggregateName(name spec.TypeName) string {
	parts := strings.Split(string(name), ".")

	if len(parts) == 0 {
		return string(name)
	}

	aggName := parts[len(parts)-2]
	return aggName

}

// GoProcessor processes go code generation.
func GoProcessor() processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {

		goMod, err := FindGoMod(ctx.SystemSpec)
		if err != nil {
			return errors.Wrap(err, "go processor failed")
		}

		tree, err := BuildGoPackageTree(goMod)
		if err != nil {
			return errors.Wrap(err, "go processor failed")
		}

		gCtx := &GoProcessingContext{
			ParentContext: ctx,
			PackageTree:   tree,
		}

		processingHandlers := map[spec.Type]func(ctx *GoProcessingContext, s spec.Spec) error{
			builtin.CommandType:      generateCommand,
			builtin.QueryType:        generateQuery,
			builtin.EventType:        generateEvent,
			builtin.StructType:       generateStruct,
			builtin.EnumType:         generateEnum,
			builtin.HTTPEndpointType: generateHTTPEndpoint,
		}

		for _, dep := range ctx.Dependencies {
			if fun, found := processingHandlers[dep.Spec.Type]; found {
				if err := fun(gCtx, dep.Spec); err != nil {
					return errors.Wrap(err, "go processor failed")
				}
			}
		}

		// Convert go files to OutputFiles
		ctx.Logger.Info("Generating Go code ...")
		for _, gf := range gCtx.PackageTree.GeneratedFilesRecursive() {
			code, err := RenderGeneratedFile(*gf)
			if err != nil {
				return errors.Wrap(err, "go processor failed")
			}
			ctx.AddOutputFile(processing.OutputFile{
				Path: gf.Path,
				Data: []byte(code),
			})
		}
		ctx.Logger.Info("Go code generated successfully.")

		return nil
	}
}

func generateStruct(ctx *GoProcessingContext, evt spec.Spec) error {
	templateCode := `
const {{ .StructName }}TypeName string = "{{ .TypeName }}"

// {{ .StructName }} {{ .Description }}
type {{ .StructName }} struct {
	{{ range $field := .Fields }}
		// {{ $field.Description }} {{ if $field.Annotations.HasKey "personal_data" }}
		// NOTE: This field contains personal data{{ end }}
		{{ $field.Name | AsExportedGoName }} {{ $field.Type | AsResolvedGoType }} {{ $field.Name | AsJsonAnnotation }}
	{{ end }}
}

func (c {{ .StructName }}) TypeName() string {
	return {{ .StructName }}TypeName
}
`

	type TemplateData struct {
		Package     string
		Imports     []string
		StructName  string
		TypeName    string
		FilePath    string
		Fields      map[string]builtin.StructField
		Description string
	}

	// Generate Go Code Snippet
	templateData := TemplateData{
		StructName:  evt.Annotations.GetOrDefault("gen:go:name", strcase.ToCamel(string(evt.TypeName))).(string),
		Description: strings.ReplaceAll(strings.TrimSuffix(evt.Description, "\n"), "\n", "\n// "),
		TypeName:    string(evt.TypeName),
		Fields:      evt.Properties.(builtin.StructSpecProperties).Fields,
	}

	//goland:noinspection GoRedundantConversion
	tem := NewGoSnippetGenerationContext(
		ctx,
		"struct",
		templateCode,
		templateData,
		[]GoType{
			{
				TypeName:         string(templateData.StructName),
				InternalTypeName: typesystem.DataType(evt.TypeName),
				ImportPath:       "",
			},
		},
		[]string{},
	)

	return GenerateCodeForSpec(tem, evt)
}

func generateEnum(ctx *GoProcessingContext, enum spec.Spec) error {
	props := enum.Properties.(builtin.EnumSpecProperties)

	templateCode := `
// {{ .EnumName }} {{ .Description }}
type {{ .EnumName }} {{ .EnumBaseType | AsResolvedGoType }}

const ({{ range $value := .Values }}
	// {{ .Name | AsExportedGoName }} {{ $value.Description }}
	{{ .Name | AsExportedGoName }} {{ $.EnumBaseType | AsResolvedGoType }} =
	{{ if eq $.EnumBaseType "string" }} "{{ .Value }}" {{ else }} {{ .Value }} {{ end }}
{{ end }})
`
	type TemplateData struct {
		EnumName     string
		EnumBaseType typesystem.DataType
		TypeName     string
		Values       map[string]builtin.EnumValue
		Description  string
	}

	// Generate Go Code Snippet
	//goland:noinspection GoRedundantConversion
	templateData := TemplateData{
		EnumName:     enum.Annotations.GetOrDefault("gen:go:name", strcase.ToCamel(string(enum.TypeName))).(string),
		Description:  strings.ReplaceAll(strings.TrimSuffix(enum.Description, "\n"), "\n", "\n// "),
		TypeName:     string(enum.TypeName),
		EnumBaseType: typesystem.DataType(props.BaseType),
		Values:       props.Values,
	}

	//goland:noinspection GoRedundantConversion
	tem := NewGoSnippetGenerationContext(
		ctx,
		"enum",
		templateCode,
		templateData,
		[]GoType{
			{
				TypeName:         string(templateData.EnumName),
				InternalTypeName: typesystem.DataType(enum.TypeName),
				ImportPath:       "",
			},
		},
		[]string{},
	)

	return GenerateCodeForSpec(tem, enum)
}

// generates the Go Code for a command.Command.
func generateCommand(ctx *GoProcessingContext, cmd spec.Spec) error {
	templateCode := `
const {{ .StructName }}TypeName command.TypeName = "{{ .TypeName }}"

// {{ .StructName }} {{ .Description }}
type {{ .StructName }} struct {
	{{ range $field := .Fields }}
		// {{ $field.Description }} {{ if $field.Annotations.HasKey "personal_data" }}
		// NOTE: This field contains personal data{{ end }}
		{{ $field.Name | AsExportedGoName }} {{ $field.Type | AsResolvedGoType }} {{ $field.Name | AsJsonAnnotation }}
	{{ end }}
}

func (c {{ .StructName }}) TypeName() command.TypeName {
	return {{ .StructName }}TypeName
}
`

	type TemplateData struct {
		Package     string
		Imports     []string
		StructName  string
		TypeName    string
		FilePath    string
		Fields      map[string]builtin.CommandField
		Description string
	}

	// Generate Go Code Snippet
	templateData := TemplateData{
		StructName:  cmd.Annotations.GetOrDefault("gen:go:name", strcase.ToCamel(string(cmd.TypeName))+"Command").(string),
		Description: strings.ReplaceAll(strings.TrimSuffix(cmd.Description, "\n"), "\n", "\n// "),
		TypeName:    string(cmd.TypeName),
		Fields:      cmd.Properties.(builtin.CommandSpecProperties).Fields,
	}

	//goland:noinspection GoRedundantConversion
	tem := NewGoSnippetGenerationContext(
		ctx,
		"command",
		templateCode,
		templateData,
		[]GoType{
			{
				TypeName:         string(templateData.StructName),
				InternalTypeName: typesystem.DataType(cmd.TypeName),
				ImportPath:       "",
			},
		},
		[]string{
			"github.com/morebec/misas-go/misas/command",
		},
	)

	return GenerateCodeForSpec(tem, cmd)
}

// generates the Go Code for a query.Query.
func generateQuery(ctx *GoProcessingContext, query spec.Spec) error {
	templateCode := `
const {{ .StructName }}TypeName query.TypeName = "{{ .TypeName }}"

// {{ .StructName }} {{ .Description }}
type {{ .StructName }} struct {
	{{ range $field := .Fields }}
		// {{ $field.Description }} {{ if $field.Annotations.HasKey "personal_data" }}
		// NOTE: This field contains personal data{{ end }}
		{{ $field.Name | AsExportedGoName }} {{ $field.Type | AsResolvedGoType }} {{ $field.Name | AsJsonAnnotation }}
	{{ end }}
}

func (c {{ .StructName }}) TypeName() query.TypeName {
	return {{ .StructName }}TypeName
}
`

	type TemplateData struct {
		Package     string
		Imports     []string
		StructName  string
		TypeName    string
		FilePath    string
		Fields      map[string]builtin.QueryField
		Description string
	}

	// Generate Go Code Snippet
	templateData := TemplateData{
		StructName:  query.Annotations.GetOrDefault("gen:go:name", strcase.ToCamel(string(query.TypeName))+"Query").(string),
		Description: strings.ReplaceAll(strings.TrimSuffix(query.Description, "\n"), "\n", "\n// "),
		TypeName:    string(query.TypeName),
		Fields:      query.Properties.(builtin.QuerySpecProperties).Fields,
	}

	//goland:noinspection GoRedundantConversion
	tem := NewGoSnippetGenerationContext(
		ctx,
		"query",
		templateCode,
		templateData,
		[]GoType{
			{
				TypeName:         string(templateData.StructName),
				InternalTypeName: typesystem.DataType(query.TypeName),
				ImportPath:       "",
			},
		},
		[]string{
			"github.com/morebec/misas-go/misas/query",
		},
	)

	return GenerateCodeForSpec(tem, query)
}

// generates the Go Code for an event.Event.
func generateEvent(ctx *GoProcessingContext, evt spec.Spec) error {
	templateCode := `
const {{ .StructName }}TypeName event.TypeName = "{{ .TypeName }}"

// {{ .StructName }} {{ .Description }}
type {{ .StructName }} struct {
	{{ range $field := .Fields }}
		// {{ $field.Description }} {{ if $field.Annotations.HasKey "personal_data" }}
		// NOTE: This field contains personal data{{ end }}
		{{ $field.Name | AsExportedGoName }} {{ $field.Type | AsResolvedGoType }} {{ $field.Name | AsJsonAnnotation }}
	{{ end }}
}

func (c {{ .StructName }}) TypeName() event.TypeName {
	return {{ .StructName }}TypeName
}
`

	type TemplateData struct {
		Package     string
		Imports     []string
		StructName  string
		TypeName    string
		FilePath    string
		Fields      map[string]builtin.EventField
		Description string
	}

	// Generate Go Code Snippet
	templateData := TemplateData{
		StructName:  evt.Annotations.GetOrDefault("gen:go:name", strcase.ToCamel(string(evt.TypeName))+"Event").(string),
		Description: strings.ReplaceAll(strings.TrimSuffix(evt.Description, "\n"), "\n", "\n// "),
		TypeName:    string(evt.TypeName),
		Fields:      evt.Properties.(builtin.EventSpecProperties).Fields,
	}

	//goland:noinspection GoRedundantConversion
	tem := NewGoSnippetGenerationContext(
		ctx,
		"event",
		templateCode,
		templateData,
		[]GoType{
			{
				TypeName:         string(templateData.StructName),
				InternalTypeName: typesystem.DataType(evt.TypeName),
				ImportPath:       "",
			},
		},
		[]string{
			"github.com/morebec/misas-go/misas/event",
		},
	)

	return GenerateCodeForSpec(tem, evt)
}

// generates the Go Code for an HTTP Endpoint.
func generateHTTPEndpoint(ctx *GoProcessingContext, endpoint spec.Spec) error {
	props := endpoint.Properties.(builtin.HTTPEndpointSpecProperties)

	templateCode := `
// {{ .EndpointFuncName }} {{ .Description }}
func {{ .EndpointFuncName }}(r chi.Router, bus {{ if eq .Method "POST" }}command.Bus{{ else }}event.Bus{{ end }}) {
	r.Get("{{ .Path }}", func(w http.ResponseWriter, r *http.Request) {
		handleError := func(w http.ResponseWriter, r *http.Request, err error) {
			if !domain.IsDomainError(err) {
				w.WriteHeader(500)
				render.JSON(w, r, NewInternalError(err))
			}

			derr := err.(domain.Error)

			conv := map[domain.ErrorTypeName]int {
				{{ range $response := .FailureResponses }}{{ if ne .Description "" }}// {{ .Description }}{{ end }}
				"{{ .ErrorType }}": {{ .StatusCode }},
				{{ end }}
			}
			c := conv[derr.TypeName()]
			w.WriteHeader(c)
			render.JSON(w, r, NewErrorResponse(derr.TypeName(), derr.Error(), derr.Data()))
		}


		// Decode request payload
		var input {{ .Request | AsResolvedGoType }}
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			w.WriteHeader(500)
			render.JSON(w, r, httpapi.NewInternalError(err))
			return
		}

		// Send to Domain Layer
		output, err := bus.Send(r.Context(), input)
		if err != nil {
			w.WriteHeader(400)
			render.JSON(w, r, httpapi.NewErrorResponse(output))
			return
		}

		render.JSON(w, r, httpapi.NewSuccessResponse(output))
	})
}
`
	type TemplateData struct {
		EndpointFuncName string
		TypeName         string
		Description      string
		Path             string
		Method           string
		Request          typesystem.DataType
		SuccessResponse  builtin.HTTPEndpointSuccessResponse
		FailureResponses []builtin.HTTPEndpointFailureResponse
	}

	// Generate Go Code Snippet
	//goland:noinspection GoRedundantConversion
	templateData := TemplateData{
		EndpointFuncName: endpoint.Annotations.GetOrDefault("gen:go:name", strcase.ToLowerCamel(string(endpoint.TypeName))).(string),
		TypeName:         string(endpoint.TypeName),
		Description:      strings.ReplaceAll(strings.TrimSuffix(endpoint.Description, "\n"), "\n", "\n// "),
		Path:             props.Path,
		Method:           props.Method,
		Request:          props.Request,
		SuccessResponse:  props.Responses.Success,
		FailureResponses: props.Responses.Failures,
	}

	//goland:noinspection GoRedundantConversion
	tem := NewGoSnippetGenerationContext(
		ctx,
		"endpoint",
		templateCode,
		templateData,
		[]GoType{
			{
				TypeName:         string(templateData.EndpointFuncName),
				InternalTypeName: typesystem.DataType(endpoint.TypeName),
				ImportPath:       "",
			},
		},
		[]string{
			"encoding/json",
			"net/http",

			"github.com/go-chi/chi/v5",
			"github.com/go-chi/render",
			"github.com/morebec/misas-go/misas/httpapi",
			"github.com/morebec/misas-go/misas/command",
			"github.com/morebec/misas-go/misas/domain",
			"github.com/morebec/misas-go/misas/event",
		},
	)

	return GenerateCodeForSpec(tem, endpoint)
}
