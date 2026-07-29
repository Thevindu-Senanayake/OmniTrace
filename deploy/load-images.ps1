# Loads locally built images into Docker Desktop's Kubernetes.
#
# Docker Desktop's Kubernetes keeps its own containerd image store, separate
# from the Docker daemon. A `docker build` (or `docker compose build`) is
# therefore invisible to the cluster, and because the chart uses
# imagePullPolicy: IfNotPresent, pods silently keep running whatever image was
# cached at first install. Symptom: a fix you just built has no effect in the
# cluster, and the pod reports "already present on machine".
#
# Usage:
#   ./deploy/load-images.ps1                     # all services
#   ./deploy/load-images.ps1 -Services order-service,payment-service
#
# Restarts the matching Deployments so the new image is picked up.

param(
    [string[]]$Services = @(
        "api-gateway",
        "product-catalog",
        "order-service",
        "inventory-service",
        "payment-service",
        "payment-gateway",
        "notification-service"
    ),
    [string]$Node = "desktop-control-plane",
    [switch]$SkipRestart
)

$ErrorActionPreference = "Stop"

foreach ($svc in $Services) {
    $image = "ieeeyp-$svc`:latest"
    $tar = Join-Path $env:TEMP "$svc.tar"

    Write-Host "==> $image"

    docker image inspect $image *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "    not built locally - skipping (run: docker compose build $svc)"
        continue
    }

    docker save $image -o $tar
    # Redirect from a file rather than piping: PowerShell's pipeline mangles
    # binary streams, which surfaces as "archive/tar: invalid tar header".
    cmd /c "docker exec -i $Node ctr -n k8s.io images import - < `"$tar`"" | Out-Null
    Remove-Item $tar -Force

    if ($LASTEXITCODE -ne 0) {
        throw "import failed for $image"
    }
    Write-Host "    imported"
}

if (-not $SkipRestart) {
    foreach ($svc in $Services) {
        kubectl rollout restart "deploy/$svc" *> $null
    }
    Write-Host ""
    Write-Host "Restarted deployments. Watch with: kubectl get pods -w"
}
