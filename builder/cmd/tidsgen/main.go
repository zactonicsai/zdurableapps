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
	Metadata   Metadata      `yaml:"metadata" json:"metadata"`
	Types      []TypeDef     `yaml:"types" json:"types"`
	Activities []ActivityDef `yaml:"activities" json:"activities"`
	Workflows  []WorkflowDef `yaml:"workflows" json:"workflows"`
	Workers    []WorkerDef   `yaml:"workers" json:"workers"`
	Clients    []ClientDef   `yaml:"clients" json:"clients"`
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
	Value interface{} `yaml:"value"`
}

type IOContract struct {
	Type string `yaml:"type" json:"type"`
}

type ActivityDef struct {
	Name                string     `yaml:"name" json:"name"`
	Description         string     `yaml:"description" json:"description"`
	Input               IOContract `yaml:"input" json:"input"`
	Output              IOContract `yaml:"output" json:"output"`
	StartToCloseTimeout string     `yaml:"start_to_close_timeout" json:"start_to_close_timeout"`
}

type WorkflowDef struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Input       IOContract     `yaml:"input" json:"input"`
	Output      IOContract     `yaml:"output" json:"output"`
	TaskQueue   string         `yaml:"task_queue" json:"task_queue"`
	Steps       []WorkflowStep `yaml:"steps" json:"steps"`
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
	TaskQueue  string   `yaml:"task_queue" json:"task_queue"`
	Namespace  string   `yaml:"namespace" json:"namespace"`
	Activities []string `yaml:"activities" json:"activities"`
	Workflows  []string `yaml:"workflows" json:"workflows"`
}

type ClientDef struct {
	Namespace        string   `yaml:"namespace" json:"namespace"`
	Target           string   `yaml:"target" json:"target"`
	AllowedWorkflows []string `yaml:"allowed_workflows" json:"allowed_workflows"`
}

type Generator struct {
	schema      *Schema
	typeMap     map[string]TypeDef
	activityMap map[string]ActivityDef
	workflowMap map[string]WorkflowDef
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
	g := &Generator{schema: schema, typeMap: map[string]TypeDef{}, activityMap: map[string]ActivityDef{}, workflowMap: map[string]WorkflowDef{}}
	for _, t := range schema.Types {
		g.typeMap[t.Name] = t
	}
	for _, a := range schema.Activities {
		g.activityMap[a.Name] = a
	}
	for _, w := range schema.Workflows {
		g.workflowMap[w.Name] = w
	}
	return g
}

