#!/bin/bash

# Compile for Linux
GOOS=linux GOARCH=amd64 go build -o dist/blog-linux ./cmd/blog/

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o dist/blog-windows.exe ./cmd/blog/

# Check build status
if [ $? -eq 0 ]; then
    echo "Build successful"
    echo "Linux binary: dist/blog-linux"
    echo "Windows binary: dist/blog-windows.exe"
else
    echo "Build failed"
    exit 1
fi
