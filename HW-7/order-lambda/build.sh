#!/bin/bash
# Build the Lambda zip for Go on provided.al2 runtime.
# Must be cross-compiled for linux/amd64 and named "bootstrap".
#
# Run this BEFORE terraform apply.
set -e

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap .
zip processor.zip bootstrap
rm bootstrap
echo "Built processor.zip"
