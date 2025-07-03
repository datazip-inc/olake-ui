{{/*
Expand the name of the chart.
*/}}
{{- define "olake.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "olake.fullname" -}}
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
{{- define "olake.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "olake.labels" -}}
helm.sh/chart: {{ include "olake.chart" . }}
{{ include "olake.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "olake.selectorLabels" -}}
app.kubernetes.io/name: {{ include "olake.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use for olake-worker
*/}}
{{- define "olake.workerServiceAccountName" -}}
{{- if .Values.olakeWorker.serviceAccount.create }}
{{- default (printf "%s-worker" (include "olake.fullname" .)) .Values.olakeWorker.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.olakeWorker.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Shared storage PVC name
*/}}
{{- define "olake.sharedStoragePVC" -}}
{{- if .Values.nfsServer.enabled -}}
{{ include "olake.fullname" . }}-shared-storage
{{- else -}}
{{ .Values.nfsServer.external.name }}
{{- end -}}
{{- end -}}

{{/*
NFS Server Static IP
*/}}
{{- define "olake.nfsStaticIP" -}}
{{- $configMap := lookup "v1" "ConfigMap" .Values.namespace.name (printf "%s-nfs-ip-config" (include "olake.fullname" .)) -}}
{{- if and $configMap $configMap.data $configMap.data.staticIP -}}
{{- $configMap.data.staticIP -}}
{{- else -}}
172.16.100.100
{{- end -}}
{{- end -}}

{{/*
Get the namespace name
*/}}
{{- define "olake.namespace" -}}
{{- .Values.namespaceOverride | default "olake" -}}
{{- end -}}