func (g *Generator) Generate(outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	files := map[string]string{
		"models.py":        g.renderModels(),
		"activities.py":    g.renderActivities(),
		"workflows.py":     g.renderWorkflows(),
		"worker.py":        g.renderWorker(),
		"client.py":        g.renderClient(),
		"requirements.txt": "temporalio==1.24.0\n",
		"README.md":        g.renderReadme(),
	}
	for name, content := range files {
		target := filepath.Join(outDir, name)
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
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

func (g *Generator) renderActivities() string {
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("from temporalio import activity\n\n")
	b.WriteString("from models import *\n\n\n")
	for _, a := range g.schema.Activities {
		b.WriteString("@activity.defn\n")
		b.WriteString("async def " + a.Name + "(input_data: " + g.pyType(a.Input.Type) + ") -> " + g.pyType(a.Output.Type) + ":\n")
		b.WriteString("    # TODO: Replace the static return below with the real processor implementation.\n")
		b.WriteString("    # Example real logic: sniff MIME type, convert to text, save to DB/shared FS/S3, etc.\n")
		b.WriteString(`    activity.logger.info("running activity", extra={"activity": "` + a.Name + `"})` + "\n")
		b.WriteString("    return " + g.staticReturnExpr(a.Output.Type) + "\n\n\n")
	}
	return b.String()
}

func (g *Generator) renderWorkflows() string {
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("from datetime import timedelta\n\n")
	b.WriteString("from temporalio import workflow\n")
	b.WriteString("from temporalio.common import RetryPolicy\n\n")
	b.WriteString("with workflow.unsafe.imports_passed_through():\n")
	b.WriteString("    import activities\n")
	b.WriteString("    from models import *\n\n\n")
	for _, w := range g.schema.Workflows {
		b.WriteString("@workflow.defn\n")
		b.WriteString("class " + w.Name + ":\n")
		b.WriteString("    @workflow.run\n")
		b.WriteString("    async def run(self, input_data: " + g.pyType(w.Input.Type) + ") -> " + g.pyType(w.Output.Type) + ":\n")
		ordered := topoSteps(w.Steps)
		for _, s := range ordered {
			if s.Kind != "activity" {
				b.WriteString("        # Unsupported step kind: " + s.Kind + "\n")
				continue
			}
			act := g.activityMap[s.Activity]
			b.WriteString("        " + safeVar(s.OutputVar, s.ID+"_result") + " = await workflow.execute_activity(\n")
			b.WriteString("            activities." + act.Name + ",\n")
			b.WriteString("            " + g.workflowStepInput(act.Input.Type) + ",\n")
			b.WriteString("            start_to_close_timeout=timedelta(seconds=" + fmt.Sprint(durationToSeconds(act.StartToCloseTimeout)) + "),\n")
			b.WriteString("            retry_policy=RetryPolicy(maximum_attempts=3),\n")
			b.WriteString("        )\n\n")
		}
		b.WriteString("        return " + g.workflowOutputExpr(w.Output.Type) + "\n\n\n")
	}
	return b.String()
}

func (g *Generator) renderWorker() string {
	worker := g.schema.Workers[0]
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("import asyncio\n\n")
	b.WriteString("from temporalio.client import Client\n")
	b.WriteString("from temporalio.worker import Worker\n\n")
	b.WriteString("import activities\n")
	b.WriteString("from workflows import *\n\n\n")
	b.WriteString("async def main() -> None:\n")
	b.WriteString("    client = await Client.connect(\"localhost:7233\", namespace=\"" + defaultString(worker.Namespace, g.schema.Metadata.Namespace) + "\")\n")
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
	b.WriteString("    )\n")
	b.WriteString("    await worker.run()\n\n\n")
	b.WriteString("if __name__ == \"__main__\":\n")
	b.WriteString("    asyncio.run(main())\n")
	return b.String()
}

func (g *Generator) renderClient() string {
	client := g.schema.Clients[0]
	wf := g.workflowMap[client.AllowedWorkflows[0]]
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("import asyncio\n")
	b.WriteString("import sys\n")
	b.WriteString("from uuid import uuid4\n\n")
	b.WriteString("from temporalio.client import Client\n\n")
	b.WriteString("from models import *\n")
	b.WriteString("from workflows import " + wf.Name + "\n\n\n")
	b.WriteString("def build_input() -> " + wf.Input.Type + ":\n")
	b.WriteString("    file_path = sys.argv[1] if len(sys.argv) > 1 else \"./sample-input/demo.pdf\"\n")
	b.WriteString("    storage_arg = sys.argv[2] if len(sys.argv) > 2 else \"s3\"\n")
	b.WriteString("    return " + wf.Input.Type + "(file_path=file_path, storage_type=FileStorageType(storage_arg))\n\n\n")
	b.WriteString("async def main() -> None:\n")
	b.WriteString("    client = await Client.connect(\"" + client.Target + "\", namespace=\"" + defaultString(client.Namespace, g.schema.Metadata.Namespace) + "\")\n")
	b.WriteString("    result = await client.execute_workflow(\n")
	b.WriteString("        " + wf.Name + ".run,\n")
	b.WriteString("        build_input(),\n")
	b.WriteString("        id=f\"" + toKebab(wf.Name) + "-{uuid4()}\",\n")
	b.WriteString("        task_queue=\"" + defaultString(wf.TaskQueue, g.schema.Metadata.DefaultTaskQueue) + "\",\n")
	b.WriteString("    )\n")
	b.WriteString("    print(result)\n\n\n")
	b.WriteString("if __name__ == \"__main__\":\n")
	b.WriteString("    asyncio.run(main())\n")
	return b.String()
}

func (g *Generator) renderReadme() string {
	return "# Generated Temporal Python project\n\npython -m venv .venv\nsource .venv/bin/activate\npip install -r requirements.txt\npython worker.py\npython client.py ./sample-input/demo.pdf s3\n"
}

func (g *Generator) workflowStepInput(inputType string) string {
	t, ok := g.typeMap[inputType]
	if !ok || t.Kind != "struct" {
		return "input_data"
	}
	parts := []string{}
	for _, f := range t.Fields {
		parts = append(parts, f.Name+"="+g.fieldExpr(f.Name, f.Type))
	}
	return inputType + "(" + strings.Join(parts, ", ") + ")"
}

func (g *Generator) fieldExpr(fieldName, fieldType string) string {
	switch fieldName {
	case "file_path":
		return "input_data.file_path"
	case "storage_type":
		return "input_data.storage_type"
	case "file_type":
		return "file_type_result.file_type"
	case "mime_type":
		return "file_type_result.mime_type"
	case "text":
		return "convert_result.text"
	case "page_count":
		return "convert_result.page_count"
	default:
		return primitiveExpr(fieldType, fieldName)
	}
}

func (g *Generator) workflowOutputExpr(outputType string) string {
	t, ok := g.typeMap[outputType]
	if !ok || t.Kind != "struct" {
		return g.staticReturnExpr(outputType)
	}
	parts := []string{}
	for _, f := range t.Fields {
		switch f.Name {
		case "file_type":
			parts = append(parts, "file_type=file_type_result.file_type")
		case "mime_type":
			parts = append(parts, "mime_type=file_type_result.mime_type")
		case "text_preview":
			parts = append(parts, "text_preview=convert_result.text[:120]")
		case "saved_to":
			parts = append(parts, "saved_to=save_result.saved_to")
		case "record_id":
			parts = append(parts, "record_id=save_result.record_id")
		default:
			parts = append(parts, f.Name+"="+primitiveExpr(f.Type, f.Name))
		}
	}
	return outputType + "(" + strings.Join(parts, ", ") + ")"
}

func (g *Generator) staticReturnExpr(typeName string) string {
	if t, ok := g.typeMap[typeName]; ok {
		switch t.Kind {
		case "enum":
			if len(t.Values) > 0 {
				return typeName + "." + t.Values[0].Name
			}
		case "struct":
			parts := []string{}
			for _, f := range t.Fields {
				parts = append(parts, f.Name+"="+g.staticFieldExpr(f.Name, f.Type))
			}
			return typeName + "(" + strings.Join(parts, ", ") + ")"
		}
	}
	return primitiveExpr(typeName, "value")
}

func (g *Generator) staticFieldExpr(fieldName, fieldType string) string {
	switch fieldName {
	case "file_type":
		return "\"pdf\""
	case "mime_type":
		return "\"application/pdf\""
	case "text":
		return "\"static extracted text\""
	case "page_count":
		return "1"
	case "saved_to":
		return "\"s3://demo-bucket/static-output.txt\""
	case "record_id":
		return "\"record-123\""
	case "text_preview":
		return "\"static extracted text\""
	}
	if t, ok := g.typeMap[fieldType]; ok && t.Kind == "enum" && len(t.Values) > 0 {
		return fieldType + "." + t.Values[0].Name
	}
	return primitiveExpr(fieldType, fieldName)
}

func (g *Generator) pyType(typeName string) string {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "string":
		return "str"
	case "int", "integer":
		return "int"
	case "float", "double", "number":
		return "float"
	case "bool", "boolean":
		return "bool"
	default:
		return typeName
	}
}

func primitiveExpr(typeName, context string) string {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "string", "str":
		return "\"static-" + context + "\""
	case "int", "integer":
		return "1"
	case "float", "double", "number":
		return "1.0"
	case "bool", "boolean":
		return "True"
	default:
		return "\"static-" + context + "\""
	}
}

