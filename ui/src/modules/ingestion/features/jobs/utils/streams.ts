import semver from "semver"

import {
	MIN_COLUMN_SELECTION_SOURCE_VERSION,
	MIN_JSON_FILTER_VERSION,
	MIN_SOURCE_NAMING_CONVENTION_VERSION,
} from "@/modules/ingestion/common/constants"
import {
	CatalogFormat,
	DiscoverResponse,
	SelectedStreams,
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamData,
	SelectedStream,
	StreamsV2Catalog,
	SyncMode,
	StreamIdentifier,
	UpsertType,
} from "@/modules/ingestion/common/types"
import { normalizeConnectorType } from "@/modules/ingestion/common/utils"

import {
	DESTINATION_SUPPORTED_INGESTION_MODES,
	SOURCE_SUPPORTED_INGESTION_MODES,
	STREAM_DEFAULTS,
} from "../constants"
import { IngestionMode } from "../enums"
import {
	AdvancedSettings,
	StreamsCatalogPayload,
	CursorFieldValues,
	StreamDifferenceRequest,
} from "../types"
import {
	configuredDestinationDatabase,
	destinationDatabaseFor,
} from "./destination-database"
import {
	castFilterConditionValue,
	validateFilter,
	validateFilterConfig,
} from "./filterUtils"
import { fromLegacyCatalog, toLegacyStreamsConfig } from "./legacyStreams"

// The catalog carries the default upsert type on the stream as
// default_stream_properties.update_type; the selected stream stores it as
// update_type. Older olake versions omit it, in which case the backend applies
// its own default and the UI leaves update_type unset.
export const getDefaultUpsertType = (
	stream?: StreamData,
): UpsertType | undefined =>
	stream?.stream.default_stream_properties?.update_type

// Same lookup, from the streams list by stream identifier.
export const getDefaultUpsertTypeFor = (
	streams: StreamData[] | undefined,
	{ streamName, namespace }: StreamIdentifier,
): UpsertType | undefined =>
	getDefaultUpsertType(
		streams?.find(
			s => s.stream.name === streamName && s.stream.namespace === namespace,
		),
	)

/**
 * Processes the raw SourceStreamsResponse into the
 * StreamsDataStructure expected by the UI.
 */
