$services = @(
    "admin-backoffice-service",
    "analytics-ingest-service",
    "api-gateway",
    "attribution-service",
    "audit-service",
    "auth-service",
    "cart-service",
    "conversion-webhook",
    "feature-flag-service",
    "landing-service",
    "notification-service",
    "offer-service",
    "product-service",
    "redirect-service",
    "reporting-service",
    "storefront-service",
    "user-service"
)

$deps = @(
    "github.com/IBM/sarama@v1.43.2",
    "github.com/google/uuid@v1.6.0",
    "github.com/prometheus/client_golang@v1.19.1",
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.52.0",
    "go.opentelemetry.io/otel@v1.27.0",
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.27.0",
    "go.opentelemetry.io/otel/sdk@v1.27.0",
    "go.uber.org/zap@v1.27.0",
    "google.golang.org/grpc@v1.64.0",
    "google.golang.org/protobuf@v1.34.2",
    "golang.org/x/net@v0.26.0"
)

# 1. Fix Shared Lib First
Write-Host "Fixing shared-lib..."
Set-Location "services/shared-lib"
(Get-Content go.mod) -replace "^go \d+\.\d+(\.\d+)?$", "go 1.22.0" | Set-Content go.mod
go mod tidy
if ($LASTEXITCODE -ne 0) { Write-Host "Error tidying shared-lib"; exit 1 }
Set-Location ../..

# 2. Fix All Services
foreach ($svc in $services) {
    Write-Host "Fixing $svc..."
    if (Test-Path "services/$svc") {
        Set-Location "services/$svc"
        (Get-Content go.mod) -replace "^go \d+\.\d+(\.\d+)?$", "go 1.22.0" | Set-Content go.mod
        
        # Explicitly downgrade deps
        $cmd = "go get " + ($deps -join " ")
        Invoke-Expression $cmd
        
        go mod tidy
        Set-Location ../..
    }
    else {
        Write-Host "Warning: Service directory not found: services/$svc"
    }
}

Write-Host "Dependency Downgrade Complete."
