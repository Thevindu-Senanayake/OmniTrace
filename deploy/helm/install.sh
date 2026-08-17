#!/usr/bin/env bash
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
#   ./deploy/helm/install.sh
#   RELEASE=eco NAMESPACE=eco ./deploy/helm/install.sh
#   ./deploy/helm/install.sh --dry-run --debug      # forwarded to helm
set -euo pipefail

# Resolve paths from the script's own location so the working directory does
# not matter. install.sh lives at deploy/helm/, so the repo root is two levels
# up and the chart is the sibling directory.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHART="$SCRIPT_DIR/ecommerce-observability"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OBS="$REPO_ROOT/observability"

RELEASE="${RELEASE:-ecommerce}"
NAMESPACE="${NAMESPACE:-}"

# key (values path under observability.configs) -> file (relative to observability/)
CONFIGS=(
  "otelCollector=otel-collector-config.yaml"
  "tempo=tempo-config.yaml"
  "loki=loki-config.yaml"
  "prometheus=prometheus.yml"
  "grafanaDatasources=grafana/provisioning/datasources/datasources.yaml"
  "grafanaDashboardsProvider=grafana/provisioning/dashboards/dashboards.yaml"
  "dashboardServiceOverview=grafana/provisioning/dashboards/service-overview.json"
  "dashboardPostgresql=grafana/provisioning/dashboards/postgresql.json"
  "dashboardOrderOutcomes=grafana/provisioning/dashboards/order-outcomes.json"
)

SET_FILE_ARGS=()
for entry in "${CONFIGS[@]}"; do
  key="${entry%%=*}"
  rel="${entry#*=}"
  path="$OBS/$rel"
  if [[ ! -f "$path" ]]; then
    echo "config file not found: $path (expected under $OBS)" >&2
    exit 1
  fi
  SET_FILE_ARGS+=("--set-file" "observability.configs.$key=$path")
done

NS_ARGS=()
if [[ -n "$NAMESPACE" ]]; then
  NS_ARGS+=("--namespace" "$NAMESPACE")
fi

echo "==> helm upgrade --install $RELEASE $CHART (+9 --set-file configs)"
helm upgrade --install "$RELEASE" "$CHART" "${NS_ARGS[@]}" "${SET_FILE_ARGS[@]}" "$@"

echo
echo "Watch rollout:   kubectl get pods -w"
echo "Open Grafana:    kubectl port-forward svc/grafana 3300:3000  ->  http://localhost:3300"
