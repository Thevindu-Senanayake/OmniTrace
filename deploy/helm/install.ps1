# Installs (or upgrades) the ecommerce-observability chart WITH the LGTM stack.
#
# Why this wrapper exists: the observability configs (collector, Tempo, Loki,
# Prometheus, Grafana datasources + dashboards) are the single source of truth
# at repo-root observability/, shared with docker-compose. Helm cannot read
# files outside the chart directory (.Files.Get is sandboxed), so instead of
# vendoring a drift-prone copy into the chart, each config's *content* is passed
# in at install time with `--set-file`. The chart's ConfigMap templates inject
# those values verbatim; because Helm treats --set-file values as data (never
# re-templated), the double-brace legend syntax in the dashboard JSON survives
# untouched. A bare `helm install` without these flags trips the `required`
# guards in the templates and fails fast — run this script instead.
#
# Prerequisites (one-time, not done here):
#   - Build the service images:      docker compose build
#   - Load them into the cluster:     ./deploy/load-images.ps1
#   - Install an ingress controller:  see templates/NOTES.txt
#
# Usage:
#   ./deploy/helm/install.ps1
#   ./deploy/helm/install.ps1 -Release eco -Namespace eco
#   ./deploy/helm/install.ps1 --dry-run --debug        # forwarded to helm

param(
    [string]$Release = "ecommerce",
    [string]$Namespace = "",
    # Anything after the named params is forwarded to helm verbatim
    # (e.g. --dry-run, --debug, --create-namespace, extra --set flags).
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$HelmArgs
)

$ErrorActionPreference = "Stop"

# Resolve paths from the script's own location so the working directory does
# not matter. install.ps1 lives at deploy/helm/, so the repo root is two levels
# up and the chart is the sibling directory.
$chart   = Join-Path $PSScriptRoot "ecommerce-observability"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$obs     = Join-Path $repoRoot "observability"

# key (values path under observability.configs) -> file on disk
$configs = [ordered]@{
    "otelCollector"           = Join-Path $obs "otel-collector-config.yaml"
    "tempo"                   = Join-Path $obs "tempo-config.yaml"
    "loki"                    = Join-Path $obs "loki-config.yaml"
    "prometheus"              = Join-Path $obs "prometheus.yml"
    "grafanaDatasources"      = Join-Path $obs "grafana\provisioning\datasources\datasources.yaml"
    "grafanaDashboardsProvider" = Join-Path $obs "grafana\provisioning\dashboards\dashboards.yaml"
    "dashboardServiceOverview" = Join-Path $obs "grafana\provisioning\dashboards\service-overview.json"
    "dashboardPostgresql"     = Join-Path $obs "grafana\provisioning\dashboards\postgresql.json"
    "dashboardOrderOutcomes"  = Join-Path $obs "grafana\provisioning\dashboards\order-outcomes.json"
}

$setFileArgs = @()
foreach ($key in $configs.Keys) {
    $path = $configs[$key]
    if (-not (Test-Path $path)) {
        throw "config file not found: $path (expected under $obs)"
    }
    $setFileArgs += "--set-file"
    $setFileArgs += "observability.configs.$key=$path"
}

$nsArgs = @()
if ($Namespace -ne "") {
    $nsArgs += "--namespace"
    $nsArgs += $Namespace
}

Write-Host "==> helm upgrade --install $Release $chart (+9 --set-file configs)"
helm upgrade --install $Release $chart @nsArgs @setFileArgs @HelmArgs
if ($LASTEXITCODE -ne 0) {
    throw "helm exited with code $LASTEXITCODE"
}

Write-Host ""
Write-Host "Watch rollout:   kubectl get pods -w"
Write-Host "Open Grafana:    kubectl port-forward svc/grafana 3300:3000  ->  http://localhost:3300"
