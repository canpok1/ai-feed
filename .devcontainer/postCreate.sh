#!/bin/sh
npm install -g @google/gemini-cli
go install github.com/goreleaser/goreleaser/v2@latest
npm install -g @anthropic-ai/claude-code

cd "$(dirname "$0")/.."
make setup

claude config set --global preferredNotifChannel terminal_bell
