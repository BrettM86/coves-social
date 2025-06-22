#!/bin/bash
# Script to validate Coves lexicon schemas

echo "🔍 Validating Coves lexicon schemas..."
echo ""

go run cmd/validate-lexicon/main.go -v

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 All schemas are valid and ready to use!"
else
    echo ""
    echo "❌ Schema validation failed. Please check the errors above."
    exit 1
fi