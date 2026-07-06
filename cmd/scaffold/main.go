// Command scaffold generates a full CRUD vertical slice (model, repository,
// service, and HTTP handlers) for a resource, using its CREATE TABLE statement
// in sql/<table>.sql as the single source of truth.
//
// Usage:
//
//	go run ./cmd/scaffold <table>
//
// where sql/<table>.sql contains a `CREATE TABLE <table> (...)` definition.
// The table must have `id uuid`, `created_at`, and `updated_at` columns; every
// other column becomes an editable field on the resource. Generated files are
// written under internal/ and the new resource is wired into the router
// (internal/api/init.go) and DI container (cmd/httpserver/container.go).
package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Field is a single database column mapped to Go.
type Field struct {
	Column   string // db column name, snake_case
	GoName   string // exported Go field name, PascalCase
	ArgName  string // unexported identifier, camelCase (used for func params)
	GoType   string // Go type for the model/request/input
	JSONName string // json + db tag value
	Kind     string // uuid | time | string | number | bool
	NotNull  bool
	System   bool // id / created_at / updated_at
	Required bool // NotNull string field that gets a presence check
}

// OutField is a column as it appears on the *Output / *Response structs, where
// uuid and time values are rendered as strings.
type OutField struct {
	GoName   string
	GoType   string
	JSONName string
	Expr     string // mapping expression relative to receiver "m"
}

// Resource is the full template context for one table.
type Resource struct {
	Module         string
	Table          string
	SingularPascal string
	PluralPascal   string
	SingularCamel  string
	PluralCamel    string
	Var            string
	Fields         []Field
	UserFields     []Field
	OutFields      []OutField
	NeedsValidation bool

	InsertColumns      string
	InsertPlaceholders string
	SelectColumns      string
	UpdateSet          string
	InsertArgs         string
	UpdateArgs         string
}

func main() {
	if len(os.Args) < 2 {
		fatal("usage: go run ./cmd/scaffold <table>  (reads sql/<table>.sql)")
	}
	table := strings.ToLower(strings.TrimSpace(os.Args[1]))
	if table == "" {
		fatal("table name is required")
	}

	root := repoRoot()
	module := readModule(filepath.Join(root, "go.mod"))

	sqlPath := filepath.Join(root, "sql", table+".sql")
	ddl, err := os.ReadFile(sqlPath)
	if err != nil {
		fatal("cannot read %s: %v\n(write the CREATE TABLE first, then re-run)", sqlPath, err)
	}

	cols, err := parseTable(string(ddl), table)
	if err != nil {
		fatal("parsing %s: %v", sqlPath, err)
	}

	res, err := buildResource(module, table, cols)
	if err != nil {
		fatal("%v", err)
	}

	// Refuse to clobber an already-scaffolded resource.
	files := map[string]string{
		filepath.Join(root, "internal", "domain", "models", table+".go"):       "model.go.tmpl",
		filepath.Join(root, "internal", "domain", "repositories", table+".go"): "repository.go.tmpl",
		filepath.Join(root, "internal", "services", table+".go"):               "service.go.tmpl",
		filepath.Join(root, "internal", "api", table+".go"):                    "api.go.tmpl",
	}
	for path := range files {
		if _, err := os.Stat(path); err == nil {
			fatal("%s already exists — resource %q looks scaffolded already", rel(root, path), table)
		}
	}

	// Render + write new files.
	for path, tmpl := range files {
		if err := renderToFile(path, tmpl, res); err != nil {
			fatal("generating %s: %v", rel(root, path), err)
		}
		fmt.Println("created", rel(root, path))
	}

	// Wire the resource into the router and container via marker comments.
	if err := wire(root, res); err != nil {
		fatal("wiring: %v", err)
	}

	fmt.Printf("\nwired %q. next:\n  go build ./...\n  go run ./cmd/migrate   # applies sql/%s.sql\n", table, table)
}

// ---- DDL parsing -----------------------------------------------------------

var (
	lineComment  = regexp.MustCompile(`--[^\n]*`)
	blockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
)

