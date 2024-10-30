#!/bin/bash

# This script recursively removes quarantine flags from all files in the current directory

echo "Removing quarantine flags recursively in the current directory..."

# Find all files and directories with quarantine attribute and remove the attribute
find . -exec xattr -d com.apple.quarantine {} \; 2>/dev/null

echo "Quarantine flags removed."