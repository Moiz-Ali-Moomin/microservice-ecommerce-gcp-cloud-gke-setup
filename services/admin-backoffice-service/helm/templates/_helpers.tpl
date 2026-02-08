{{/*
Expand the name of the chart.
*/}}
{{- define "admin-backoffice-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "admin-backoffice-service.fullname" -}}
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
{{- define "admin-backoffice-service.labels" -}}
helm.sh/chart: {{ include "admin-backoffice-service.chart" . }}
{{ include "admin-backoffice-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "admin-backoffice-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "admin-backoffice-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Chart name and version
*/}}
{{- define "admin-backoffice-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Construct full image reference
*/}}
{{- define "admin-backoffice-service.image" -}}
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
{{- define "admin-backoffice-service.serviceAccountEmail" -}}
{{- $saName := .Values.serviceAccount.gcpServiceAccount -}}
{{- $projectId := .Values.global.image.project | default .Values.global.gcp.projectId -}}
{{- printf "%s@%s.iam.gserviceaccount.com" $saName $projectId -}}
{{- end -}}
