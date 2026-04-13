import semver from "semver"

import {
	MIN_COLUMN_SELECTION_SOURCE_VERSION,
	MIN_JSON_FILTER_VERSION,
} from "@/modules/ingestion/common/constants"
import {
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamData,
	SelectedStream,
	SyncMode,
	StreamIdentifier,
} from "@/modules/ingestion/common/types"
import { normalizeConnectorType } from "@/modules/ingestion/common/utils"

import {
	DESTINATION_SUPPORTED_INGESTION_MODES,
	SOURCE_SUPPORTED_INGESTION_MODES,
	STREAM_DEFAULTS,
} from "../constants"
import { IngestionMode } from "../enums"
import { CursorFieldValues } from "../types"
import {
	castFilterConditionValue,
	validateFilter,
	validateFilterConfig,
} from "./filterUtils"

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

// Filters out disabled streams
const getSelectedStreams = (
	selectedStreams: SelectedStreamsByNamespace,
): SelectedStreamsByNamespace => {
	const result: SelectedStreamsByNamespace = {}

	Object.keys(selectedStreams).forEach(key => {
		result[key] = selectedStreams[key].filter(stream => !stream.disabled)
	})

	return result
}

// Formats the selected streams configuration for the API payload
export const formatSelectedStreamsPayload = (
	streamsConfig: StreamsDataStructure,
): SelectedStreamsByNamespace => {
	const filteredStreams = getSelectedStreams(streamsConfig.selected_streams)

	const typeSchemaByName = new Map(
		streamsConfig.streams?.map(s => [
			`${s.stream.namespace}.${s.stream.name}`,
			s.stream.type_schema?.properties,
		]) ?? [],
	)

	return Object.fromEntries(
		Object.entries(filteredStreams).map(([namespace, namespaceStreams]) => [
			namespace,
			namespaceStreams.map(stream => {
				const typeSchemaProps = typeSchemaByName.get(
					`${namespace}.${stream.stream_name}`,
				)
				if (!stream.filter_config || !typeSchemaProps) return stream

				return {
					...stream,
					// Cast each condition's value to its schema-defined native type
					filter_config: {
						...stream.filter_config,
						conditions: stream.filter_config.conditions.map(cond =>
							castFilterConditionValue(cond, typeSchemaProps[cond.column]),
						),
					},
				}
			}),
		]),
	)
}

