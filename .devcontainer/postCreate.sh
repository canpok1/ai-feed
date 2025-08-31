#!/bin/sh
npm install -g @google/gemini-cli
go install github.com/goreleaser/goreleaser/v2@latest

cd "$(dirname "$0")/.."
make setup
