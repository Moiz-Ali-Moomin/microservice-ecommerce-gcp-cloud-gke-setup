{{/*
Expand the name of the chart.
*/}}
{{- define "landing-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "landing-service.fullname" -}}
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
{{- define "landing-service.labels" -}}
helm.sh/chart: {{ include "landing-service.chart" . }}
{{ include "landing-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "landing-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "landing-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Chart name and version
*/}}
{{- define "landing-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Construct full image reference
*/}}
{{- define "landing-service.image" -}}
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
{{- define "landing-service.serviceAccountEmail" -}}
{{- $saName := .Values.serviceAccount.gcpServiceAccount -}}
{{- $projectId := .Values.global.image.project | default .Values.global.gcp.projectId -}}
{{- printf "%s@%s.iam.gserviceaccount.com" $saName $projectId -}}
{{- end -}}
