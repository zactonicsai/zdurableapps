# Generated Temporal Python project

python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python worker.py
python client.py ./sample-input/demo.pdf s3