func topoSteps(steps []WorkflowStep) []WorkflowStep {
	byID := map[string]WorkflowStep{}
	for _, s := range steps {
		byID[s.ID] = s
	}
	visited := map[string]bool{}
	out := []WorkflowStep{}
	var visit func(string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		s := byID[id]
		for _, dep := range s.DependsOn {
			visit(dep)
		}
		visited[id] = true
		out = append(out, s)
	}
	ids := make([]string, 0, len(steps))
	for _, s := range steps {
		ids = append(ids, s.ID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		visit(id)
	}
	return out
}

func durationToSeconds(v string) int {
	re := regexp.MustCompile(`^(\d+)(ms|s|m|h)$`)
	m := re.FindStringSubmatch(strings.TrimSpace(v))
	if len(m) != 3 {
		return 30
	}
	n := 0
	for _, r := range m[1] {
		n = n*10 + int(r-'0')
	}
	switch m[2] {
	case "ms":
		if n < 1000 {
			return 1
		}
		return n / 1000
	case "s":
		return n
	case "m":
		return n * 60
	case "h":
		return n * 3600
	}
	return 30
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func safeVar(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		v = fallback
	}
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, ".", "_")
	return v
}

func toKebab(v string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	return strings.ToLower(re.ReplaceAllString(v, `${1}-${2}`))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
