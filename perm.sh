#!/bin/bash

# This script recursively removes quarantine flags from all files in the current directory

echo "Removing quarantine flags recursively in the current directory..."

ROOT="$(pwd)"

# find current script folder
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Find all files and directories with quarantine attribute and remove the attribute in the current directory
cd DIR
find . -exec xattr -d com.apple.quarantine {} \; 2>/dev/null
cd ROOT
echo "Quarantine flags removed."