export const getStreamsDataFromSourceStreamsResponse = (
	response: StreamsV2Catalog,
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

	// Computed once: streams without a database follow the one the selected entries use.
	const configuredDatabase = configuredDestinationDatabase(
		response.selected_streams ?? {},
	)

	// Iterate through all streams
	response.available_streams.streams.forEach((stream: StreamData) => {
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

		const streamDefaults = stream.stream.default_stream_properties
		const defaults = {
			...STREAM_DEFAULTS,
			...streamDefaults,
		}

		if (matchingSelectedStream) {
			// Stream is selected, use the selected stream configuration
			const appendMode =
				matchingSelectedStream.append_mode ?? defaults.append_mode
			const upsertMode =
				!appendMode && isDestUpsertModeSupported && isSourceUpsertModeSupported
			const syncMode =
				matchingSelectedStream.sync_mode ?? stream.stream.sync_mode
			const cursorField =
				matchingSelectedStream.cursor_field ?? stream.stream.cursor_field
			const { update_type: savedUpdateType, ...selectedStreamRest } =
				matchingSelectedStream

			mergedSelectedStreams[namespace].push({
				// Absent selected_columns means all columns; a saved value below wins.
				...(supportsColumnSelection && {
					selected_columns: {
						columns: Object.keys(stream.stream.type_schema?.properties ?? {}),
						sync_new_columns: true,
					},
				}),
				...selectedStreamRest,
				disabled: false,
				sync_mode: syncMode,
				cursor_field:
					syncMode === SyncMode.INCREMENTAL ? cursorField : undefined,
				destination_database:
					matchingSelectedStream.destination_database ??
					destinationDatabaseFor(stream.stream, configuredDatabase),
				append_mode: appendMode,
				normalization:
					matchingSelectedStream.normalization ?? defaults.normalization,
				// update_type only applies while the stream runs in upsert mode;
				// older saved jobs carry no value, so fall back to the catalog default.
				...(upsertMode && {
					update_type: savedUpdateType ?? getDefaultUpsertType(stream),
				}),
			})
		} else {
			// Stream is not selected, use defaults from default_stream_properties
			// Missing properties in default_stream_properties are treated as false/empty
			// Backward compatibility: fall back to hardcoded defaults if default_stream_properties is not present (older olake versions)
			mergedSelectedStreams[namespace].push({
				...defaults,
				stream_name: streamName,
				disabled: true,
				sync_mode: stream.stream.sync_mode,
				cursor_field:
					stream.stream.sync_mode === SyncMode.INCREMENTAL
						? stream.stream.cursor_field
						: undefined,
				destination_database: destinationDatabaseFor(
					stream.stream,
					configuredDatabase,
				),
				append_mode: !isDestUpsertModeSupported || !isSourceUpsertModeSupported, // Default to append if either source or destination does not support upsert
				// update_type only applies while the stream runs in upsert mode.
				...(isDestUpsertModeSupported &&
					isSourceUpsertModeSupported && {
						update_type: getDefaultUpsertType(stream),
					}),
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
		streams: response.available_streams.streams,
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

// Returns true if the source version supports the use_source_column_names flag.
export function isUseSourceColumnNamesSupported(
	sourceVersion?: string,
): boolean {
	return (
		!!sourceVersion &&
		!!semver.valid(sourceVersion) &&
		semver.gte(sourceVersion, MIN_SOURCE_NAMING_CONVENTION_VERSION)
	)
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
			`${s.stream.namespace || ""}.${s.stream.name}`,
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

// The server returns available_streams only when the source (new job) or the
// job (existing job) is on streams v2; the frontend needs no version check.
export const parseDiscoverResponse = (
	response: DiscoverResponse,
): { format: CatalogFormat; response: StreamsV2Catalog } =>
	"available_streams" in response
		? { format: CatalogFormat.V2, response }
		: { format: CatalogFormat.LEGACY, response: fromLegacyCatalog(response) }

// streams[] is never edited, so it is the discover output as-is.
export const buildCatalogPayload = (
	streamsData: StreamsDataStructure,
	format: CatalogFormat,
): StreamsCatalogPayload => {
	const selectedStreams = formatSelectedStreamsPayload(streamsData)
	if (format === CatalogFormat.LEGACY) {
		return {
			streams_config: toLegacyStreamsConfig(streamsData, selectedStreams),
		}
	}
	return {
		available_streams_config: JSON.stringify({ streams: streamsData.streams }),
		selected_streams_config: JSON.stringify({
			selected_streams: selectedStreams,
		}),
	}
}

// Narrows a catalog payload to its v2 member (same rule as isStreamsV2Job for a job).
const isStreamsV2Payload = (
	payload: StreamsCatalogPayload,
): payload is {
	available_streams_config: string
	selected_streams_config: string
} => "available_streams_config" in payload

export const buildStreamDifferenceRequest = (
	streamsData: StreamsDataStructure,
	format: CatalogFormat,
): StreamDifferenceRequest => {
	const payload = buildCatalogPayload(streamsData, format)
	return isStreamsV2Payload(payload)
		? {
				updated_available_streams_config: payload.available_streams_config,
				updated_selected_streams_config: payload.selected_streams_config,
			}
		: { updated_streams_config: payload.streams_config }
}

// Positional deletes and delete vectors need the destination row index; equality deletes don't.
const INDEXED_UPSERT_TYPES: UpsertType[] = [
	UpsertType.POSITIONAL,
	UpsertType.DELETION_VECTOR,
]

const usesIndexedUpsert = (stream: SelectedStream): boolean =>
	!stream.append_mode &&
	!!stream.update_type &&
	INDEXED_UPSERT_TYPES.includes(stream.update_type)

const indexedStreamIds = (
	streamsConfig?: SelectedStreams | null,
): Set<string> =>
	new Set(
		Object.entries(
			getSelectedStreams(streamsConfig?.selected_streams ?? {}),
		).flatMap(([namespace, streams]) =>
			streams
				.filter(usesIndexedUpsert)
				.map(stream => `${namespace}.${stream.stream_name}`),
		),
	)

export const hasIndexedUpsertStream = (
	streamsConfig?: SelectedStreams | null,
): boolean => indexedStreamIds(streamsConfig).size > 0

// True when the difference covers every enabled stream, meaning clear destination
// runs across the whole job rather than a subset of its streams.
export const coversAllSelectedStreams = (
	streamsConfig: StreamsDataStructure | null | undefined,
	streamDifference: SelectedStreams,
): boolean => {
	const streamIds = (selectedStreams: SelectedStreamsByNamespace) =>
		new Set(
			Object.entries(getSelectedStreams(selectedStreams)).flatMap(
				([namespace, streams]) =>
					streams.map(stream => `${namespace}.${stream.stream_name}`),
			),
		)

	const selected = streamIds(streamsConfig?.selected_streams ?? {})
	const impacted = streamIds(streamDifference.selected_streams ?? {})

	return selected.size > 0 && [...selected].every(id => impacted.has(id))
}

// Engines only reach the catalog through a discover, so a selection that differs
// from the saved one has to go through the streams step before it can be saved.
export const queryEnginesChanged = (
	advancedSettings: AdvancedSettings | null | undefined,
	savedAdvancedSettings: AdvancedSettings | null | undefined,
): boolean => {
	const selected = advancedSettings?.target_query_engines ?? []
	const saved = savedAdvancedSettings?.target_query_engines ?? []

	return (
		selected.length !== saved.length ||
		selected.some(engine => !saved.includes(engine))
	)
}

// The next sync builds the destination index for any pos/dv stream that wasn't
// already pos/dv in the saved config. Returns true when a stream is:
//   - newly selected with pos/dv (not in the saved config, or disabled there)
//   - switched from eq to pos/dv
//   - switched from append mode to pos/dv
// Returns false when every pos/dv stream was already pos/dv in the saved config,
// so no new index is needed. With no saved config (job creation), every pos/dv
// stream counts as new.
export const willBuildIndex = (
	streamsConfig: StreamsDataStructure | null | undefined,
	savedStreamsConfig?: SelectedStreams | null,
): boolean => {
	const alreadyIndexed = indexedStreamIds(savedStreamsConfig)
	const nowIndexed = indexedStreamIds(streamsConfig)

	return [...nowIndexed].some(id => !alreadyIndexed.has(id))
}

// index_required is derived from the streams config on every write; the rest of
// advanced settings stays user-configured.
export const withIndexRequired = (
	advancedSettings: AdvancedSettings | null | undefined,
	streamsConfig?: SelectedStreams | null,
): AdvancedSettings => ({
	...advancedSettings,
	index_required: hasIndexedUpsertStream(streamsConfig),
})

// Returns null if all selected stream configurations are valid, or a descriptive error string otherwise.
export const validateStreams = (
	streamsConfig: StreamsDataStructure,
): string | null => {
	// Map typeSchemaProperties by stream name for quick lookup
	const typeSchemaByName = new Map(
		streamsConfig.streams?.map(s => [
			`${s.stream.namespace || ""}.${s.stream.name}`,
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
					{ streamName: sel.stream_name, namespace },
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
// used as the basis for bulk editing. Catalog metadata only; the configured
// values (sync_mode etc.) come from buildBulkSelectedStreams.
//
// Intersection rules:
// - type_schema columns: only columns present in every stream with identical types
// - available_cursor_fields: intersection across all streams, filtered to intersected columns only
// - source_defined_primary_key: intersection across all streams
// - supported_sync_mode: taken from the first selected stream
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

	return {
		stream: {
			name: "",
			namespace: "",
			json_schema: {},
			type_schema: { properties: intersectedProperties },
			available_cursor_fields: intersectedCursors,
			source_defined_primary_key: intersectedPks,
			supported_sync_modes: streams[0].stream.supported_sync_modes || [],
			default_stream_properties: streams[0].stream.default_stream_properties,
		},
	}
}

// Builds the default SelectedStream for a bulk edit session, seeded from the
// first selected stream's entry (the configured values live there, not on the catalog).
// Returns EMPTY_BULK_STREAM_DEFAULTS when no valid stream is provided.
export const buildBulkSelectedStreams = (
	commonStream: StreamData,
	firstSelected?: SelectedStream,
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

	const appendMode = !isDestUpsertModeSupported || !isSourceUpsertModeSupported

	const supportedModes = commonStream.stream.supported_sync_modes || []
	const noCursorFields =
		(commonStream.stream.available_cursor_fields || []).length === 0
	const noCdcSupport =
		!supportedModes.includes(SyncMode.CDC) &&
		!supportedModes.includes(SyncMode.STRICT_CDC)
	// Edge case: no intersected cursors disables incremental, and if CDC/strict_cdc
	// are also unsupported, only full_refresh is enabled. If full_refresh is also
	// the default, the user can't interact with the radio group at all — so SyncMode
	// never gets marked dirty and can't be bulk applied. Setting to undefined
	// leaves no radio pre-selected, so the user must explicitly pick full_refresh,
	// ensuring sync mode is only bulk applied when the user explicitly changes it.
	const syncMode =
		noCursorFields && noCdcSupport ? undefined : firstSelected?.sync_mode

	// update_type is the catalog's key for the default upsert type; it is carried
	// on the selected stream as update_type instead.
	const { update_type: defaultUpsertType, ...defaultProperties } =
		commonStream.stream.default_stream_properties ?? {}

	return {
		...STREAM_DEFAULTS,
		...defaultProperties,
		stream_name: commonStream.stream.name,
		sync_mode: syncMode,
		append_mode: appendMode,
		...(!appendMode && { update_type: defaultUpsertType }),
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
	const first = selectedStreamsInput[0]
	const firstSelected = first
		? streamsData?.selected_streams[first.namespace]?.find(
				s => s.stream_name === first.streamName,
			)
		: undefined
	const defaults = buildBulkSelectedStreams(
		stream,
		firstSelected,
		sourceType,
		destinationType,
	)
	return { stream, defaults }
}
