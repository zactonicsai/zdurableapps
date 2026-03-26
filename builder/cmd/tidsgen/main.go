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
