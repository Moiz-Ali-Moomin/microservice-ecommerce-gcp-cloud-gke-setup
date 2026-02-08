{{/*
Expand the name of the chart.
*/}}
{{- define "conversion-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "conversion-webhook.fullname" -}}
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
Common labels
*/}}
{{- define "conversion-webhook.labels" -}}
helm.sh/chart: {{ include "conversion-webhook.chart" . }}
{{ include "conversion-webhook.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "conversion-webhook.selectorLabels" -}}
app.kubernetes.io/name: {{ include "conversion-webhook.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Chart name and version
*/}}
{{- define "conversion-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Construct full image reference
*/}}
{{- define "conversion-webhook.image" -}}
{{- $registry := .Values.global.image.registry | default .Values.image.registry -}}
{{- $project := .Values.global.image.project | default .Values.image.project -}}
{{- $repo := .Values.global.image.repository | default .Values.image.repository -}}
{{- $name := .Values.image.name | default .Chart.Name -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion | default "latest" -}}
{{- printf "%s/%s/%s/%s:%s" $registry $project $repo $name $tag -}}
{{- end -}}

{{/*
Construct GCP Service Account Email
*/}}
{{- define "conversion-webhook.serviceAccountEmail" -}}
{{- $saName := .Values.serviceAccount.gcpServiceAccount -}}
{{- $projectId := .Values.global.image.project | default .Values.global.gcp.projectId -}}
{{- printf "%s@%s.iam.gserviceaccount.com" $saName $projectId -}}
{{- end -}}
