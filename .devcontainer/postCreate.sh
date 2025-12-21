#!/bin/sh
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLAUDE_DIR=/home/vscode/.claude

ln -s ${SCRIPT_DIR}/.claude ${CLAUDE_DIR}

npm install -g @google/gemini-cli
npm install -g @anthropic-ai/claude-code

cd "$(dirname "$0")/.."
make setup
