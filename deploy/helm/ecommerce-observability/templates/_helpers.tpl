{{/*
Common labels applied to every resource.
*/}}
{{- define "eco.labels" -}}
app.kubernetes.io/part-of: ecommerce-observability
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}

{{/*
Per-component selector labels. Call with a dict: (dict "name" "order-service").
*/}}
{{- define "eco.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
{{- end -}}
