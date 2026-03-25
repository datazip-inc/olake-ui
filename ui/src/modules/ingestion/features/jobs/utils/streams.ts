import semver from "semver"

import { MIN_COLUMN_SELECTION_SOURCE_VERSION } from "@/modules/ingestion/common/constants"
import {
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamData,
	SelectedStream,
} from "@/modules/ingestion/common/types"
import { normalizeConnectorType } from "@/modules/ingestion/common/utils"

import {
	DESTINATION_SUPPORTED_INGESTION_MODES,
	FILTER_REGEX,
	SOURCE_SUPPORTED_INGESTION_MODES,
	STREAM_DEFAULTS,
} from "../constants"
import { IngestionMode } from "../enums"
import { CursorFieldValues } from "../types"

/**
 * Processes the raw SourceStreamsResponse into the
 * StreamsDataStructure expected by the UI.
 */
export const getStreamsDataFromSourceStreamsResponse = (
	response: StreamsDataStructure,
	destinationType?: string,
	sourceType?: string,
	sourceVersion?: string,
): StreamsDataStructure => {
	const mergedSelectedStreams: SelectedStreamsByNamespace = {}

	const isDestUpsertModeSupported = isDestinationIngestionModeSupported(
		IngestionMode.UPSERT,
		destinationType,
	)

	const isSourceUpsertModeSupported = isSourceIngestionModeSupported(
		IngestionMode.UPSERT,
		sourceType,
	)

	// Column selection is supported from source version v0.4.0 onwards.
	const supportsColumnSelection =
		!!sourceVersion &&
		!!semver.valid(sourceVersion) &&
		semver.gte(sourceVersion, MIN_COLUMN_SELECTION_SOURCE_VERSION)

	// Iterate through all streams
	response.streams.forEach((stream: StreamData) => {
		const namespace = stream.stream.namespace || ""
		const streamName = stream.stream.name

		// Initialize namespace array if it doesn't exist
		if (!mergedSelectedStreams[namespace]) {
			mergedSelectedStreams[namespace] = []
		}

		// Check if this stream is in selected_streams
		const selectedNamespaceStreams =
			response.selected_streams?.[namespace] || []
		const matchingSelectedStream = selectedNamespaceStreams.find(
			s => s.stream_name === streamName,
		)

		if (matchingSelectedStream) {
			// Stream is selected, use the selected stream configuration
			mergedSelectedStreams[namespace].push({
				...matchingSelectedStream,
				disabled: false,
			})
		} else {
			// Stream is not selected, use defaults from default_stream_properties
			// Missing properties in default_stream_properties are treated as false/empty
			// Backward compatibility: fall back to hardcoded defaults if default_stream_properties is not present (older olake versions)
			const streamDefaults = stream.stream.default_stream_properties
			const defaults = {
				...STREAM_DEFAULTS,
				...streamDefaults,
			}

			mergedSelectedStreams[namespace].push({
				...defaults,
				stream_name: streamName,
				disabled: true,
				append_mode: !isDestUpsertModeSupported || !isSourceUpsertModeSupported, // Default to append if either source or destination does not support upsert
				// Add selected_columns only when the source supports it.
				...(supportsColumnSelection && {
					selected_columns: {
						columns: Object.keys(stream.stream.type_schema?.properties ?? {}),
						sync_new_columns: true,
					},
				}),
			})
		}
	})

	return {
		streams: response.streams,
		selected_streams: mergedSelectedStreams,
	}
}

// Returns true if the selected stream supports explicit column selection via the `selected_columns` field.
export function isColumnSelectionSupported(
	selectedStream: SelectedStream,
): boolean {
	return selectedStream.selected_columns !== undefined
}

// Returns true if the specified column is enabled for the selected stream.
// For legacy drivers, all columns are considered enabled by default.
export function isColumnEnabled(
	columnName: string,
	selectedStream: SelectedStream,
): boolean {
	if (!isColumnSelectionSupported(selectedStream)) return true
	return selectedStream.selected_columns!.columns.includes(columnName)
}

// Returns a copy of the selected streams map with all disabled streams removed
export const getSelectedStreams = (selectedStreams: {
	[key: string]: SelectedStream[]
}): { [key: string]: SelectedStream[] } => {
	return Object.fromEntries(
		Object.entries(selectedStreams).map(([key, streams]) => [
			key,
			streams.filter(stream => !stream.disabled),
		]),
	)
}

// validates filter expression
export const validateFilter = (filter: string): boolean => {
	if (!filter.trim()) return false
	return FILTER_REGEX.test(filter.trim())
}

export const validateStreams = (selections: {
	[key: string]: SelectedStream[]
}): boolean => {
	return !Object.values(selections).some(streams =>
		streams.some(sel => sel.filter && !validateFilter(sel.filter)),
	)
}

export const getIngestionMode = (
	selectedStreams: SelectedStreamsByNamespace,
	sourceType?: string,
): IngestionMode => {
	// Fallback to APPEND if source doesn't support UPSERT
	if (!isSourceIngestionModeSupported(IngestionMode.UPSERT, sourceType)) {
		return IngestionMode.APPEND
	}

	const selectedStreamsObj = getSelectedStreams(selectedStreams)
	const allSelectedStreams: SelectedStream[] = []

	// Flatten all streams from all namespaces
	Object.values(selectedStreamsObj).forEach((streams: SelectedStream[]) => {
		allSelectedStreams.push(...streams)
	})

	if (allSelectedStreams.length === 0) return IngestionMode.UPSERT

	const appendCount = allSelectedStreams.filter(
		s => s.append_mode === true,
	).length
	const upsertCount = allSelectedStreams.filter(s => !s.append_mode).length

	if (appendCount === allSelectedStreams.length) return IngestionMode.APPEND
	if (upsertCount === allSelectedStreams.length) return IngestionMode.UPSERT
	return IngestionMode.CUSTOM
}

// Checks if the source connector supports a specific ingestion mode
export const isSourceIngestionModeSupported = (
	mode: IngestionMode,
	sourceType?: string,
): boolean => {
	if (!sourceType) return false

	const normSourceType = normalizeConnectorType(
		sourceType,
	).toLowerCase() as keyof typeof SOURCE_SUPPORTED_INGESTION_MODES
	const sourceModes = SOURCE_SUPPORTED_INGESTION_MODES[normSourceType]

	return sourceModes?.some(m => m === mode) ?? false
}

// Checks if the destination connector supports a specific ingestion mode
export const isDestinationIngestionModeSupported = (
	mode: IngestionMode,
	destinationType?: string,
): boolean => {
	if (!destinationType) return false

	const normDestType = normalizeConnectorType(destinationType).toLowerCase()
	const destModes =
		DESTINATION_SUPPORTED_INGESTION_MODES[
			normDestType as keyof typeof DESTINATION_SUPPORTED_INGESTION_MODES
		]

	return destModes?.some(m => m === mode) ?? false
}

export const getCursorFieldValues = (
	cursorValue?: string,
): CursorFieldValues => {
	if (!cursorValue) {
		return {
			primary: "",
			fallback: "",
		}
	}

	const [primary, fallback] = cursorValue.split(":")

	return {
		primary,
		fallback: fallback || "",
	}
}
