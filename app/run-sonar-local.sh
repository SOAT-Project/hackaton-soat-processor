#!/bin/bash

set -e

echo "🔍 Running SonarQube Scanner locally with Docker..."
echo ""

cd "$(dirname "$0")"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: Docker is not running${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is running${NC}"

# 2. Verificar se o coverage existe
if [ ! -f coverage.out ]; then
    echo -e "${YELLOW}📊 Coverage file not found. Generating...${NC}"
    go test -coverprofile=coverage.out -covermode=atomic ./internal/...
else
    echo -e "${BLUE}ℹ️  Using existing coverage.out${NC}"
    read -p "Regenerate coverage? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}📊 Regenerating coverage...${NC}"
        go test -coverprofile=coverage.out -covermode=atomic ./internal/...
    fi
fi

# 3. Backup do coverage original
echo -e "${BLUE}💾 Creating backup...${NC}"
cp coverage.out coverage.out.backup

# 4. Ajustar os paths do coverage
echo -e "${YELLOW}🔧 Fixing coverage paths...${NC}"
echo "📝 Original coverage file (first 5 lines):"
head -n 5 coverage.out

sed -i 's|github.com/SOAT-Project/hackaton-soat-processor/||g' coverage.out

echo ""
echo "✅ Modified coverage file (first 5 lines):"
head -n 5 coverage.out

# 5. Verificar se os paths correspondem aos arquivos reais
echo ""
echo -e "${BLUE}🔍 Verifying path matching...${NC}"
echo "Coverage paths:"
grep "^internal/" coverage.out | cut -d: -f1 | sort -u | head -n 5

echo ""
echo "Actual files:"
find internal -name "*.go" ! -name "*_test.go" | sort | head -n 5

# 6. Verificar variáveis de ambiente
if [ -z "$SONAR_TOKEN" ]; then
    echo ""
    echo -e "${RED}❌ ERROR: SONAR_TOKEN not set!${NC}"
    echo -e "${YELLOW}💡 Export your token first:${NC}"
    echo "   export SONAR_TOKEN=your-token-here"
    echo ""
    echo "Get your token at: https://sonarcloud.io/account/security"
    exit 1
fi
echo -e "${GREEN}✅ SONAR_TOKEN is set${NC}"

# 7. Verificar sonar-project.properties
if [ ! -f sonar-project.properties ]; then
    echo -e "${RED}❌ Error: sonar-project.properties not found!${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Found sonar-project.properties${NC}"

# 8. Executar análise do SonarQube
echo ""
echo -e "${BLUE}🐳 Running SonarQube Scanner in Docker...${NC}"
echo "===================================="

docker run \
    --rm \
    -e SONAR_TOKEN="$SONAR_TOKEN" \
    -v "$(pwd):/usr/src" \
    sonarsource/sonar-scanner-cli:latest \
    -Dsonar.projectBaseDir=/usr/src \
    -Dsonar.verbose=true

# 9. Restaurar backup
echo ""
echo -e "${BLUE}🔄 Restoring original coverage.out...${NC}"
mv coverage.out.backup coverage.out

# 10. Resumo final
echo ""
echo -e "${GREEN}✅ Scan complete!${NC}"
echo "===================================="
echo ""
echo -e "${BLUE}🌐 View results at:${NC}"
echo "   https://sonarcloud.io/project/overview?id=SOAT-Project_hackaton-soat-processor"
echo ""
echo -e "${YELLOW}💡 Tips:${NC}"
echo "   - Check the 'Coverage' tab in SonarCloud"
echo "   - Look for 'internal/' files in the coverage report"
echo "   - Verify that coverage percentage matches local tests"
echo ""
