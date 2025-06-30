#!/bin/bash
# Simple validation script for Helm chart structure

set -e

CHART_DIR="./olake"
echo "🔍 Validating OLake Helm Chart structure..."

# Check if chart directory exists
if [ ! -d "$CHART_DIR" ]; then
    echo "❌ Chart directory not found: $CHART_DIR"
    exit 1
fi

echo "✅ Chart directory exists"

# Check required files
required_files=(
    "Chart.yaml"
    "values.yaml"
    "templates/_helpers.tpl"
    "templates/namespace.yaml"
)

for file in "${required_files[@]}"; do
    if [ -f "$CHART_DIR/$file" ]; then
        echo "✅ Found: $file"
    else
        echo "❌ Missing: $file"
        exit 1
    fi
done

# Check template directories
template_dirs=(
    "templates/storage"
    "templates/postgresql"
    "templates/elasticsearch"
    "templates/temporal"
    "templates/olake-ui"
    "templates/olake-worker"
)

for dir in "${template_dirs[@]}"; do
    if [ -d "$CHART_DIR/$dir" ]; then
        echo "✅ Found template directory: $dir"
        # Count files in directory
        file_count=$(find "$CHART_DIR/$dir" -name "*.yaml" | wc -l)
        echo "   📄 Contains $file_count YAML files"
    else
        echo "❌ Missing template directory: $dir"
        exit 1
    fi
done

# Check values files
values_files=(
    "values/development.yaml"
    "values/staging.yaml" 
    "values/production.yaml"
)

for file in "${values_files[@]}"; do
    if [ -f "$CHART_DIR/$file" ]; then
        echo "✅ Found values file: $file"
    else
        echo "❌ Missing values file: $file"
        exit 1
    fi
done

# Basic YAML syntax check (if yq is available)
if command -v yq &> /dev/null; then
    echo "🔍 Checking YAML syntax..."
    for yaml_file in $(find "$CHART_DIR" -name "*.yaml"); do
        if yq eval '.' "$yaml_file" > /dev/null 2>&1; then
            echo "✅ Valid YAML: $yaml_file"
        else
            echo "❌ Invalid YAML: $yaml_file"
            exit 1
        fi
    done
else
    echo "⚠️  yq not found, skipping YAML syntax validation"
fi

# Summary
total_templates=$(find "$CHART_DIR/templates" -name "*.yaml" -o -name "*.tpl" | wc -l)
echo ""
echo "📊 Chart Summary:"
echo "   📁 Total template files: $total_templates"
echo "   🎯 Environment values: ${#values_files[@]}"
echo "   🔧 Component directories: ${#template_dirs[@]}"
echo ""
echo "🎉 Helm chart validation completed successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Install Helm: https://helm.sh/docs/intro/install/"
echo "   2. Test with: helm template olake ./olake"
echo "   3. Deploy with: helm install olake ./olake -f ./olake/values.yaml"