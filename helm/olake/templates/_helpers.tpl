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

{{/*
Generate job scheduling environment variables for an activity type
Usage: {{- include "olake.job.schedulingEnvVars" (dict "Values" .Values "activityName" "sync") }}
*/}}
{{- define "olake.job.schedulingEnvVars" -}}
{{- $config := index .Values.global.job .activityName | default dict -}}
{{- $activityUpper := upper .activityName -}}
OLAKE_{{ $activityUpper }}_JOB_NODE_SELECTOR: {{ $config.nodeSelector | toJson | quote }}
OLAKE_{{ $activityUpper }}_JOB_TOLERATIONS: {{ $config.tolerations | toJson | quote }}
OLAKE_{{ $activityUpper }}_JOB_ANTI_AFFINITY_ENABLED: {{ $config.antiAffinity.enabled | quote }}
OLAKE_{{ $activityUpper }}_JOB_ANTI_AFFINITY_STRATEGY: {{ $config.antiAffinity.strategy | quote }}
OLAKE_{{ $activityUpper }}_JOB_ANTI_AFFINITY_TOPOLOGY_KEY: {{ $config.antiAffinity.topologyKey | quote }}
OLAKE_{{ $activityUpper }}_JOB_ANTI_AFFINITY_WEIGHT: {{ $config.antiAffinity.weight | quote }}
{{- end -}}