#!/bin/bash
set -e

mkdir -p keys

if [ ! -f keys/id_rs256 ]; then
    echo "Generating RSA private key..."
    openssl genpkey -algorithm RSA -out keys/id_rs256 -pkeyopt rsa_keygen_bits:2048
fi

if [ ! -f keys/id_rs256.pub ]; then
    echo "Generating RSA public key..."
    openssl rsa -pubout -in keys/id_rs256 -out keys/id_rs256.pub
fi

echo "RSA keys are ready in keys/ directory."
