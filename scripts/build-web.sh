#!/bin/bash
set -e

echo "Building static files for Web UI..."
cd web
NUXT_PUBLIC_API_BASE=__API_BASE_PLACEHOLDER__
NUXT_PUBLIC_APP_NAME=__APP_NAME_PLACEHOLDER__
npx nuxt generate

echo "Static files for Web UI built successfully!"
