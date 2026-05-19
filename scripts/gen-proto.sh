#!/bin/bash
set -e

# Clean up old generated files
echo "Cleaning up old generated files..."
rm -rf api/v1/*
rm -rf api/swagger/*.json

# Generate new files
echo "Generating new files..."
buf generate

# Moving generated Go files to api/v1
echo "Moving generated Go files to api/v1..."
mv api/v1/api/proto/admin/v1 api/v1/admin
mv api/v1/api/proto/auth/v1 api/v1/auth
mv api/v1/api/proto/options/v1 api/v1/options
mv api/v1/api/proto/greeter/v1 api/v1/greeter
rm -rf api/v1/api

echo "Code generation complete!"