// parseTable extracts the column definitions of `table` from a SQL script.
func parseTable(sqlText, table string) ([]Field, error) {
	s := blockComment.ReplaceAllString(sqlText, "")
	s = lineComment.ReplaceAllString(s, "")

	// Locate: CREATE TABLE [IF NOT EXISTS] <table> (
	head := regexp.MustCompile(`(?is)create\s+table\s+(?:if\s+not\s+exists\s+)?"?` + regexp.QuoteMeta(table) + `"?\s*\(`)
	loc := head.FindStringIndex(s)
	if loc == nil {
		return nil, fmt.Errorf("no `CREATE TABLE %s (...)` found", table)
	}

	// Scan from the opening paren to its match.
	open := loc[1] - 1
	depth, end := 0, -1
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		return nil, fmt.Errorf("unbalanced parentheses in CREATE TABLE %s", table)
	}
	body := s[open+1 : end]

	var fields []Field
	for _, def := range splitTopLevel(body) {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}
		if isTableConstraint(def) {
			continue
		}
		f, err := parseColumn(def)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("no columns found in CREATE TABLE %s", table)
	}
	return fields, nil
}

// splitTopLevel splits a column list on commas that are not inside parentheses.
func splitTopLevel(body string) []string {
	var parts []string
	depth, start := 0, 0
	for i, r := range body {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, body[start:])
	return parts
}

var constraintKeywords = []string{"primary", "unique", "foreign", "constraint", "check", "exclude", "like"}

func isTableConstraint(def string) bool {
	first := strings.ToLower(strings.Fields(def)[0])
	for _, k := range constraintKeywords {
		if first == k {
			return true
		}
	}
	return false
}

// typeRule maps a SQL type prefix to a Go type and kind. Ordered most-specific
// first so that e.g. "int8" wins over "int".
var typeRules = []struct{ sql, goType, kind string }{
	{"uuid", "uuid.UUID", "uuid"},
	{"bigserial", "int64", "number"},
	{"smallserial", "int", "number"},
	{"serial", "int", "number"},
	{"int8", "int64", "number"},
	{"int4", "int", "number"},
	{"int2", "int", "number"},
	{"bigint", "int64", "number"},
	{"smallint", "int", "number"},
	{"integer", "int", "number"},
	{"int", "int", "number"},
	{"float8", "float64", "number"},
	{"float4", "float64", "number"},
	{"double precision", "float64", "number"},
	{"numeric", "float64", "number"},
	{"decimal", "float64", "number"},
	{"real", "float64", "number"},
	{"money", "float64", "number"},
	{"boolean", "bool", "bool"},
	{"bool", "bool", "bool"},
	{"timestamptz", "time.Time", "time"},
	{"timestamp with time zone", "time.Time", "time"},
	{"timestamp without time zone", "time.Time", "time"},
	{"timestamp", "time.Time", "time"},
	{"date", "time.Time", "time"},
	{"time", "time.Time", "time"},
	{"character varying", "string", "string"},
	{"varchar", "string", "string"},
	{"character", "string", "string"},
	{"citext", "string", "string"},
	{"char", "string", "string"},
	{"text", "string", "string"},
	{"jsonb", "string", "string"},
	{"json", "string", "string"},
}

func parseColumn(def string) (Field, error) {
	fields := strings.Fields(def)
	col := strings.Trim(fields[0], `"`)
	rest := strings.ToLower(strings.TrimSpace(def[len(fields[0]):]))

	goType, kind := "", ""
	for _, r := range typeRules {
		if hasTypePrefix(rest, r.sql) {
			goType, kind = r.goType, r.kind
			break
		}
	}
	if goType == "" {
		return Field{}, fmt.Errorf("column %q: unrecognized SQL type in %q", col, def)
	}

	notNull := strings.Contains(rest, "not null") ||
		strings.Contains(rest, "primary key") ||
		strings.Contains(rest, "serial")

	system := col == "id" || col == "created_at" || col == "updated_at"
	return Field{
		Column:   col,
		GoName:   pascal(col),
		ArgName:  camel(col),
		GoType:   goType,
		JSONName: col,
		Kind:     kind,
		NotNull:  notNull,
		System:   system,
		Required: notNull && kind == "string" && !system,
	}, nil
}

