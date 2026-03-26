# TIDS Temporal Python Code Generator

This project contains:

- a Go generator that reads a **TIDS** YAML file and generates Temporal Python SDK stubs
- a sample **document ingest** schema
- generated Python example files that run against a Temporal server on `localhost:7233`

## What it generates

For a given schema, the generator emits:

- `models.py` — enums and dataclasses from `types`
- `activities.py` — activity stubs with static return values
- `workflows.py` — Temporal workflows that call the generated activities
- `worker.py` — Temporal worker bootstrap
- `client.py` — test client for starting workflows
- `requirements.txt` — Python dependencies

The generated activity methods include comments showing where to add real processors.

## Current generator scope

This generator focuses on the core path needed to bootstrap a real project quickly:

- `types.kind`: `struct`, `enum`, `alias`
- `activities`
- `workflows`
- `workers`
- `clients`
- workflow steps with `kind: activity`
- simple `input_mapping` values like:
  - `{{input}}`
  - `{{input.field_name}}`
  - `{{step_var}}`
  - `{{step_var.field_name}}`

It emits static values for activity outputs based on the declared output type.

## Build the Go generator

```bash
go build -o tidsgen ./cmd/tidsgen
```

## Generate code from the sample schema

```bash
./tidsgen \
  -schema ./examples/document_ingest_schema.yaml \
  -out ./generated/document_ingest
```

## Run the generated worker

Start Temporal locally first:

```bash
docker run --rm -d --name temporal-dev \
  -p 7233:7233 \
  temporalio/auto-setup:latest
```

Then:

```bash
cd ./generated/document_ingest
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python worker.py
```

## Run the generated client

```bash
cd ./generated/document_ingest
source .venv/bin/activate
python client.py ./sample-input/demo.pdf s3
```

## Note

The uploaded file is a **schema specification**, not a concrete executable instance. The included `examples/document_ingest_schema.yaml` is a concrete instance derived from that spec.
