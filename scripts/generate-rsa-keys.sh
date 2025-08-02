#!/bin/bash

# Script to generate RSA private key for the Signed Blob Service
# This generates a 2048-bit RSA private key in PKCS#1 format

set -euo pipefail  # Exit on any error

echo "Generating RSA private key for Signed Blob Service..."

# Step 1: Generate RSA private key (this creates PKCS#8 format by default)
echo "Step 1: Generating 2048-bit RSA private key..."
openssl genrsa -out private_key_temp.pem 2048

# Step 2: Convert from PKCS#8 to PKCS#1 format (traditional RSA format)
# This is required because our Go code expects "-----BEGIN RSA PRIVATE KEY-----"
echo "Step 2: Converting to PKCS#1 format (traditional RSA format)..."
openssl rsa -in private_key_temp.pem -out private_key.pem -traditional

# Step 3: Remove the temporary file
echo "Step 3: Cleaning up temporary files..."
rm private_key_temp.pem

# Step 4: Set proper permissions (readable by Docker container)
echo "Step 4: Setting file permissions..."
chmod 644 private_key.pem

# Step 5: Verify the key format
echo "Step 5: Verifying key format..."
if head -1 private_key.pem | grep -q "BEGIN RSA PRIVATE KEY"; then
    echo "Success! RSA private key generated in correct PKCS#1 format"
    echo "File: private_key.pem"
    echo "Size: $(wc -c < private_key.pem) bytes"
    echo "Permissions: $(ls -la private_key.pem | awk '{print $1, $3, $4, $9}')"
else
    echo "Error: Key format verification failed"
    exit 1
fi

echo ""
echo "RSA private key generation complete!"
echo "Summary:"
echo "   - Format: PKCS#1 (traditional RSA)"
echo "   - Size: 2048-bit"
echo "   - File: private_key.pem"
echo "   - Permissions: 644 (readable by Docker container)"
echo ""
echo "You can now start your Docker containers with:"
echo "   docker-compose up --build"
