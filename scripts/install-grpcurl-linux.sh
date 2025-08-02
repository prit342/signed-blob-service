#/bin/bash

set -euo pipefail

# Install grpcurl (for testing gRPC services)
echo "Installing grpcurl..."
GO111MODULE=on go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

echo "Installation complete!"
echo "Make sure to add $HOME/tools to your PATH:"
echo "export PATH=\"\$HOME/tools:\$PATH\""
echo "Make sure $HOME/go/bin is in your PATH to use grpcurl."