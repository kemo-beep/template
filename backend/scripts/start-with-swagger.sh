#!/bin/bash

echo "🚀 Starting Mobile Backend with Swagger regeneration..."

# Check if swag is available
if command -v swag &> /dev/null; then
    echo "📝 Regenerating Swagger documentation..."
    swag init
    if [ $? -eq 0 ]; then
        echo "✅ Swagger documentation regenerated successfully"
    else
        echo "⚠️  Warning: Failed to regenerate Swagger docs, continuing with existing docs"
    fi
else
    echo "⚠️  Warning: swag command not found, skipping Swagger regeneration"
fi

echo "🎯 Starting application..."
exec ./main