// hasTypePrefix reports whether rest begins with the SQL type token `t`,
// respecting a token boundary so "int" does not match "integer" spuriously
// in a way that changes meaning.
func hasTypePrefix(rest, t string) bool {
	if !strings.HasPrefix(rest, t) {
		return false
	}
	if len(rest) == len(t) {
		return true
	}
	switch rest[len(t)] {
	case ' ', '\t', '\n', '(', ',':
		return true
	}
	return false
}

// ---- resource assembly -----------------------------------------------------

func buildResource(module, table string, fields []Field) (*Resource, error) {
	has := map[string]bool{}
	for _, f := range fields {
		has[f.Column] = true
	}
	for _, req := range []string{"id", "created_at", "updated_at"} {
		if !has[req] {
			return nil, fmt.Errorf("table %q must have a %q column (scaffold convention)", table, req)
		}
	}

	singular := singularize(table)
	res := &Resource{
		Module:         module,
		Table:          table,
		SingularPascal: pascal(singular),
		PluralPascal:   pascal(table),
		SingularCamel:  camel(singular),
		PluralCamel:    camel(table),
		Var:            camel(singular),
		Fields:         fields,
	}

	var insertCols, placeholders, insertArgs []string
	for i, f := range fields {
		insertCols = append(insertCols, f.Column)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		insertArgs = append(insertArgs, res.Var+"."+f.GoName)
		if !f.System {
			res.UserFields = append(res.UserFields, f)
		}
		if f.Required {
			res.NeedsValidation = true
		}
	}
	res.InsertColumns = strings.Join(insertCols, ", ")
	res.SelectColumns = res.InsertColumns
	res.InsertPlaceholders = strings.Join(placeholders, ", ")
	res.InsertArgs = strings.Join(insertArgs, ", ")

	// UPDATE ... SET <cols except id, created_at> WHERE id = $1
	updateArgs := []string{res.Var + ".Id"}
	var setClauses []string
	ph := 2
	for _, f := range fields {
		if f.Column == "id" || f.Column == "created_at" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", f.Column, ph))
		updateArgs = append(updateArgs, res.Var+"."+f.GoName)
		ph++
	}
	res.UpdateSet = strings.Join(setClauses, ", ")
	res.UpdateArgs = strings.Join(updateArgs, ", ")

	// Output/Response fields: id, user fields, created_at, updated_at.
	res.OutFields = append(res.OutFields, outFieldFor(field(fields, "id")))
	for _, f := range res.UserFields {
		res.OutFields = append(res.OutFields, outFieldFor(f))
	}
	res.OutFields = append(res.OutFields, outFieldFor(field(fields, "created_at")))
	res.OutFields = append(res.OutFields, outFieldFor(field(fields, "updated_at")))

	return res, nil
}

func field(fields []Field, col string) Field {
	for _, f := range fields {
		if f.Column == col {
			return f
		}
	}
	return Field{}
}

func outFieldFor(f Field) OutField {
	switch f.Kind {
	case "uuid":
		return OutField{f.GoName, "string", f.JSONName, "m." + f.GoName + ".String()"}
	case "time":
		return OutField{f.GoName, "string", f.JSONName, "m." + f.GoName + ".Format(timeFormat)"}
	default:
		return OutField{f.GoName, f.GoType, f.JSONName, "m." + f.GoName}
	}
}

// ---- wiring ----------------------------------------------------------------

const (
	initFile      = "internal/api/init.go"
	containerFile = "cmd/httpserver/container.go"
)

func wire(root string, res *Resource) error {
	inserts := []struct{ file, marker, tmpl string }{
		{initFile, "// scaffold:services", "\t{{.SingularCamel}}Service services.{{.SingularPascal}}Service"},
		{initFile, "// scaffold:opts", "\t{{.SingularPascal}}Service services.{{.SingularPascal}}Service"},
		{initFile, "// scaffold:assign", "\t\t{{.SingularCamel}}Service: o.{{.SingularPascal}}Service,"},
		{initFile, "// scaffold:routes", routesSnippet},
		{containerFile, "// scaffold:construct", constructSnippet},
		{containerFile, "// scaffold:apiopts", "\t\t{{.SingularPascal}}Service: {{.SingularCamel}}Service,"},
	}
	for _, ins := range inserts {
		snippet, err := render(ins.tmpl, res)
		if err != nil {
			return err
		}
		if err := insertBeforeMarker(filepath.Join(root, ins.file), ins.marker, snippet); err != nil {
			return err
		}
	}
	fmt.Println("wired", initFile, "and", containerFile)
	return nil
}

