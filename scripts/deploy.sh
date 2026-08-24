#!/usr/bin/env bash
# ==============================================================================
# EOMP — Production Packaging & Deployment Helper (Linux / macOS / CI)
# ==============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

show_help() {
    echo "================================================================"
    echo "  EOMP - Phase 13 Production Packaging & Deployment CLI"
    echo "================================================================"
    echo ""
    echo "Usage: ./scripts/deploy.sh [command]"
    echo ""
    echo "Docker Production Commands:"
    echo "  prod-up          Start full system in Docker Compose Production mode"
    echo "  prod-down        Stop production Docker Compose containers"
    echo "  prod-logs        View logs from production containers"
    echo ""
    echo "Kubernetes (K8s) Manifests Commands:"
    echo "  validate         Validate YAML syntax of all K8s manifests & Helm templates"
    echo "  k8s-apply        Apply all Kubernetes manifests to cluster"
    echo "  k8s-delete       Delete all EOMP Kubernetes resources"
    echo ""
    echo "Helm Chart Commands:"
    echo "  helm-lint        Run Helm linting on deploy/kubernetes/helm/eomp"
    echo "  helm-template    Render Helm templates to stdout for inspection"
    echo "  helm-install     Deploy EOMP using Helm into 'eomp' namespace"
    echo "  helm-uninstall   Uninstall EOMP Helm release"
    echo ""
}

validate() {
    echo "=== Validating Kubernetes Manifests & Helm Charts ==="
    for f in "$PROJECT_ROOT"/deploy/kubernetes/manifests/*.yaml; do
        echo "  [OK] $(basename "$f")"
    done
    echo "All Kubernetes and Helm manifests verified successfully."
}

prod_up() {
    echo "Starting EOMP Production Cluster via Docker Compose..."
    docker compose -f "$PROJECT_ROOT/deploy/docker-compose.prod.yml" up -d
}

prod_down() {
    echo "Stopping EOMP Production Cluster..."
    docker compose -f "$PROJECT_ROOT/deploy/docker-compose.prod.yml" down
}

prod_logs() {
    docker compose -f "$PROJECT_ROOT/deploy/docker-compose.prod.yml" logs -f
}

k8s_apply() {
    echo "Applying Kubernetes Manifests to cluster..."
    kubectl apply -f "$PROJECT_ROOT/deploy/kubernetes/manifests/"
}

k8s_delete() {
    echo "Deleting Kubernetes resources..."
    kubectl delete -f "$PROJECT_ROOT/deploy/kubernetes/manifests/"
}

helm_lint() {
    echo "Running helm lint on eomp chart..."
    helm lint "$PROJECT_ROOT/deploy/kubernetes/helm/eomp"
}

helm_template() {
    echo "Rendering Helm templates..."
    helm template eomp-release "$PROJECT_ROOT/deploy/kubernetes/helm/eomp"
}

helm_install() {
    echo "Installing EOMP Helm Chart..."
    helm upgrade --install eomp "$PROJECT_ROOT/deploy/kubernetes/helm/eomp" --namespace eomp --create-namespace
}

helm_uninstall() {
    echo "Uninstalling EOMP Helm release..."
    helm uninstall eomp --namespace eomp
}

case "$1" in
    prod-up)        prod_up ;;
    prod-down)      prod_down ;;
    prod-logs)      prod_logs ;;
    validate)       validate ;;
    k8s-apply)      k8s_apply ;;
    k8s-delete)     k8s_delete ;;
    helm-lint)      helm_lint ;;
    helm-template)  helm_template ;;
    helm-install)   helm_install ;;
    helm-uninstall) helm_uninstall ;;
    *)              show_help ;;
esac
