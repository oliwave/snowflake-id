{{/*
Expand the name of the chart.
*/}}
{{- define "snowflake-id.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "snowflake-id.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "snowflake-id.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "snowflake-id.labels" -}}
helm.sh/chart: {{ include "snowflake-id.chart" . }}
{{ include "snowflake-id.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "snowflake-id.selectorLabels" -}}
app.kubernetes.io/name: {{ include "snowflake-id.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "snowflake-id.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "snowflake-id.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Chart name prefix for resource names
Strip the "-controller" suffix from the default .Chart.Name if the nameOverride is not specified.
This enables using a shorter name for the resources, for example aws-load-balancer-webhook.
*/}}
{{- define "snowflake-id.namePrefix" -}}
{{- $defaultNamePrefix := .Chart.Name | trimSuffix "-controller" -}}
{{- default $defaultNamePrefix .Values.nameOverride | trunc 42 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the webhook service
*/}}
{{- define "snowflake-id.webhookService" -}}
{{- printf "%s-webhook-service" (include "snowflake-id.namePrefix" .) -}}
{{- end -}}

{{/*
Create the name of the webhook cert secret
*/}}
{{- define "snowflake-id.webhookCertSecret" -}}
{{- printf "%s-tls" (include "snowflake-id.namePrefix" .) -}}
{{- end -}}

{{/*
Generate certificates for webhook
*/}}
{{- define "snowflake-id.webhookCerts" -}}
{{- $serviceName := (include "snowflake-id.webhookService" .) -}}
{{- $secretName := (include "snowflake-id.webhookCertSecret" .) -}}
{{- $secret := lookup "v1" "Secret" .Release.Namespace $secretName -}}
{{- if and .Values.keepTLSSecret $secret }}
caCert: {{ index $secret.data "ca.crt" }}
clientCert: {{ index $secret.data "tls.crt" }}
clientKey: {{ index $secret.data "tls.key" }}
{{- else -}}
{{- $altNames := list (printf "%s.%s" $serviceName .Release.Namespace) (printf "%s.%s.svc" $serviceName .Release.Namespace) (printf "%s.%s.svc.cluster.local" $serviceName .Release.Namespace) -}}
{{- $ca := genCA "snowflake-id-ca" 3650 -}}
{{- $cert := genSignedCert (include "snowflake-id.fullname" .) nil $altNames 3650 $ca -}}
caCert: {{ $ca.Cert | b64enc }}
clientCert: {{ $cert.Cert | b64enc }}
clientKey: {{ $cert.Key | b64enc }}
{{- end -}}
{{- end -}}