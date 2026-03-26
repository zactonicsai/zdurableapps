# Generated Temporal Python Project

This folder was generated from the schema and includes:
- Python Temporal workflow stubs
- activity stubs with structured logging
- retry policies and activity timeouts generated from the schema
- dynamic requirements.txt generation based on schema features
- Dockerfile and docker-compose.yml for local development

## Files
- models.py
- logging_config.py
- activities.py
- workflows.py
- worker.py
- client.py
- requirements.txt
- Dockerfile
- docker-compose.yml

## Local run

    python3 -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt
    python worker.py

In another terminal:

    python client.py

## Docker run

    docker compose up --build

Temporal gRPC: localhost:7233
Temporal Web UI: http://localhost:8233

## Generated workflow
The sample client starts DocumentIngestWorkflow with static sample input.

## Notes
Each activity contains a TODO comment showing where to add your real processor implementation.