const routesSnippet = `	a.router.Methods(http.MethodPost).Path("/{{.Route}}").HandlerFunc(a.Create{{.SingularPascal}})
	a.router.Methods(http.MethodGet).Path("/{{.Route}}").HandlerFunc(a.List{{.PluralPascal}})
	a.router.Methods(http.MethodGet).Path("/{{.Route}}/{id}").HandlerFunc(a.Get{{.SingularPascal}})
	a.router.Methods(http.MethodPut).Path("/{{.Route}}/{id}").HandlerFunc(a.Update{{.SingularPascal}})
	a.router.Methods(http.MethodDelete).Path("/{{.Route}}/{id}").HandlerFunc(a.Delete{{.SingularPascal}})`

const constructSnippet = `	{{.PluralCamel}}Repo := repositories.New{{.PluralPascal}}Repository(repositories.{{.PluralPascal}}RepoOpts{DB: db})
	{{.SingularCamel}}Service := services.New{{.SingularPascal}}Service(services.{{.SingularPascal}}ServiceOpts{
		{{.PluralPascal}}Repo: {{.PluralCamel}}Repo,
	})`

// insertBeforeMarker inserts snippet on its own line(s) immediately above the
// line containing marker, then gofmt-formats the whole file. The marker is left
// in place so the next scaffold run can insert again.
func insertBeforeMarker(path, marker, snippet string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(data)
	idx := strings.Index(s, marker)
	if idx < 0 {
		return fmt.Errorf("marker %q not found in %s", marker, path)
	}
	lineStart := strings.LastIndex(s[:idx], "\n") + 1
	out := s[:lineStart] + snippet + "\n" + s[lineStart:]
	formatted, err := format.Source([]byte(out))
	if err != nil {
		return fmt.Errorf("formatting %s after insert: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0o644)
}

// ---- rendering + helpers ---------------------------------------------------

func renderToFile(path, tmplName string, res *Resource) error {
	// Route is only used by the wiring snippets, but harmless to expose.
	data := struct {
		*Resource
		Route string
	}{res, res.Table}

	t, err := template.ParseFS(templates, "templates/"+tmplName)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("gofmt: %w\n---\n%s", err, buf.String())
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, formatted, 0o644)
}

func render(tmpl string, res *Resource) (string, error) {
	data := struct {
		*Resource
		Route string
	}{res, res.Table}
	t, err := template.New("snippet").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func readModule(gomod string) string {
	data, err := os.ReadFile(gomod)
	if err != nil {
		fatal("cannot read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	fatal("no module directive in go.mod")
	return ""
}

// repoRoot walks up from the working directory to the go.mod that owns cmd/scaffold.
func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		fatal("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fatal("could not locate go.mod (run from within the project)")
		}
		dir = parent
	}
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}

// ---- naming ----------------------------------------------------------------

func pascal(snake string) string {
	parts := strings.Split(snake, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		// Match the codebase style: "id" -> "Id", not "ID".
		b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
	}
	return b.String()
}

func camel(snake string) string {
	p := pascal(snake)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

func singularize(plural string) string {
	switch {
	case strings.HasSuffix(plural, "ies"):
		return plural[:len(plural)-3] + "y"
	case strings.HasSuffix(plural, "ses"), strings.HasSuffix(plural, "xes"),
		strings.HasSuffix(plural, "zes"), strings.HasSuffix(plural, "ches"),
		strings.HasSuffix(plural, "shes"):
		return plural[:len(plural)-2]
	case strings.HasSuffix(plural, "ss"):
		return plural
	case strings.HasSuffix(plural, "s"):
		return plural[:len(plural)-1]
	default:
		return plural
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "scaffold: "+format+"\n", args...)
	os.Exit(1)
}
