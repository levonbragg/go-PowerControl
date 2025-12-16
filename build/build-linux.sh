#!/bin/bash
echo "Building Go PowerControl for Linux..."
cd ..
wails build -platform linux/amd64
echo "Build complete! Output is in build/bin/"
