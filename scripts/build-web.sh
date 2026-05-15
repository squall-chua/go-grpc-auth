#!/bin/bash
set -e

echo "Building static files for Web UI..."
cd web
npx nuxt generate

echo "Static files for Web UI built successfully!"
