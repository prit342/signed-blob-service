#!/bin/bash

set -euo pipefail

echo "Installing buf tools and Go protobuf plugins..."

# Substitute PREFIX for your install prefix.
# Substitute VERSION for the current released version.
PREFIX="$HOME/tools" && \
VERSION="1.55.1" && \
mkdir -p "$HOME/tools" &&
curl -sSL \
"https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m).tar.gz" | \
tar -xvzf - -C "${PREFIX}" --strip-components 1

## move buf tools
echo "Moving buf tools to $HOME/tools..."
mv -v "${HOME}/tools/bin/protoc-gen-buf-breaking" "${HOME}/tools/"
mv -v "${HOME}/tools/bin/protoc-gen-buf-lint" "${HOME}/tools/"
mv -v "${HOME}/tools/bin/buf" "${HOME}/tools/"

## install Go protobuf plugins
echo "Installing protoc-gen-go..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

echo "Installing protoc-gen-go-grpc..."
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
