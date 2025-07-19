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
{{ include "olake.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
olake.io/part-of: olake
{{- end }}

{{/*
Selector labels
*/}}
{{- define "olake.selectorLabels" -}}
app.kubernetes.io/name: {{ include "olake.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use for olake-workers
*/}}
{{- define "olake.workerServiceAccountName" -}}
{{- if .Values.olakeWorker.serviceAccount.create }}
{{- default (printf "%s-workers" (include "olake.fullname" .)) .Values.olakeWorker.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.olakeWorker.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for job pods
*/}}
{{- define "olake.jobServiceAccountName" -}}
{{- if .Values.global.jobServiceAccount.name }}
{{- .Values.global.jobServiceAccount.name }}
{{- else if .Values.global.jobServiceAccount.create }}
{{- printf "%s-job" (include "olake.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- "" }}
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
Get the namespace name
*/}}
{{- define "olake.namespace" -}}
{{- .Release.Namespace -}}
{{- end -}}

{{/*
Calculate shared storage size based on NFS server backing storage
Reserves 2Gi for filesystem overhead
*/}}
{{- define "olake.sharedStorageSize" -}}
{{- $nfsSize := .Values.nfsServer.persistence.size | default "20Gi" -}}
{{- $sizeValue := regexReplaceAll "([0-9]+).*" $nfsSize "${1}" | int -}}
{{- $sizeUnit := regexReplaceAll "[0-9]+(.*)" $nfsSize "${1}" -}}
{{- $adjustedSize := sub $sizeValue 2 -}}
{{- printf "%d%s" $adjustedSize $sizeUnit -}}
{{- end -}}

