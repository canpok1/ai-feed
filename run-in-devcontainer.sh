#!/bin/bash
cd $(dirname "$0")
devcontainer up
devcontainer exec $*
