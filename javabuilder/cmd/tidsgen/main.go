package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Schema struct {
	Metadata      Metadata         `yaml:"metadata" json:"metadata"`
	Types         []TypeDef        `yaml:"types" json:"types"`
	RetryPolicies []RetryPolicyDef `yaml:"retry_policies" json:"retry_policies"`
	Activities    []ActivityDef    `yaml:"activities" json:"activities"`
	Workflows     []WorkflowDef    `yaml:"workflows" json:"workflows"`
	Workers       []WorkerDef      `yaml:"workers" json:"workers"`
	Clients       []ClientDef      `yaml:"clients" json:"clients"`
}

type Metadata struct {
	Name             string   `yaml:"name" json:"name"`
	Namespace        string   `yaml:"namespace" json:"namespace"`
	DefaultTaskQueue string   `yaml:"default_task_queue" json:"default_task_queue"`
	Language         []string `yaml:"language" json:"language"`
}

type TypeDef struct {
	Name    string      `yaml:"name" json:"name"`
	Kind    string      `yaml:"kind" json:"kind"`
	AliasOf string      `yaml:"alias_of" json:"alias_of"`
	Fields  []FieldDef  `yaml:"fields" json:"fields"`
	Values  []EnumValue `yaml:"values" json:"values"`
}

type FieldDef struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type" json:"type"`
	Required *bool  `yaml:"required" json:"required"`
}

type EnumValue struct {
	Name  string      `yaml:"name" json:"name"`
	Value interface{} `yaml:"value" json:"value"`
}

type IOContract struct {
	Type string `yaml:"type" json:"type"`
}

type RetryPolicyDef struct {
	Name                   string   `yaml:"name" json:"name"`
	InitialInterval        string   `yaml:"initial_interval" json:"initial_interval"`
	BackoffCoefficient     float64  `yaml:"backoff_coefficient" json:"backoff_coefficient"`
	MaxInterval            string   `yaml:"max_interval" json:"max_interval"`
	MaxAttempts            int      `yaml:"max_attempts" json:"max_attempts"`
	NonRetryableErrorTypes []string `yaml:"non_retryable_error_types" json:"non_retryable_error_types"`
}

type ActivityDef struct {
	Name                   string     `yaml:"name" json:"name"`
	Description            string     `yaml:"description" json:"description"`
	Input                  IOContract `yaml:"input" json:"input"`
	Output                 IOContract `yaml:"output" json:"output"`
	StartToCloseTimeout    string     `yaml:"start_to_close_timeout" json:"start_to_close_timeout"`
	ScheduleToCloseTimeout string     `yaml:"schedule_to_close_timeout" json:"schedule_to_close_timeout"`
	RetryPolicy            string     `yaml:"retry_policy" json:"retry_policy"`
}

type WorkflowDef struct {
	Name             string         `yaml:"name" json:"name"`
	Description      string         `yaml:"description" json:"description"`
	Input            IOContract     `yaml:"input" json:"input"`
	Output           IOContract     `yaml:"output" json:"output"`
	TaskQueue        string         `yaml:"task_queue" json:"task_queue"`
	ExecutionTimeout string         `yaml:"execution_timeout" json:"execution_timeout"`
	RunTimeout       string         `yaml:"run_timeout" json:"run_timeout"`
	TaskTimeout      string         `yaml:"task_timeout" json:"task_timeout"`
	Steps            []WorkflowStep `yaml:"steps" json:"steps"`
}

type WorkflowStep struct {
	ID           string   `yaml:"id" json:"id"`
	Kind         string   `yaml:"kind" json:"kind"`
	Activity     string   `yaml:"activity" json:"activity"`
	InputMapping string   `yaml:"input_mapping" json:"input_mapping"`
	OutputVar    string   `yaml:"output_var" json:"output_var"`
	DependsOn    []string `yaml:"depends_on" json:"depends_on"`
}

type WorkerDef struct {
	Name                       string   `yaml:"name" json:"name"`
	TaskQueue                  string   `yaml:"task_queue" json:"task_queue"`
	Namespace                  string   `yaml:"namespace" json:"namespace"`
	Activities                 []string `yaml:"activities" json:"activities"`
	Workflows                  []string `yaml:"workflows" json:"workflows"`
	GracefulShutdownTimeout    string   `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`
	MaxConcurrentActivities    int      `yaml:"max_concurrent_activities" json:"max_concurrent_activities"`
	MaxConcurrentWorkflowTasks int      `yaml:"max_concurrent_workflow_tasks" json:"max_concurrent_workflow_tasks"`
}

type ClientDef struct {
	Name             string   `yaml:"name" json:"name"`
	Namespace        string   `yaml:"namespace" json:"namespace"`
	Target           string   `yaml:"target" json:"target"`
	AllowedWorkflows []string `yaml:"allowed_workflows" json:"allowed_workflows"`
	RPCTimeout       string   `yaml:"rpc_timeout" json:"rpc_timeout"`
}

type Generator struct {
	schema         *Schema
	typeMap        map[string]TypeDef
	activityMap    map[string]ActivityDef
	workflowMap    map[string]WorkflowDef
	retryPolicyMap map[string]RetryPolicyDef
}

func main() {
	schemaPath := flag.String("schema", "", "Path to schema YAML")
	outDir := flag.String("out", "./generated", "Output directory")
	flag.Parse()
	if *schemaPath == "" {
		fatal(errors.New("-schema is required"))
	}
	schema, err := loadSchema(*schemaPath)
	if err != nil {
		fatal(err)
	}
	gen := newGenerator(schema)
	if err := gen.Generate(*outDir); err != nil {
		fatal(err)
	}
	fmt.Println("generated project in", *outDir)
}

func loadSchema(path string) (*Schema, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Schema
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if len(s.Activities) == 0 || len(s.Workflows) == 0 || len(s.Workers) == 0 || len(s.Clients) == 0 {
		return nil, errors.New("schema must include activities, workflows, workers, and clients")
	}
	return &s, nil
}

func newGenerator(schema *Schema) *Generator {
	g := &Generator{
		schema:         schema,
		typeMap:        map[string]TypeDef{},
		activityMap:    map[string]ActivityDef{},
		workflowMap:    map[string]WorkflowDef{},
		retryPolicyMap: map[string]RetryPolicyDef{},
	}
	for _, t := range schema.Types {
		g.typeMap[t.Name] = t
	}
	for _, a := range schema.Activities {
		g.activityMap[a.Name] = a
	}
	for _, w := range schema.Workflows {
		g.workflowMap[w.Name] = w
	}
	for _, rp := range schema.RetryPolicies {
		g.retryPolicyMap[rp.Name] = rp
	}
	return g
}

func (g *Generator) Generate(outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// Always generate Python
	files := map[string]string{
		"__init__.py":        "",
		"models.py":          g.renderModels(),
		"logging_config.py":  g.renderLoggingConfig(),
		"activities.py":      g.renderActivities(),
		"workflows.py":       g.renderWorkflows(),
		"worker.py":          g.renderWorker(),
		"client.py":          g.renderClient(),
		"requirements.txt":   g.renderRequirements(),
		"Dockerfile":         g.renderDockerfile(),
		"docker-compose.yml": g.renderDockerCompose(),
		".env.example":       g.renderEnvExample(),
		"README.md":          g.renderReadme(),
	}
	for name, content := range files {
		target := filepath.Join(outDir, name)
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return err
		}
	}
	// Generate Java
	if err := g.generateJava(outDir); err != nil {
		return fmt.Errorf("java generation failed: %w", err)
	}
	return nil
}

