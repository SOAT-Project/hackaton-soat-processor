#!/bin/bash

set -e

echo "🧪 Testing coverage path transformation..."
echo ""

cd "$(dirname "$0")"

# Gera coverage se não existir
if [ ! -f coverage.out ]; then
    echo "📊 Generating coverage..."
    go test -coverprofile=coverage.out -covermode=atomic ./internal/...
    echo ""
fi

echo "📄 Original coverage file (first 10 lines):"
echo "=========================================="
head -n 10 coverage.out

# Backup
cp coverage.out coverage-test.out

# Transforma os paths
echo ""
echo "🔧 Transforming paths..."
sed -i 's|github.com/SOAT-Project/hackaton-soat-processor/||g' coverage-test.out

echo ""
echo "✅ Modified coverage file (first 10 lines):"
echo "=========================================="
head -n 10 coverage-test.out

echo ""
echo "📊 Checking if paths match actual files:"
echo "=========================================="

echo ""
echo "1️⃣ Coverage paths (unique files):"
grep "^internal/" coverage-test.out | cut -d: -f1 | sort -u

echo ""
echo "2️⃣ Actual files in project:"
find internal -name "*.go" ! -name "*_test.go" | sort

echo ""
echo "3️⃣ Verifying each coverage path exists:"
for file in $(grep "^internal/" coverage-test.out | cut -d: -f1 | sort -u); do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file (NOT FOUND)"
    fi
done

echo ""
echo "📊 Coverage summary:"
echo "=========================================="
go tool cover -func=coverage-test.out | tail -n 1

# Cleanup
rm coverage-test.out

echo ""
echo "✅ Test complete!"
echo ""
echo "💡 If all paths show ✅, the transformation is working correctly!"
