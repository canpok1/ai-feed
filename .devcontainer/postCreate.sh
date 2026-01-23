#!/bin/sh
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLAUDE_DIR=/home/vscode/.claude

ln -s ${SCRIPT_DIR}/.claude ${CLAUDE_DIR}

npm install -g @google/gemini-cli
curl -fsSL https://claude.ai/install.sh | bash

cd "$(dirname "$0")/.."
make setup
