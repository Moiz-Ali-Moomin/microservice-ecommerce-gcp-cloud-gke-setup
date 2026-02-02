#!/bin/bash
# scripts/setup-sonar.sh
# 🤖 AUTOMATED SONARQUBE SETUP
# Requirements: curl, jq, gh (GitHub CLI)

set -e

PROJECT_ID=$1
if [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <PROJECT_ID>"
    exit 1
fi

echo "🔍 Getting SonarQube Ingress IP..."
INGRESS_IP=$(kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

if [ -z "$INGRESS_IP" ]; then
    echo "❌ Could not find Ingress IP. Ensure Istio IngressGateway has an External IP."
    exit 1
fi

SONAR_URL="http://$INGRESS_IP/sonar"
SONAR_API="$SONAR_URL/api"
ADMIN_AUTH="admin:admin" # Default creds

echo "⏳ Waiting for SonarQube to be ready at $SONAR_URL..."
until curl -s -f -u $ADMIN_AUTH "$SONAR_API/system/status" | grep -q "UP"; do
    echo "    Still waiting..."
    sleep 10
done

echo "✅ SonarQube is UP!"

# 1. Generate Token
echo "🔑 Generating Analysis Token..."
TOKEN_RESPONSE=$(curl -s -u $ADMIN_AUTH -X POST "$SONAR_API/user_tokens/generate?name=github-ci-token")
SONAR_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.token')

if [ "$SONAR_TOKEN" == "null" ]; then
    echo "⚠️ Token might already exist. Attempting to revoke and regenerate..."
    curl -s -u $ADMIN_AUTH -X POST "$SONAR_API/user_tokens/revoke?name=github-ci-token"
    TOKEN_RESPONSE=$(curl -s -u $ADMIN_AUTH -X POST "$SONAR_API/user_tokens/generate?name=github-ci-token")
    SONAR_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.token')
fi

# 2. Configure Webhook
echo "🔗 Configuring Quality Gate Webhook..."
curl -s -u $ADMIN_AUTH -X POST "$SONAR_API/webhooks/create?name=GitHub-Actions&url=https://api.github.com/repos/$(gh repo view --json nameWithOwner -q .nameWithOwner)/dispatches"

# 3. Set GitHub Secrets
echo "🐙 Setting GitHub Secrets..."
gh secret set SONAR_TOKEN --body "$SONAR_TOKEN"
gh secret set SONAR_HOST_URL --body "$SONAR_URL"

echo "--------------------------------------------------"
echo "🎉 SONARQUBE AUTOMATION COMPLETE!"
echo "URL: $SONAR_URL"
echo "Secrets configured in GitHub: SONAR_TOKEN, SONAR_HOST_URL"
echo "--------------------------------------------------"