// ─── Java code generation ───────────────────────────────────────────────────

const javaBasePackage = "com.generated.temporal"

func (g *Generator) generateJava(outDir string) error {
	javaRoot := filepath.Join(outDir, "java-cli")
	srcRoot := filepath.Join(javaRoot, "src", "main", "java", "com", "generated", "temporal")
	dirs := []string{
		filepath.Join(srcRoot, "model"),
		filepath.Join(srcRoot, "activities"),
		filepath.Join(srcRoot, "workflow"),
		filepath.Join(srcRoot, "worker"),
		filepath.Join(srcRoot, "client"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	javaFiles := map[string]string{}

	// Model records / enums
	for _, t := range g.schema.Types {
		switch t.Kind {
		case "enum":
			javaFiles[filepath.Join(srcRoot, "model", t.Name+".java")] = g.renderJavaEnum(t)
		case "struct":
			javaFiles[filepath.Join(srcRoot, "model", t.Name+".java")] = g.renderJavaRecord(t)
		}
	}

	// Activities interface and impl
	javaFiles[filepath.Join(srcRoot, "activities", "GeneratedActivities.java")] = g.renderJavaActivitiesInterface()
	javaFiles[filepath.Join(srcRoot, "activities", "GeneratedActivitiesImpl.java")] = g.renderJavaActivitiesImpl()

	// Workflow interface and impl
	javaFiles[filepath.Join(srcRoot, "workflow", "GeneratedWorkflow.java")] = g.renderJavaWorkflowInterface()
	javaFiles[filepath.Join(srcRoot, "workflow", "GeneratedWorkflowImpl.java")] = g.renderJavaWorkflowImpl()

	// Worker main
	javaFiles[filepath.Join(srcRoot, "worker", "WorkerMain.java")] = g.renderJavaWorkerMain()

	// Client main
	javaFiles[filepath.Join(srcRoot, "client", g.javaClientClassName()+".java")] = g.renderJavaClientMain()

	// pom.xml
	javaFiles[filepath.Join(javaRoot, "pom.xml")] = g.renderPomXml()

	// .gitignore
	javaFiles[filepath.Join(javaRoot, ".gitignore")] = "target/\n"

	// README
	javaFiles[filepath.Join(javaRoot, "README.md")] = g.renderJavaReadme()

	for path, content := range javaFiles {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ─── Java helpers ───────────────────────────────────────────────────────────

func toPascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	var out strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		out.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return out.String()
}

func lowerCamel(pascal string) string {
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func (g *Generator) javaType(in string) string {
	switch strings.ToLower(in) {
	case "string":
		return "String"
	case "int", "integer":
		return "int"
	case "float", "double", "number":
		return "double"
	case "bool", "boolean":
		return "boolean"
	default:
		return in // type names like DocumentIngestInput pass through
	}
}

func (g *Generator) javaBoxedType(in string) string {
	switch strings.ToLower(in) {
	case "int", "integer":
		return "Integer"
	case "float", "double", "number":
		return "Double"
	case "bool", "boolean":
		return "Boolean"
	default:
		return g.javaType(in)
	}
}

func javaPrimitiveExpr(typeName, fieldName string) string {
	switch strings.ToLower(typeName) {
	case "string":
		return `""`
	case "int", "integer":
		return "0"
	case "float", "double", "number":
		return "0.0"
	case "bool", "boolean":
		return "false"
	default:
		return "null"
	}
}

func (g *Generator) javaRecordAccessor(fieldName string) string {
	return lowerCamel(toPascal(fieldName)) + "()"
}

func (g *Generator) javaClientClassName() string {
	wf := g.schema.Workflows[0]
	return toPascal(strings.ReplaceAll(wf.Name, "Workflow", "")) + "ClientMain"
}

// ─── Java enum ──────────────────────────────────────────────────────────────

func (g *Generator) renderJavaEnum(t TypeDef) string {
	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".model;\n\n")
	b.WriteString("public enum " + t.Name + " {\n")
	for i, v := range t.Values {
		b.WriteString("    " + v.Name + "(\"" + fmt.Sprint(v.Value) + "\")")
		if i < len(t.Values)-1 {
			b.WriteString(",\n")
		} else {
			b.WriteString(";\n")
		}
	}
	b.WriteString("\n    private final String value;\n\n")
	b.WriteString("    " + t.Name + "(String value) { this.value = value; }\n\n")
	b.WriteString("    public String value() { return value; }\n\n")
	b.WriteString("    public static " + t.Name + " fromValue(String value) {\n")
	b.WriteString("        for (" + t.Name + " item : values()) {\n")
	b.WriteString("            if (item.value.equalsIgnoreCase(value)) return item;\n")
	b.WriteString("        }\n")
	b.WriteString("        throw new IllegalArgumentException(\"Unsupported value: \" + value);\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

// ─── Java record ────────────────────────────────────────────────────────────

func (g *Generator) renderJavaRecord(t TypeDef) string {
	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".model;\n\n")
	parts := []string{}
	for _, f := range t.Fields {
		parts = append(parts, g.javaType(f.Type)+" "+lowerCamel(toPascal(f.Name)))
	}
	b.WriteString("public record " + t.Name + "(" + strings.Join(parts, ", ") + ") {}\n")
	return b.String()
}

// ─── Java Activities interface ──────────────────────────────────────────────

func (g *Generator) renderJavaActivitiesInterface() string {
	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".activities;\n\n")
	b.WriteString("import io.temporal.activity.ActivityInterface;\n")
	b.WriteString("import io.temporal.activity.ActivityMethod;\n")
	b.WriteString("import " + javaBasePackage + ".model.*;\n\n")
	b.WriteString("@ActivityInterface\n")
	b.WriteString("public interface GeneratedActivities {\n")
	for _, a := range g.schema.Activities {
		methodName := lowerCamel(toPascal(a.Name))
		b.WriteString("    @ActivityMethod(name = \"" + a.Name + "\")\n")
		b.WriteString("    " + g.javaType(a.Output.Type) + " " + methodName + "(" + g.javaType(a.Input.Type) + " input);\n\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// ─── Java Activities impl ───────────────────────────────────────────────────

func (g *Generator) renderJavaActivitiesImpl() string {
	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".activities;\n\n")
	b.WriteString("import org.slf4j.Logger;\n")
	b.WriteString("import org.slf4j.LoggerFactory;\n")
	b.WriteString("import " + javaBasePackage + ".model.*;\n\n")
	b.WriteString("public class GeneratedActivitiesImpl implements GeneratedActivities {\n")
	b.WriteString("    private static final Logger log = LoggerFactory.getLogger(GeneratedActivitiesImpl.class);\n\n")
	for _, a := range g.schema.Activities {
		methodName := lowerCamel(toPascal(a.Name))
		b.WriteString("    @Override\n")
		b.WriteString("    public " + g.javaType(a.Output.Type) + " " + methodName + "(" + g.javaType(a.Input.Type) + " input) {\n")
		b.WriteString("        log.info(\"Running activity: " + a.Name + "\");\n")
		b.WriteString("        // TODO: add real processor implementation here.\n")
		b.WriteString("        return " + g.javaStaticReturnExpr(a.Output.Type) + ";\n")
		b.WriteString("    }\n\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func (g *Generator) javaStaticReturnExpr(typeName string) string {
	t, ok := g.typeMap[typeName]
	if !ok {
		return javaPrimitiveExpr(typeName, "")
	}
	if t.Kind == "enum" {
		if len(t.Values) > 0 {
			return t.Name + "." + t.Values[0].Name
		}
		return "null"
	}
	if t.Kind == "struct" {
		parts := []string{}
		for _, f := range t.Fields {
			parts = append(parts, g.javaStaticFieldValue(f))
		}
		return "new " + typeName + "(" + strings.Join(parts, ", ") + ")"
	}
	return javaPrimitiveExpr(typeName, "")
}

func (g *Generator) javaStaticFieldValue(f FieldDef) string {
	switch strings.ToLower(f.Type) {
	case "string":
		return `"static-` + f.Name + `"`
	case "int", "integer":
		return "1"
	case "float", "double", "number":
		return "1.0"
	case "bool", "boolean":
		return "false"
	}
	td, ok := g.typeMap[f.Type]
	if ok && td.Kind == "enum" && len(td.Values) > 0 {
		return td.Name + "." + td.Values[0].Name
	}
	return "null"
}

// ─── Java Workflow interface ────────────────────────────────────────────────

func (g *Generator) renderJavaWorkflowInterface() string {
	wf := g.schema.Workflows[0]
	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".workflow;\n\n")
	b.WriteString("import io.temporal.workflow.WorkflowInterface;\n")
	b.WriteString("import io.temporal.workflow.WorkflowMethod;\n")
	b.WriteString("import " + javaBasePackage + ".model.*;\n\n")
	b.WriteString("@WorkflowInterface\n")
	b.WriteString("public interface GeneratedWorkflow {\n")
	b.WriteString("    @WorkflowMethod(name = \"" + wf.Name + "\")\n")
	b.WriteString("    " + g.javaType(wf.Output.Type) + " run(" + g.javaType(wf.Input.Type) + " input);\n")
	b.WriteString("}\n")
	return b.String()
}

// ─── Java Workflow impl (with stepVars fix) ─────────────────────────────────

func (g *Generator) renderJavaWorkflowImpl() string {
	wf := g.schema.Workflows[0]
	ordered := topoSteps(wf.Steps)

	// Build stepVars: activity_name -> java variable name
	stepVars := map[string]string{}
	for _, s := range ordered {
		act := g.activityMap[s.Activity]
		resultVar := safeVar(s.OutputVar, s.ID+"Result")
		stepVars[act.Name] = resultVar
	}

	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".workflow;\n\n")
	b.WriteString("import java.time.Duration;\n")
	b.WriteString("import io.temporal.activity.ActivityOptions;\n")
	b.WriteString("import io.temporal.common.RetryOptions;\n")
	b.WriteString("import io.temporal.workflow.Workflow;\n")
	b.WriteString("import " + javaBasePackage + ".activities.GeneratedActivities;\n")
	b.WriteString("import " + javaBasePackage + ".model.*;\n\n")
	b.WriteString("public class GeneratedWorkflowImpl implements GeneratedWorkflow {\n")

	// Activity stub with configurable timeout from first activity (or default)
	timeout := 60
	if len(g.schema.Activities) > 0 && g.schema.Activities[0].StartToCloseTimeout != "" {
		timeout = durationToSeconds(g.schema.Activities[0].StartToCloseTimeout)
	}
	rp := g.retryPolicyMap[g.schema.Activities[0].RetryPolicy]
	maxAttempts := 3
	if rp.MaxAttempts > 0 {
		maxAttempts = rp.MaxAttempts
	}
	b.WriteString("    private final GeneratedActivities activities = Workflow.newActivityStub(\n")
	b.WriteString("        GeneratedActivities.class,\n")
	b.WriteString("        ActivityOptions.newBuilder()\n")
	b.WriteString(fmt.Sprintf("            .setStartToCloseTimeout(Duration.ofSeconds(%d))\n", timeout))
	b.WriteString(fmt.Sprintf("            .setRetryOptions(RetryOptions.newBuilder().setMaximumAttempts(%d).build())\n", maxAttempts))
	b.WriteString("            .build()\n")
	b.WriteString("    );\n\n")

	b.WriteString("    @Override\n")
	b.WriteString("    public " + g.javaType(wf.Output.Type) + " run(" + g.javaType(wf.Input.Type) + " input) {\n")

	// Emit each step calling the activity with proper input construction
	for _, s := range ordered {
		act := g.activityMap[s.Activity]
		resultVar := safeVar(s.OutputVar, s.ID+"Result")
		methodName := lowerCamel(toPascal(act.Name))
		inputExpr := g.javaWorkflowStepInput(act.Input.Type, stepVars)
		b.WriteString("        var " + resultVar + " = activities." + methodName + "(" + inputExpr + ");\n")
	}

	// Build the return expression
	b.WriteString("        return " + g.javaWorkflowOutputExpr(wf.Output.Type, stepVars) + ";\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

// javaWorkflowStepInput builds the constructor call for an activity input type,
// resolving field values from workflow input or prior step output variables.
func (g *Generator) javaWorkflowStepInput(inputType string, stepVars map[string]string) string {
	t, ok := g.typeMap[inputType]
	if !ok || t.Kind != "struct" {
		return "input"
	}
	parts := []string{}
	for _, f := range t.Fields {
		parts = append(parts, g.javaFieldExpr(f.Name, f.Type, stepVars))
	}
	return "new " + inputType + "(" + strings.Join(parts, ", ") + ")"
}

// javaFieldExpr resolves a single field to its Java expression. It checks:
//  1. Fields from the workflow input (e.g. file_path, storage_type)
//  2. Fields from previous step outputs (e.g. file_type from determine_file_type step)
//  3. Falls back to a static primitive expression
func (g *Generator) javaFieldExpr(fieldName, fieldType string, stepVars map[string]string) string {
	// Check if this field comes directly from the workflow input
	wfInput := g.schema.Workflows[0].Input.Type
	if wfInputType, ok := g.typeMap[wfInput]; ok {
		for _, wf := range wfInputType.Fields {
			if wf.Name == fieldName {
				return "input." + g.javaRecordAccessor(fieldName)
			}
		}
	}

	// Check if this field is an output of any previous step's activity
	for _, act := range g.schema.Activities {
		outType, ok := g.typeMap[act.Output.Type]
		if !ok {
			continue
		}
		for _, of := range outType.Fields {
			if of.Name == fieldName {
				if varName, exists := stepVars[act.Name]; exists {
					return varName + "." + g.javaRecordAccessor(fieldName)
				}
			}
		}
	}

	return javaPrimitiveExpr(fieldType, fieldName)
}

// javaWorkflowOutputExpr builds the return constructor for the workflow output type.
func (g *Generator) javaWorkflowOutputExpr(outputType string, stepVars map[string]string) string {
	t, ok := g.typeMap[outputType]
	if !ok || t.Kind != "struct" {
		return g.javaStaticReturnExpr(outputType)
	}
	parts := []string{}
	for _, f := range t.Fields {
		expr := g.javaOutputFieldExpr(f.Name, f.Type, stepVars)
		parts = append(parts, expr)
	}
	return "new " + outputType + "(" + strings.Join(parts, ",\n            ") + "\n        )"
}

// javaOutputFieldExpr resolves a workflow output field. Handles special cases
// like text_preview (substring truncation) generically by checking if the field
// name contains "preview" or "excerpt" and the source type is string.
func (g *Generator) javaOutputFieldExpr(fieldName, fieldType string, stepVars map[string]string) string {
	// Check for text_preview-like fields: truncated string from a prior step
	if strings.Contains(fieldName, "preview") || strings.Contains(fieldName, "excerpt") {
		// Find the source field — look for a "text" output from any activity
		for _, act := range g.schema.Activities {
			outType, ok := g.typeMap[act.Output.Type]
			if !ok {
				continue
			}
			for _, of := range outType.Fields {
				if of.Name == "text" && strings.ToLower(of.Type) == "string" {
					if varName, exists := stepVars[act.Name]; exists {
						accessor := varName + "." + g.javaRecordAccessor("text")
						return accessor + ".substring(0, Math.min(120, " + accessor + ".length()))"
					}
				}
			}
		}
	}

	// Check workflow input fields
	wfInput := g.schema.Workflows[0].Input.Type
	if wfInputType, ok := g.typeMap[wfInput]; ok {
		for _, wf := range wfInputType.Fields {
			if wf.Name == fieldName {
				return "input." + g.javaRecordAccessor(fieldName)
			}
		}
	}

	// Check activity output fields
	for _, act := range g.schema.Activities {
		outType, ok := g.typeMap[act.Output.Type]
		if !ok {
			continue
		}
		for _, of := range outType.Fields {
			if of.Name == fieldName {
				if varName, exists := stepVars[act.Name]; exists {
					return varName + "." + g.javaRecordAccessor(fieldName)
				}
			}
		}
	}

	return javaPrimitiveExpr(fieldType, fieldName)
}

// ─── Java Worker ────────────────────────────────────────────────────────────

func (g *Generator) renderJavaWorkerMain() string {
	worker := g.schema.Workers[0]
	taskQueue := nonEmpty(worker.TaskQueue, g.schema.Metadata.DefaultTaskQueue, "default")

	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".worker;\n\n")
	b.WriteString("import io.temporal.client.WorkflowClient;\n")
	b.WriteString("import io.temporal.serviceclient.WorkflowServiceStubs;\n")
	b.WriteString("import io.temporal.worker.Worker;\n")
	b.WriteString("import io.temporal.worker.WorkerFactory;\n")
	b.WriteString("import " + javaBasePackage + ".activities.GeneratedActivitiesImpl;\n")
	b.WriteString("import " + javaBasePackage + ".workflow.GeneratedWorkflowImpl;\n\n")
	b.WriteString("public class WorkerMain {\n")
	b.WriteString("    public static void main(String[] args) {\n")
	b.WriteString("        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();\n")
	b.WriteString("        WorkflowClient client = WorkflowClient.newInstance(service);\n")
	b.WriteString("        WorkerFactory factory = WorkerFactory.newInstance(client);\n")
	b.WriteString("        Worker worker = factory.newWorker(\"" + taskQueue + "\");\n")
	b.WriteString("        worker.registerWorkflowImplementationTypes(GeneratedWorkflowImpl.class);\n")
	b.WriteString("        worker.registerActivitiesImplementations(new GeneratedActivitiesImpl());\n")
	b.WriteString("        factory.start();\n")
	b.WriteString("        System.out.println(\"Java 21 worker started on task queue: " + taskQueue + "\");\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

// ─── Java Client ────────────────────────────────────────────────────────────

func (g *Generator) renderJavaClientMain() string {
	client := g.schema.Clients[0]
	workflowName := client.AllowedWorkflows[0]
	wf := g.workflowMap[workflowName]
	taskQueue := nonEmpty(wf.TaskQueue, g.schema.Metadata.DefaultTaskQueue, "default")
	className := g.javaClientClassName()

	var b strings.Builder
	b.WriteString("package " + javaBasePackage + ".client;\n\n")
	b.WriteString("import java.util.UUID;\n")
	b.WriteString("import io.temporal.client.WorkflowClient;\n")
	b.WriteString("import io.temporal.client.WorkflowOptions;\n")
	b.WriteString("import io.temporal.serviceclient.WorkflowServiceStubs;\n")
	b.WriteString("import " + javaBasePackage + ".model.*;\n")
	b.WriteString("import " + javaBasePackage + ".workflow.GeneratedWorkflow;\n\n")
	b.WriteString("public class " + className + " {\n")
	b.WriteString("    public static void main(String[] args) {\n")

	// CLI args parsing for workflow input
	b.WriteString(g.javaClientArgsParser(wf.Input.Type))

	b.WriteString("        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();\n")
	b.WriteString("        WorkflowClient client = WorkflowClient.newInstance(service);\n")
	b.WriteString("        WorkflowOptions options = WorkflowOptions.newBuilder()\n")
	b.WriteString("            .setTaskQueue(\"" + taskQueue + "\")\n")
	b.WriteString("            .setWorkflowId(\"" + strings.ToLower(strings.ReplaceAll(workflowName, "Workflow", "")) + "-workflow-\" + UUID.randomUUID())\n")
	b.WriteString("            .build();\n")
	b.WriteString("        GeneratedWorkflow workflow = client.newWorkflowStub(GeneratedWorkflow.class, options);\n")
	b.WriteString("        " + g.javaType(wf.Input.Type) + " input = " + g.javaClientInputExpr(wf.Input.Type) + ";\n")
	b.WriteString("        var result = workflow.run(input);\n")
	b.WriteString("        System.out.println(result);\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

func (g *Generator) javaClientArgsParser(inputType string) string {
	t, ok := g.typeMap[inputType]
	if !ok || t.Kind != "struct" {
		return ""
	}
	var b strings.Builder
	for i, f := range t.Fields {
		accessor := lowerCamel(toPascal(f.Name))
		switch {
		case strings.ToLower(f.Type) == "string":
			b.WriteString(fmt.Sprintf("        String %s = args.length > %d ? args[%d] : \"%s\";\n", accessor, i, i, g.javaCliDefault(f)))
		default:
			td, ok := g.typeMap[f.Type]
			if ok && td.Kind == "enum" {
				b.WriteString(fmt.Sprintf("        String %sValue = args.length > %d ? args[%d] : \"%s\";\n", accessor, i, i, g.javaCliDefault(f)))
			}
		}
	}
	return b.String()
}

func (g *Generator) javaCliDefault(f FieldDef) string {
	switch f.Name {
	case "file_path":
		return "./sample-input/demo.pdf"
	case "storage_type":
		return "s3"
	}
	td, ok := g.typeMap[f.Type]
	if ok && td.Kind == "enum" && len(td.Values) > 0 {
		return fmt.Sprint(td.Values[0].Value)
	}
	return ""
}

func (g *Generator) javaClientInputExpr(inputType string) string {
	t, ok := g.typeMap[inputType]
	if !ok || t.Kind != "struct" {
		return "null"
	}
	parts := []string{}
	for _, f := range t.Fields {
		accessor := lowerCamel(toPascal(f.Name))
		td, ok := g.typeMap[f.Type]
		if ok && td.Kind == "enum" {
			parts = append(parts, td.Name+".fromValue("+accessor+"Value)")
		} else {
			parts = append(parts, accessor)
		}
	}
	return "new " + inputType + "(" + strings.Join(parts, ", ") + ")"
}

// ─── pom.xml ────────────────────────────────────────────────────────────────

func (g *Generator) renderPomXml() string {
	clientClass := javaBasePackage + ".client." + g.javaClientClassName()
	workerClass := javaBasePackage + ".worker.WorkerMain"
	return `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.generated</groupId>
    <artifactId>temporal-generated-java-cli</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <temporal.version>1.27.0</temporal.version>
        <slf4j.version>2.0.16</slf4j.version>
        <jackson.version>2.18.2</jackson.version>
        <mainClass>` + workerClass + `</mainClass>
    </properties>
    <dependencies>
        <dependency>
            <groupId>io.temporal</groupId>
            <artifactId>temporal-sdk</artifactId>
            <version>${temporal.version}</version>
        </dependency>
        <dependency>
            <groupId>org.slf4j</groupId>
            <artifactId>slf4j-simple</artifactId>
            <version>${slf4j.version}</version>
        </dependency>
        <dependency>
            <groupId>com.fasterxml.jackson.core</groupId>
            <artifactId>jackson-databind</artifactId>
            <version>${jackson.version}</version>
        </dependency>
    </dependencies>
    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-compiler-plugin</artifactId>
                <version>3.13.0</version>
                <configuration>
                    <release>21</release>
                </configuration>
            </plugin>
            <plugin>
                <groupId>org.codehaus.mojo</groupId>
                <artifactId>exec-maven-plugin</artifactId>
                <version>3.5.0</version>
            </plugin>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-jar-plugin</artifactId>
                <version>3.4.2</version>
                <configuration>
                    <archive>
                        <manifest>
                            <mainClass>${mainClass}</mainClass>
                        </manifest>
                    </archive>
                </configuration>
            </plugin>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-shade-plugin</artifactId>
                <version>3.6.0</version>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals><goal>shade</goal></goals>
                        <configuration>
                            <transformers>
                                <transformer implementation="org.apache.maven.plugins.shade.resource.ManifestResourceTransformer">
                                    <mainClass>${mainClass}</mainClass>
                                </transformer>
                                <transformer implementation="org.apache.maven.plugins.shade.resource.ServicesResourceTransformer"/>
                            </transformers>
                            <filters>
                                <filter>
                                    <artifact>*:*</artifact>
                                    <excludes>
                                        <exclude>META-INF/*.SF</exclude>
                                        <exclude>META-INF/*.DSA</exclude>
                                        <exclude>META-INF/*.RSA</exclude>
                                    </excludes>
                                </filter>
                            </filters>
                        </configuration>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
    <profiles>
        <profile>
            <id>worker</id>
            <properties>
                <mainClass>` + workerClass + `</mainClass>
            </properties>
        </profile>
        <profile>
            <id>client</id>
            <properties>
                <mainClass>` + clientClass + `</mainClass>
            </properties>
        </profile>
    </profiles>
</project>
`
}

// ─── Java README ────────────────────────────────────────────────────────────

func (g *Generator) renderJavaReadme() string {
	taskQueue := nonEmpty(g.schema.Workers[0].TaskQueue, g.schema.Metadata.DefaultTaskQueue, "default")
	clientClass := javaBasePackage + ".client." + g.javaClientClassName()
	workerClass := javaBasePackage + ".worker.WorkerMain"
	return "# Generated Java 21 Temporal CLI\n\n" +
		"## Build\n" +
		"```bash\nmvn clean package\n```\n\n" +
		"## Run worker\n\n" +
		"Fat JAR (default main class is the worker):\n" +
		"```bash\njava -jar target/temporal-generated-java-cli-1.0.0.jar\n```\n\n" +
		"Or via Maven:\n" +
		"```bash\nmvn exec:java -Dexec.mainClass=" + workerClass + "\n```\n\n" +
		"## Run client\n\n" +
		"Build the client fat JAR:\n" +
		"```bash\nmvn clean package -Pclient\njava -jar target/temporal-generated-java-cli-1.0.0.jar ./sample-input/demo.pdf s3\n```\n\n" +
		"Or via Maven (no rebuild needed):\n" +
		"```bash\nmvn exec:java -Dexec.mainClass=" + clientClass + " -Dexec.args=\"./sample-input/demo.pdf s3\"\n```\n\n" +
		"Or run the client class directly from the worker JAR:\n" +
		"```bash\njava -cp target/temporal-generated-java-cli-1.0.0.jar " + clientClass + " ./sample-input/demo.pdf s3\n```\n\n" +
		"## Maven profiles\n\n" +
		"| Profile   | Main class | Usage |\n" +
		"|-----------|-----------|-------|\n" +
		"| (default) | `" + workerClass + "` | `mvn clean package` |\n" +
		"| worker    | `" + workerClass + "` | `mvn clean package -Pworker` |\n" +
		"| client    | `" + clientClass + "` | `mvn clean package -Pclient` |\n\n" +
		"Task queue: `" + taskQueue + "`\n\n" +
		"The generated code returns static values for each activity. Add your real\n" +
		"processor logic in `GeneratedActivitiesImpl`.\n"
}

func (g *Generator) renderModels() string {
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("from dataclasses import dataclass\n")
	b.WriteString("from enum import Enum\n")
	b.WriteString("from typing import Optional\n\n\n")
	for _, t := range g.schema.Types {
		switch t.Kind {
		case "enum":
			b.WriteString("class " + t.Name + "(str, Enum):\n")
			for _, v := range t.Values {
				b.WriteString("    " + v.Name + ` = "` + fmt.Sprint(v.Value) + `"` + "\n")
			}
			b.WriteString("\n\n")
		case "struct":
			b.WriteString("@dataclass\n")
			b.WriteString("class " + t.Name + ":\n")
			if len(t.Fields) == 0 {
				b.WriteString("    pass\n\n\n")
				continue
			}
			for _, f := range t.Fields {
				ptype := g.pyType(f.Type)
				if f.Required != nil && !*f.Required {
					ptype = "Optional[" + ptype + "]"
				}
				b.WriteString("    " + f.Name + ": " + ptype + "\n")
			}
			b.WriteString("\n\n")
		case "alias":
			b.WriteString(t.Name + " = " + g.pyType(t.AliasOf) + "\n\n")
		}
	}
	return b.String()
}

func (g *Generator) renderLoggingConfig() string {
	return `from __future__ import annotations

import logging
import os
import sys

import structlog


def configure_logging() -> None:
    level_name = os.getenv("LOG_LEVEL", "INFO").upper()
    level = getattr(logging, level_name, logging.INFO)

    timestamper = structlog.processors.TimeStamper(fmt="iso", utc=True)

    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.stdlib.add_log_level,
            structlog.stdlib.add_logger_name,
            timestamper,
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.stdlib.BoundLogger,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )

    logging.basicConfig(
        level=level,
        format="%(message)s",
        stream=sys.stdout,
    )


def get_logger(name: str):
    return structlog.get_logger(name)
`
}

func (g *Generator) renderActivities() string {
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("from pathlib import Path\n\n")
	b.WriteString("from temporalio import activity\n\n")
	b.WriteString("from logging_config import get_logger\n")
	b.WriteString("from models import *\n\n")
	b.WriteString("logger = get_logger(__name__)\n\n\n")
	for _, a := range g.schema.Activities {
		b.WriteString("@activity.defn\n")
		b.WriteString("async def " + a.Name + "(input_data: " + g.pyType(a.Input.Type) + ") -> " + g.pyType(a.Output.Type) + ":\n")
		b.WriteString("    logger.info(\n")
		b.WriteString("        \"activity.start\",\n")
		b.WriteString("        activity=\"" + a.Name + "\",\n")
		b.WriteString("        input_type=type(input_data).__name__,\n")
		b.WriteString("    )\n")
		b.WriteString("    # TODO: Replace this static stub with the real processor implementation.\n")
		b.WriteString("    # Add your file parser, OCR, persistence, or external service code here.\n")
		b.WriteString(g.activityBody(a.Name, a.Output.Type))
		b.WriteString("\n\n")
	}
	return b.String()
}

func (g *Generator) activityBody(name, outType string) string {
	switch outType {
	case "DetermineFileTypeOutput":
		return `    suffix = Path(input_data.file_path).suffix.lower()
    file_type = "text"
    mime_type = "text/plain"
    if suffix == ".pdf":
        file_type = "pdf"
        mime_type = "application/pdf"
    elif suffix in {".png", ".jpg", ".jpeg", ".tif", ".tiff", ".bmp"}:
        file_type = "image"
        mime_type = "image/" + suffix.lstrip(".")
    elif suffix in {".doc", ".docx"}:
        file_type = "word"
        mime_type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    result = DetermineFileTypeOutput(file_type=file_type, mime_type=mime_type)
    logger.info("activity.complete", activity="` + name + `", result=result)
    return result`
	case "ConvertToTextOutput":
		return `    text = f"static extracted text for {input_data.file_path} as {input_data.file_type}"
    result = ConvertToTextOutput(text=text, page_count=1)
    logger.info("activity.complete", activity="` + name + `", page_count=result.page_count)
    return result`
	case "SaveConvertedTextOutput":
		return `    storage = getattr(input_data.storage_type, "value", str(input_data.storage_type))
    if storage == "s3":
        saved_to = "s3://demo-bucket/converted/output.txt"
    elif storage == "database":
        saved_to = "database://document_ingest/records/record-123"
    else:
        saved_to = "/shared/output/converted-output.txt"
    result = SaveConvertedTextOutput(saved_to=saved_to, record_id="record-123")
    logger.info("activity.complete", activity="` + name + `", saved_to=result.saved_to)
    return result`
	}
	return "    return " + g.staticReturnExpr(outType)
}

func (g *Generator) renderWorkflows() string {
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("from datetime import timedelta\n\n")
	b.WriteString("from temporalio import workflow\n")
	b.WriteString("from temporalio.common import RetryPolicy\n\n")
	b.WriteString("with workflow.unsafe.imports_passed_through():\n")
	b.WriteString("    import activities\n")
	b.WriteString("    from logging_config import get_logger\n")
	b.WriteString("    from models import *\n\n")
	b.WriteString("logger = get_logger(__name__)\n\n\n")
	for _, w := range g.schema.Workflows {
		b.WriteString("@workflow.defn\n")
		b.WriteString("class " + w.Name + ":\n")
		b.WriteString("    @workflow.run\n")
		b.WriteString("    async def run(self, input_data: " + g.pyType(w.Input.Type) + ") -> " + g.pyType(w.Output.Type) + ":\n")
		b.WriteString("        logger.info(\"workflow.start\", workflow=\"" + w.Name + "\")\n")
		ordered := topoSteps(w.Steps)
		for _, s := range ordered {
			if s.Kind != "activity" {
				b.WriteString("        # Unsupported step kind: " + s.Kind + "\n")
				continue
			}
			act := g.activityMap[s.Activity]
			rp := g.retryPolicyMap[act.RetryPolicy]
			b.WriteString("        " + safeVar(s.OutputVar, s.ID+"_result") + " = await workflow.execute_activity(\n")
			b.WriteString("            activities." + act.Name + ",\n")
			b.WriteString("            " + g.workflowStepInput(act.Input.Type) + ",\n")
			if act.StartToCloseTimeout != "" {
				b.WriteString("            start_to_close_timeout=timedelta(seconds=" + fmt.Sprint(durationToSeconds(act.StartToCloseTimeout)) + "),\n")
			}
			if act.ScheduleToCloseTimeout != "" {
				b.WriteString("            schedule_to_close_timeout=timedelta(seconds=" + fmt.Sprint(durationToSeconds(act.ScheduleToCloseTimeout)) + "),\n")
			}
			b.WriteString("            retry_policy=" + g.retryPolicyExpr(rp) + ",\n")
			b.WriteString("        )\n\n")
		}
		b.WriteString("        result = " + g.workflowOutputExpr(w.Output.Type) + "\n")
		b.WriteString("        logger.info(\"workflow.complete\", workflow=\"" + w.Name + "\")\n")
		b.WriteString("        return result\n\n\n")
	}
	return b.String()
}

func (g *Generator) retryPolicyExpr(rp RetryPolicyDef) string {
	if rp.Name == "" {
		return "RetryPolicy(maximum_attempts=3)"
	}
	parts := []string{}
	if rp.InitialInterval != "" {
		parts = append(parts, "initial_interval=timedelta(seconds="+fmt.Sprint(durationToSeconds(rp.InitialInterval))+")")
	}
	if rp.BackoffCoefficient > 0 {
		parts = append(parts, fmt.Sprintf("backoff_coefficient=%v", rp.BackoffCoefficient))
	}
	if rp.MaxInterval != "" {
		parts = append(parts, "maximum_interval=timedelta(seconds="+fmt.Sprint(durationToSeconds(rp.MaxInterval))+")")
	}
	if rp.MaxAttempts > 0 {
		parts = append(parts, fmt.Sprintf("maximum_attempts=%d", rp.MaxAttempts))
	}
	if len(rp.NonRetryableErrorTypes) > 0 {
		quoted := []string{}
		for _, v := range rp.NonRetryableErrorTypes {
			quoted = append(quoted, fmt.Sprintf("%q", v))
		}
		parts = append(parts, "non_retryable_error_types=["+strings.Join(quoted, ", ")+"]")
	}
	if len(parts) == 0 {
		return "RetryPolicy(maximum_attempts=3)"
	}
	return "RetryPolicy(" + strings.Join(parts, ", ") + ")"
}

func (g *Generator) renderWorker() string {
	worker := g.schema.Workers[0]
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("import asyncio\n")
	b.WriteString("import os\n\n")
	b.WriteString("from dotenv import load_dotenv\n")
	b.WriteString("from temporalio.client import Client\n")
	b.WriteString("from temporalio.worker import Worker\n\n")
	b.WriteString("import activities\n")
	b.WriteString("from logging_config import configure_logging, get_logger\n")
	b.WriteString("from workflows import *\n\n")
	b.WriteString("load_dotenv()\n")
	b.WriteString("configure_logging()\n")
	b.WriteString("logger = get_logger(__name__)\n\n\n")
	b.WriteString("async def main() -> None:\n")
	b.WriteString("    target = os.getenv(\"TEMPORAL_TARGET\", \"localhost:7233\")\n")
	b.WriteString("    namespace = os.getenv(\"TEMPORAL_NAMESPACE\", \"" + nonEmpty(worker.Namespace, g.schema.Metadata.Namespace, "default") + "\")\n")
	b.WriteString("    client = await Client.connect(target, namespace=namespace)\n")
	b.WriteString("    logger.info(\"worker.start\", task_queue=\"" + worker.TaskQueue + "\", target=target, namespace=namespace)\n")
	b.WriteString("    worker = Worker(\n")
	b.WriteString("        client,\n")
	b.WriteString("        task_queue=\"" + worker.TaskQueue + "\",\n")
	b.WriteString("        workflows=[\n")
	for _, wf := range worker.Workflows {
		b.WriteString("            " + wf + ",\n")
	}
	b.WriteString("        ],\n")
	b.WriteString("        activities=[\n")
	for _, act := range worker.Activities {
		b.WriteString("            activities." + act + ",\n")
	}
	b.WriteString("        ],\n")
	if worker.MaxConcurrentActivities > 0 {
		b.WriteString(fmt.Sprintf("        max_concurrent_activities=%d,\n", worker.MaxConcurrentActivities))
	}
	if worker.MaxConcurrentWorkflowTasks > 0 {
		b.WriteString(fmt.Sprintf("        max_concurrent_workflow_tasks=%d,\n", worker.MaxConcurrentWorkflowTasks))
	}
	b.WriteString("    )\n")
	b.WriteString("    await worker.run()\n\n\n")
	b.WriteString("if __name__ == \"__main__\":\n")
	b.WriteString("    asyncio.run(main())\n")
	return b.String()
}

func (g *Generator) renderClient() string {
	client := g.schema.Clients[0]
	workflowName := client.AllowedWorkflows[0]
	wf := g.workflowMap[workflowName]
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("import asyncio\n")
	b.WriteString("import os\n\n")
	b.WriteString("from dotenv import load_dotenv\n")
	b.WriteString("from temporalio.client import Client\n\n")
	b.WriteString("from logging_config import configure_logging, get_logger\n")
	b.WriteString("from models import *\n")
	b.WriteString("from workflows import " + workflowName + "\n\n")
	b.WriteString("load_dotenv()\n")
	b.WriteString("configure_logging()\n")
	b.WriteString("logger = get_logger(__name__)\n\n\n")
	b.WriteString("async def main() -> None:\n")
	b.WriteString("    target = os.getenv(\"TEMPORAL_TARGET\", \"" + client.Target + "\")\n")
	b.WriteString("    namespace = os.getenv(\"TEMPORAL_NAMESPACE\", \"" + nonEmpty(client.Namespace, g.schema.Metadata.Namespace, "default") + "\")\n")
	b.WriteString("    client = await Client.connect(target, namespace=namespace)\n")
	b.WriteString("    request = " + g.sampleValueExpr(wf.Input.Type) + "\n")
	b.WriteString("    handle = await client.start_workflow(\n")
	b.WriteString("        " + workflowName + ".run,\n")
	b.WriteString("        request,\n")
	b.WriteString("        id=\"document-ingest-sample\",\n")
	b.WriteString("        task_queue=\"" + nonEmpty(wf.TaskQueue, g.schema.Metadata.DefaultTaskQueue, "default") + "\",\n")
	b.WriteString("    )\n")
	b.WriteString("    logger.info(\"client.started\", workflow_id=handle.id, run_id=handle.result_run_id)\n")
	b.WriteString("    result = await handle.result()\n")
	b.WriteString("    logger.info(\"client.result\", result=result)\n")
	b.WriteString("    print(result)\n\n\n")
	b.WriteString("if __name__ == \"__main__\":\n")
	b.WriteString("    asyncio.run(main())\n")
	return b.String()
}

func (g *Generator) renderRequirements() string {
	deps := g.collectRequirements()
	return strings.Join(deps, "\n") + "\n"
}

func (g *Generator) collectRequirements() []string {
	deps := []string{
		"temporalio==1.24.0",
		"structlog==25.5.0",
		"python-dotenv==1.2.2",
	}

	needMagic := false
	needPillow := false
	needTesseract := false
	needPypdf := false
	needBoto3 := false
	needSQLA := false

	for _, a := range g.schema.Activities {
		text := strings.ToLower(a.Name + " " + a.Description)
		if strings.Contains(text, "mime") || strings.Contains(text, "file type") || strings.Contains(text, "determine_file_type") {
			needMagic = true
		}
		if strings.Contains(text, "ocr") || strings.Contains(text, "image") {
			needPillow = true
			needTesseract = true
		}
		if strings.Contains(text, "pdf") || strings.Contains(text, "convert") || strings.Contains(text, "text") {
			needPypdf = true
		}
		if strings.Contains(text, "s3") {
			needBoto3 = true
		}
		if strings.Contains(text, "database") || strings.Contains(text, "sql") {
			needSQLA = true
		}
	}
	for _, t := range g.schema.Types {
		if strings.Contains(strings.ToLower(t.Name), "storage") {
			for _, v := range t.Values {
				lower := strings.ToLower(fmt.Sprint(v.Value) + " " + v.Name)
				if strings.Contains(lower, "s3") {
					needBoto3 = true
				}
				if strings.Contains(lower, "database") || strings.Contains(lower, "db") {
					needSQLA = true
				}
			}
		}
	}

	if needMagic {
		deps = append(deps, "python-magic>=0.4.27")
	}
	if needPypdf {
		deps = append(deps, "pypdf>=4.0.0")
	}
	if needPillow {
		deps = append(deps, "Pillow>=10.0.0")
	}
	if needTesseract {
		deps = append(deps, "pytesseract>=0.3.10")
	}
	if needBoto3 {
		deps = append(deps, "boto3==1.42.76")
	}
	if needSQLA {
		deps = append(deps, "SQLAlchemy>=2.0.0")
	}

	return deps
}

func (g *Generator) renderDockerfile() string {
	return `FROM python:3.13-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libmagic1 \
    tesseract-ocr \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "worker.py"]
`
}

func (g *Generator) renderDockerCompose() string {
	worker := g.schema.Workers[0]
	return `version: "3.9"

services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    ports:
      - "5432:5432"
    volumes:
      - temporal-postgres:/var/lib/postgresql/data

  temporal:
    image: temporalio/auto-setup:1.28.2
    depends_on:
      - postgres
    environment:
      DB: postgres12
      DB_PORT: 5432
      POSTGRES_USER: temporal
      POSTGRES_PWD: temporal
      POSTGRES_SEEDS: postgres
      DYNAMIC_CONFIG_FILE_PATH: config/dynamicconfig/development-sql.yaml
    ports:
      - "7233:7233"
      - "8233:8233"

  worker:
    build: .
    depends_on:
      - temporal
    environment:
      TEMPORAL_TARGET: temporal:7233
      TEMPORAL_NAMESPACE: ` + nonEmpty(worker.Namespace, g.schema.Metadata.Namespace, "default") + `
      LOG_LEVEL: INFO
    volumes:
      - ./shared:/shared
    command: ["python", "worker.py"]

  client:
    build: .
    depends_on:
      - temporal
      - worker
    environment:
      TEMPORAL_TARGET: temporal:7233
      TEMPORAL_NAMESPACE: ` + nonEmpty(worker.Namespace, g.schema.Metadata.Namespace, "default") + `
      LOG_LEVEL: INFO
    volumes:
      - ./shared:/shared
    command: ["python", "client.py"]

  java-worker:
    build:
      context: ./java-cli
      dockerfile: Dockerfile
    depends_on:
      - temporal
    environment:
      TEMPORAL_TARGET: temporal:7233
    command: ["mvn", "exec:java", "-Dexec.mainClass=` + javaBasePackage + `.worker.WorkerMain"]

  java-client:
    build:
      context: ./java-cli
      dockerfile: Dockerfile
    depends_on:
      - temporal
      - java-worker
    command: ["mvn", "exec:java", "-Dexec.mainClass=` + javaBasePackage + `.client.` + g.javaClientClassName() + `"]

volumes:
  temporal-postgres:
`
}

func (g *Generator) renderEnvExample() string {
	ns := nonEmpty(g.schema.Metadata.Namespace, "default")
	return "TEMPORAL_TARGET=localhost:7233\nTEMPORAL_NAMESPACE=" + ns + "\nLOG_LEVEL=INFO\n"
}

func (g *Generator) renderReadme() string {
	workflowName := g.schema.Workflows[0].Name
	return "# Generated Temporal Python Project\n\n" +
		"This folder was generated from the schema and includes:\n" +
		"- Python Temporal workflow stubs\n" +
		"- activity stubs with structured logging\n" +
		"- retry policies and activity timeouts generated from the schema\n" +
		"- dynamic requirements.txt generation based on schema features\n" +
		"- Dockerfile and docker-compose.yml for local development\n\n" +
		"## Files\n" +
		"- models.py\n" +
		"- logging_config.py\n" +
		"- activities.py\n" +
		"- workflows.py\n" +
		"- worker.py\n" +
		"- client.py\n" +
		"- requirements.txt\n" +
		"- Dockerfile\n" +
		"- docker-compose.yml\n\n" +
		"## Local run\n\n" +
		"    python3 -m venv .venv\n" +
		"    source .venv/bin/activate\n" +
		"    pip install -r requirements.txt\n" +
		"    python worker.py\n\n" +
		"In another terminal:\n\n" +
		"    python client.py\n\n" +
		"## Docker run\n\n" +
		"    docker compose up --build\n\n" +
		"Temporal gRPC: localhost:7233\n" +
		"Temporal Web UI: http://localhost:8233\n\n" +
		"## Generated workflow\n" +
		"The sample client starts " + workflowName + " with static sample input.\n\n" +
		"## Notes\n" +
		"Each activity contains a TODO comment showing where to add your real processor implementation.\n"
}

func (g *Generator) pyType(in string) string {
	switch strings.ToLower(in) {
	case "string":
		return "str"
	case "int", "integer":
		return "int"
	case "float", "double", "number":
		return "float"
	case "bool", "boolean":
		return "bool"
	default:
		return in
	}
}

func (g *Generator) staticReturnExpr(typeName string) string {
	switch typeName {
	case "str":
		return `"static-string"`
	case "int":
		return "1"
	case "float":
		return "1.0"
	case "bool":
		return "True"
	}
	td, ok := g.typeMap[typeName]
	if !ok {
		return "None"
	}
	if td.Kind == "enum" {
		if len(td.Values) > 0 {
			return td.Name + "." + td.Values[0].Name
		}
		return "None"
	}
	if td.Kind == "struct" {
		parts := []string{}
		for _, f := range td.Fields {
			parts = append(parts, f.Name+"="+g.staticReturnExpr(g.pyType(f.Type)))
		}
		return td.Name + "(" + strings.Join(parts, ", ") + ")"
	}
	return "None"
}

func (g *Generator) workflowStepInput(typeName string) string {
	switch typeName {
	case "DetermineFileTypeInput":
		return "DetermineFileTypeInput(file_path=input_data.file_path)"
	case "ConvertToTextInput":
		return "ConvertToTextInput(file_path=input_data.file_path, file_type=file_type_result.file_type)"
	case "SaveConvertedTextInput":
		return "SaveConvertedTextInput(file_path=input_data.file_path, storage_type=input_data.storage_type, text=convert_result.text)"
	default:
		return "input_data"
	}
}

func (g *Generator) workflowOutputExpr(typeName string) string {
	switch typeName {
	case "DocumentIngestResult":
		return "DocumentIngestResult(file_type=file_type_result.file_type, mime_type=file_type_result.mime_type, text_preview=convert_result.text[:120], saved_to=save_result.saved_to, record_id=save_result.record_id)"
	default:
		return g.staticReturnExpr(typeName)
	}
}

func (g *Generator) sampleValueExpr(typeName string) string {
	switch typeName {
	case "DocumentIngestInput":
		return `DocumentIngestInput(file_path="/data/sample.pdf", storage_type=FileStorageType.S3)`
	}
	return g.staticReturnExpr(typeName)
}

func topoSteps(steps []WorkflowStep) []WorkflowStep {
	byID := map[string]WorkflowStep{}
	indegree := map[string]int{}
	edges := map[string][]string{}
	for _, s := range steps {
		byID[s.ID] = s
		if _, ok := indegree[s.ID]; !ok {
			indegree[s.ID] = 0
		}
	}
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			edges[dep] = append(edges[dep], s.ID)
			indegree[s.ID]++
		}
	}
	queue := []string{}
	for id, d := range indegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)
	out := []WorkflowStep{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		out = append(out, byID[id])
		for _, nxt := range edges[id] {
			indegree[nxt]--
			if indegree[nxt] == 0 {
				queue = append(queue, nxt)
				sort.Strings(queue)
			}
		}
	}
	if len(out) != len(steps) {
		return steps
	}
	return out
}

func durationToSeconds(s string) int {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 60
	}
	re := regexp.MustCompile(`^(\d+)(s|m|h)$`)
	m := re.FindStringSubmatch(s)
	if len(m) != 3 {
		return 60
	}
	n := atoi(m[1])
	switch m[2] {
	case "s":
		return n
	case "m":
		return n * 60
	case "h":
		return n * 3600
	default:
		return 60
	}
}

func atoi(s string) int {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

func safeVar(name, fallback string) string {
	if strings.TrimSpace(name) == "" {
		return fallback
	}
	return name
}

func nonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