// Returns null if all selected stream configurations are valid, or a descriptive error string otherwise.
export const validateStreams = (
	streamsConfig: StreamsDataStructure,
): string | null => {
	// Map typeSchemaProperties by stream name for quick lookup
	const typeSchemaByName = new Map(
		streamsConfig.streams?.map(s => [
			`${s.stream.namespace}.${s.stream.name}`,
			s.stream.type_schema?.properties,
		]) ?? [],
	)

	const selectedStreams = getSelectedStreams(streamsConfig.selected_streams)

	for (const [namespace, nsStreams] of Object.entries(selectedStreams)) {
		for (const sel of nsStreams) {
			if (sel.filter && !validateFilter(sel.filter)) {
				return `[${namespace ? `${namespace}.` : ""}${sel.stream_name}] Invalid filter expression`
			}
			if (sel.filter_config) {
				const typeSchemaProps = typeSchemaByName.get(
					`${namespace}.${sel.stream_name}`,
				)
				const error = validateFilterConfig(
					sel.filter_config,
					sel.stream_name,
					namespace,
					typeSchemaProps,
				)
				if (error) return error
			}
		}
	}

	return null
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

// Returns true if filter_config (JSON) should be used instead of the legacy filter string.
// Requires source >= v0.6.0 AND no selected stream already carries a non-empty legacy filter.
export function shouldUseFilterConfig(
	selectedStreams: SelectedStreamsByNamespace,
	sourceVersion: string,
): boolean {
	if (!sourceVersion || !semver.valid(sourceVersion)) return false
	if (!semver.gte(sourceVersion, MIN_JSON_FILTER_VERSION)) return false

	// If ANY stream already carries a legacy filter string, keep legacy path.
	return !Object.values(selectedStreams).some(streams =>
		streams.some(s => typeof s.filter === "string" && s.filter.trim() !== ""),
	)
}

// Returns true when grouped stream namespaces or namespace stream counts change.
export const hasGroupedStreamsStructureChanged = (
	prev: Record<string, StreamData[]>,
	current: Record<string, StreamData[]>,
): boolean => {
	const prevKeys = Object.keys(prev)
	const currentKeys = Object.keys(current)

	if (prevKeys.length !== currentKeys.length) return true

	for (const key of currentKeys) {
		if (!prev[key]) return true
		if (prev[key].length !== current[key].length) return true
	}

	return false
}

// Sorts grouped streams by checked-first order while preserving alphabetical order within buckets.
export const sortGroupedStreamsByCheckedState = (
	groupedStreams: Record<string, StreamData[]>,
	checkedStreamsByNamespace: {
		[ns: string]: { [streamName: string]: boolean }
	},
): [string, StreamData[]][] => {
	const sortByStreamName = (a: StreamData, b: StreamData) =>
		a.stream.name.localeCompare(b.stream.name)
	const sortByNamespaceName = (
		a: [string, StreamData[]],
		b: [string, StreamData[]],
	) => a[0].localeCompare(b[0])

	const withChecked: [string, StreamData[]][] = []
	const withoutChecked: [string, StreamData[]][] = []

	Object.entries(groupedStreams).forEach(([ns, streams]) => {
		const checked: StreamData[] = []
		const unchecked: StreamData[] = []

		streams.forEach(stream => {
			if (checkedStreamsByNamespace[ns]?.[stream.stream.name]) {
				checked.push(stream)
			} else {
				unchecked.push(stream)
			}
		})

		checked.sort(sortByStreamName)
		unchecked.sort(sortByStreamName)
		const sortedNamespace: [string, StreamData[]] = [
			ns,
			[...checked, ...unchecked],
		]

		if (checked.length > 0) {
			withChecked.push(sortedNamespace)
		} else {
			withoutChecked.push(sortedNamespace)
		}
	})

	withChecked.sort(sortByNamespaceName)
	withoutChecked.sort(sortByNamespaceName)
	return [...withChecked, ...withoutChecked]
}

const EMPTY_BULK_STREAM: StreamData = {
	sync_mode: SyncMode.FULL_REFRESH,
	destination_sync_mode: "",
	sort_key: null,
	stream: {
		name: "",
		namespace: "",
		json_schema: {},
		type_schema: { properties: {} },
		available_cursor_fields: [],
		source_defined_primary_key: [],
		supported_sync_modes: [],
		default_stream_properties: {
			normalization: false,
			append_mode: false,
		},
	},
}

// Intersection of string lists across streams: start from the first stream’s array, then keep only
// entries that also appear in every later stream’s array (order follows the first stream).
const intersectArrays = (
	streams: StreamData[],
	getArr: (s: StreamData) => string[] | undefined,
): string[] =>
	streams.reduce<string[]>((acc, s, index) => {
		const arr = getArr(s) || []
		return index === 0 ? [...arr] : acc.filter(item => arr.includes(item))
	}, [])

// Builds a StreamData representing the intersection of all selected streams,
// used as the basis for bulk editing.
//
// Intersection rules:
// - type_schema columns: only columns present in every stream with identical types
// - available_cursor_fields: intersection across all streams, filtered to intersected columns only
// - source_defined_primary_key: intersection across all streams
// - supported_sync_mode: taken from the first selected strea
// - sync_mode: taken from the first selected stream
// - default_stream_properties: taken from the first selected stream
//
// Returns EMPTY_BULK_STREAM when no valid streams are selected.
export const buildBulkCommonStream = (
	selectedStreamsInput: StreamIdentifier[],
	streamsData: StreamsDataStructure | null,
): StreamData => {
	if (!streamsData || selectedStreamsInput.length === 0) {
		return EMPTY_BULK_STREAM
	}

	const streams = selectedStreamsInput
		.map(({ streamName, namespace }) =>
			streamsData.streams.find(
				s =>
					s.stream.name === streamName &&
					(s.stream.namespace || "") === namespace,
			),
		)
		.filter((s): s is StreamData => s !== undefined)

	if (streams.length === 0) {
		return EMPTY_BULK_STREAM
	}

	const intersectedProperties = streams.reduce<Record<string, any>>(
		(acc, s, index) => {
			const props = s.stream.type_schema?.properties || {}
			if (index === 0) {
				return Object.fromEntries(
					Object.entries(props).map(([key, value]) => [key, { ...value }]),
				)
			}
			return Object.fromEntries(
				Object.entries(acc).filter(([key]) => {
					if (!props[key]) return false
					const typeA = JSON.stringify([...acc[key].type].sort())
					const typeB = JSON.stringify([...props[key].type].sort())
					return typeA === typeB
				}),
			)
		},
		{},
	)

	const intersectedCursors = intersectArrays(
		streams,
		s => s.stream.available_cursor_fields,
	).filter(c => intersectedProperties[c])

	const intersectedPks = intersectArrays(
		streams,
		s => s.stream.source_defined_primary_key,
	)

	const supportedModes = streams[0].stream.supported_sync_modes || []
	const rawSyncMode =
		(streams[0].stream.sync_mode as SyncMode) ?? SyncMode.FULL_REFRESH
	// Fall back to full_refresh if incremental has no intersected cursor fields.
	const commonSyncMode =
		rawSyncMode === SyncMode.INCREMENTAL && intersectedCursors.length === 0
			? SyncMode.FULL_REFRESH
			: rawSyncMode

	return {
		sync_mode: commonSyncMode,
		destination_sync_mode: "",
		sort_key: null,
		stream: {
			name: "",
			namespace: "",
			json_schema: {},
			type_schema: { properties: intersectedProperties },
			available_cursor_fields: intersectedCursors,
			source_defined_primary_key: intersectedPks,
			supported_sync_modes: supportedModes,
			default_stream_properties: streams[0].stream.default_stream_properties,
		},
	}
}

// Builds the default SelectedStream for a bulk edit session
// Returns EMPTY_BULK_STREAM_DEFAULTS when no valid stream is provided.
export const buildBulkSelectedStreams = (
	commonStream: StreamData,
	sourceType?: string,
	destinationType?: string,
): SelectedStream => {
	const isDestUpsertModeSupported = isDestinationIngestionModeSupported(
		IngestionMode.UPSERT,
		destinationType,
	)
	const isSourceUpsertModeSupported = isSourceIngestionModeSupported(
		IngestionMode.UPSERT,
		sourceType,
	)

	return {
		...STREAM_DEFAULTS,
		...commonStream.stream.default_stream_properties,
		stream_name: commonStream.stream.name,
		append_mode: !isDestUpsertModeSupported || !isSourceUpsertModeSupported,
	}
}
// Returns the stream data and default selected stream data
export const buildBulkStreamsData = (
	selectedStreamsInput: StreamIdentifier[],
	streamsData: StreamsDataStructure | null,
	sourceType?: string,
	destinationType?: string,
): { stream: StreamData; defaults: SelectedStream } => {
	const stream = buildBulkCommonStream(selectedStreamsInput, streamsData)
	const defaults = buildBulkSelectedStreams(stream, sourceType, destinationType)
	return { stream, defaults }
}
