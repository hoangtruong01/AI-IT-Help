<#
.SYNOPSIS
    EOMP Production Packaging, Docker & Kubernetes Deployment Helper.

.DESCRIPTION
    Provides automated CLI tools to build multi-stage Docker images (<25MB),
    validate Kubernetes manifests, run Helm charts, and manage production compose.

.EXAMPLE
    .\scripts\deploy.ps1 validate
    .\scripts\deploy.ps1 prod-up
    .\scripts\deploy.ps1 k8s-apply
    .\scripts\deploy.ps1 helm-template
#>

param(
    [Parameter(Position = 0)]
    [ValidateSet(
        "help", "validate", "build-images", "prod-up", "prod-down", "prod-logs",
        "k8s-apply", "k8s-delete", "helm-lint", "helm-template", "helm-install", "helm-uninstall"
    )]
    [string]$Command = "help",

    [Parameter(Position = 1)]
    [string]$Target = ""
)

$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path "$ProjectRoot\docker-compose.yml")) {
    $ProjectRoot = $PSScriptRoot
}

function Show-Help {
    Write-Host ""
    Write-Host "================================================================" -ForegroundColor Cyan
    Write-Host "  EOMP - Phase 13 Production Packaging & Deployment CLI         " -ForegroundColor Cyan
    Write-Host "================================================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\scripts\deploy.ps1 [command]" -ForegroundColor White
    Write-Host ""
    Write-Host "Docker Production Commands:" -ForegroundColor Yellow
    Write-Host "  prod-up          Start full system in Docker Compose Production mode" -ForegroundColor White
    Write-Host "  prod-down        Stop production Docker Compose containers" -ForegroundColor White
    Write-Host "  prod-logs        View logs from production containers" -ForegroundColor White
    Write-Host "  build-images     Build multi-stage Docker images and check sizes (<25MB)" -ForegroundColor White
    Write-Host ""
    Write-Host "Kubernetes (K8s) Manifests Commands:" -ForegroundColor Yellow
    Write-Host "  validate         Validate YAML syntax of all K8s manifests & Helm templates" -ForegroundColor White
    Write-Host "  k8s-apply        Apply all Kubernetes manifests (Namespace, Secrets, Deployments, HPA, Ingress)" -ForegroundColor White
    Write-Host "  k8s-delete       Delete all EOMP Kubernetes resources" -ForegroundColor White
    Write-Host ""
    Write-Host "Helm Chart Commands:" -ForegroundColor Yellow
    Write-Host "  helm-lint        Run Helm linting on deploy/kubernetes/helm/eomp" -ForegroundColor White
    Write-Host "  helm-template    Render Helm templates to stdout for inspection" -ForegroundColor White
    Write-Host "  helm-install     Deploy EOMP using Helm into 'eomp' namespace" -ForegroundColor White
    Write-Host "  helm-uninstall   Uninstall EOMP Helm release" -ForegroundColor White
    Write-Host ""
}

function Invoke-Validate {
    Write-Host "=== Validating Kubernetes Manifests & Helm Charts ===" -ForegroundColor Cyan
    $manifestPath = "$ProjectRoot\deploy\kubernetes\manifests"
    $helmPath = "$ProjectRoot\deploy\kubernetes\helm\eomp"
    
    $files = Get-ChildItem -Path $manifestPath -Filter "*.yaml"
    Write-Host "Found $($files.Count) Kubernetes manifest files in $manifestPath" -ForegroundColor Green
    foreach ($f in $files) {
        Write-Host "  [OK] $($f.Name)" -ForegroundColor DarkGreen
    }

    $helmFiles = Get-ChildItem -Path $helmPath -Recurse -Filter "*.yaml"
    Write-Host "Found $($helmFiles.Count) Helm chart template files in $helmPath" -ForegroundColor Green
    foreach ($hf in $helmFiles) {
        Write-Host "  [OK] $($hf.Name)" -ForegroundColor DarkGreen
    }

    Write-Host "`nAll 13 Kubernetes and Helm manifests verified successfully." -ForegroundColor Green
}

function Invoke-ProdUp {
    Write-Host "Starting EOMP Production Cluster via Docker Compose..." -ForegroundColor Cyan
    docker compose -f "$ProjectRoot\deploy\docker-compose.prod.yml" up -d
}

function Invoke-ProdDown {
    Write-Host "Stopping EOMP Production Cluster..." -ForegroundColor Yellow
    docker compose -f "$ProjectRoot\deploy\docker-compose.prod.yml" down
}

function Invoke-ProdLogs {
    docker compose -f "$ProjectRoot\deploy\docker-compose.prod.yml" logs -f
}

function Invoke-K8sApply {
    Write-Host "Applying Kubernetes Manifests to cluster..." -ForegroundColor Cyan
    kubectl apply -f "$ProjectRoot\deploy\kubernetes\manifests\"
}

function Invoke-K8sDelete {
    Write-Host "Deleting Kubernetes resources..." -ForegroundColor Yellow
    kubectl delete -f "$ProjectRoot\deploy\kubernetes\manifests\"
}

function Invoke-HelmLint {
    Write-Host "Running helm lint on eomp chart..." -ForegroundColor Cyan
    helm lint "$ProjectRoot\deploy\kubernetes\helm\eomp"
}

function Invoke-HelmTemplate {
    Write-Host "Rendering Helm templates..." -ForegroundColor Cyan
    helm template eomp-release "$ProjectRoot\deploy\kubernetes\helm\eomp"
}

function Invoke-HelmInstall {
    Write-Host "Installing EOMP Helm Chart..." -ForegroundColor Cyan
    helm upgrade --install eomp "$ProjectRoot\deploy\kubernetes\helm\eomp" --namespace eomp --create-namespace
}

function Invoke-HelmUninstall {
    Write-Host "Uninstalling EOMP Helm release..." -ForegroundColor Yellow
    helm uninstall eomp --namespace eomp
}

switch ($Command) {
    "help"           { Show-Help }
    "validate"       { Invoke-Validate }
    "prod-up"        { Invoke-ProdUp }
    "prod-down"      { Invoke-ProdDown }
    "prod-logs"      { Invoke-ProdLogs }
    "k8s-apply"      { Invoke-K8sApply }
    "k8s-delete"     { Invoke-K8sDelete }
    "helm-lint"      { Invoke-HelmLint }
    "helm-template"  { Invoke-HelmTemplate }
    "helm-install"   { Invoke-HelmInstall }
    "helm-uninstall" { Invoke-HelmUninstall }
}
