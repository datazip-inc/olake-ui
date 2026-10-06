import { create } from "zustand"

import {
	CatalogFormat,
	StreamsDataStructure,
	StreamData,
	SelectedStream,
	SelectedStreamsByNamespace,
	SelectedColumns,
	StreamIdentifier,
	SyncMode,
	FilterConfig,
	UpsertType,
} from "@/modules/ingestion/common/types"

import { extractNamespaceFromDestination } from "../utils"
import { getDefaultUpsertTypeFor } from "../utils/streams"

export interface BulkStreamConfig {
	syncMode?: SyncMode
	cursorField?: string
	appendMode?: boolean
	upsertType?: UpsertType
	normalization?: boolean
	partitionRegex?: string
	filterValue?: string
	filterConfig?: FilterConfig
	useSourceColumnNames?: boolean
}

interface StreamSelectionState {
	streamsData: StreamsDataStructure | null

	// Format the discover response came in; decides the save payload shape.
	catalogFormat: CatalogFormat

	// Frozen snapshot from initial discovery; used by DestinationDatabaseModal.
	initialStreamsSnapshot: StreamsDataStructure | null
	isDiscovering: boolean
	discoverError: string | null
	activeStreamKey: StreamIdentifier | null

	// Per-stream filter toggle keyed by `${namespace}_${name}`.
	streamFilterStates: Record<string, boolean>

	// Whether to use structured filter_config (new) vs legacy filter string.
	useFilterConfig: boolean

	// Incremented each time bulkUpdateStreams successfully applies changes.
	// used to trigger a re-sort only after bulk apply.
	bulkApplyVersion: number

	initializeFromDiscovery: (
		data: StreamsDataStructure,
		format: CatalogFormat,
	) => void
	setDiscovering: (loading: boolean) => void
	setDiscoverError: (message: string | null) => void

	// Toggles a stream on (disabled:false) or off (disabled:true).
	toggleStream: (stream: StreamIdentifier, checked: boolean) => void

	// Updates sync_mode and optional cursor_field.
	updateSyncMode: (
		stream: StreamIdentifier,
		syncMode: SyncMode,
		cursorField?: string,
	) => void

	updateNormalization: (
		stream: StreamIdentifier,
		normalization: boolean,
	) => void

	updateUseSourceColumnNames: (
		stream: StreamIdentifier,
		useSourceColumnNames: boolean,
	) => void

	updatePartitionRegex: (stream: StreamIdentifier, regex: string) => void

	// Empty string removes the `filter` key entirely.
	updateFilter: (stream: StreamIdentifier, filterValue: string) => void

	updateFilterConfig: (
		stream: StreamIdentifier,
		filterConfig: FilterConfig | undefined,
	) => void

	setUseFilterConfig: (value: boolean) => void

	updateIngestionMode: (stream: StreamIdentifier, appendMode: boolean) => void

	// Only meaningful for streams in upsert mode (append_mode falsy).
	updateUpsertType: (stream: StreamIdentifier, upsertType: UpsertType) => void

	bulkUpdateStreams: (
		streamsToUpdate: StreamIdentifier[],
		config: BulkStreamConfig,
	) => void

	// Applies append_mode to every stream in selected_streams.
	updateAllIngestionMode: (appendMode: boolean) => void

	// Updates destination_database on all streams.
	updateDestinationDatabase: (format: string, databaseName: string) => void

	updateSelectedColumns: (
		stream: StreamIdentifier,
		columns: SelectedColumns,
	) => void

	setStreamFilterState: (streamKey: string, value: boolean) => void
	setActiveStreamKey: (key: StreamIdentifier | null) => void
	reset: () => void
}

const initialState = {
	streamsData: null,
	catalogFormat: CatalogFormat.LEGACY,
	initialStreamsSnapshot: null,
	isDiscovering: false,
	discoverError: null,
	activeStreamKey: null,
	streamFilterStates: {} as Record<string, boolean>,
	useFilterConfig: false,
	bulkApplyVersion: 0,
}

// update_type is only carried by streams running in upsert mode.
// Mutates and returns the same object so callers can use it inline.
const withUpsertTypeSynced = (
	stream: SelectedStream,
	defaultUpsertType?: UpsertType,
): SelectedStream => {
	if (stream.append_mode) {
		delete stream.update_type
	} else if (!stream.update_type && defaultUpsertType) {
		stream.update_type = defaultUpsertType
	}
	return stream
}

export const useStreamSelectionStore = create<StreamSelectionState>()(set => ({
	...initialState,

	initializeFromDiscovery: (data, format) =>
		set(state => ({
			streamsData: data,
			catalogFormat: format,
			initialStreamsSnapshot: state.initialStreamsSnapshot ?? data,
			isDiscovering: false,
			discoverError: null,
		})),

	setDiscovering: loading => set({ isDiscovering: loading }),
	setDiscoverError: error => set({ discoverError: error }),

	toggleStream: (stream, checked) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const updated = {
				...prev,
				selected_streams: { ...prev.selected_streams },
			}
			let changed = false

			const existingStream = updated.selected_streams[namespace]?.find(
				s => s.stream_name === streamName,
			)

			if (checked) {
				if (!updated.selected_streams[namespace]) {
					updated.selected_streams[namespace] = []
				}
				// No insert branch: getStreamsDataFromSourceStreamsResponse creates an
				// entry (disabled when not selected) for every discovered stream, so a
				// missing entry cannot happen; toggling only flips `disabled`.
				if (existingStream?.disabled) {
					updated.selected_streams[namespace] = updated.selected_streams[
						namespace
					].map(s =>
						s.stream_name === streamName ? { ...s, disabled: false } : s,
					)
					changed = true
				}
			} else {
				if (existingStream && !existingStream.disabled) {
					updated.selected_streams[namespace] = updated.selected_streams[
						namespace
					].map(s =>
						s.stream_name === streamName ? { ...s, disabled: true } : s,
					)
					changed = true
				}
			}

			return changed ? { streamsData: updated } : state
		}),

	updateSyncMode: (stream, newSyncMode, cursorField) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s => {
							if (s.stream_name !== streamName) return s
							const updated = { ...s, sync_mode: newSyncMode }
							if (newSyncMode !== SyncMode.INCREMENTAL) {
								delete updated.cursor_field
							} else if (cursorField !== undefined) {
								updated.cursor_field = cursorField
							}
							return updated
						}),
					},
				},
			}
		}),

	updateNormalization: (stream, normalization) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s =>
							s.stream_name === streamName ? { ...s, normalization } : s,
						),
					},
				},
			}
		}),

	updateUseSourceColumnNames: (stream, useSourceColumnNames) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s =>
							s.stream_name === streamName
								? { ...s, use_source_column_names: useSourceColumnNames }
								: s,
						),
					},
				},
			}
		}),

	updatePartitionRegex: (stream, regex) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s =>
							s.stream_name === streamName
								? { ...s, partition_regex: regex }
								: s,
						),
					},
				},
			}
		}),

	updateFilter: (stream, filterValue) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s => {
							if (s.stream_name !== streamName) return s
							if (filterValue === "") {
								const updated = { ...s }
								delete updated.filter
								return updated
							}
							return { ...s, filter: filterValue }
						}),
					},
				},
			}
		}),

	updateFilterConfig: (stream, filterConfig) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s => {
							if (s.stream_name !== streamName) return s
							if (filterConfig === undefined) {
								// Remove filter_config when filter is disabled
								const updated = { ...s }
								delete updated.filter_config
								return updated
							}
							return { ...s, filter_config: filterConfig }
						}),
					},
				},
			}
		}),

	setUseFilterConfig: value => set({ useFilterConfig: value }),

	updateIngestionMode: (stream, appendMode) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s =>
							s.stream_name === streamName
								? withUpsertTypeSynced(
										{ ...s, append_mode: appendMode },
										getDefaultUpsertTypeFor(prev.streams, stream),
									)
								: s,
						),
					},
				},
			}
		}),

	updateUpsertType: (stream, upsertType) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			const streamExists = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)
			if (!streamExists) return state

			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace].map(s =>
							s.stream_name === streamName && !s.append_mode
								? { ...s, update_type: upsertType }
								: s,
						),
					},
				},
			}
		}),

	bulkUpdateStreams: (streamsToUpdate, config) =>
		set(state => {
			if (!state.streamsData) return state

			const prev = state.streamsData
			const updatedSelected = { ...prev.selected_streams }
			const updatedFilterStates = { ...state.streamFilterStates }
			let changed = false

			streamsToUpdate.forEach(({ streamName, namespace }) => {
				const streamList = updatedSelected[namespace] || []
				const streamIndex = streamList.findIndex(
					s => s.stream_name === streamName,
				)

				if (streamIndex !== -1) {
					const newStream = { ...streamList[streamIndex] }

					if (config.syncMode !== undefined) {
						newStream.sync_mode = config.syncMode
						if (config.syncMode !== SyncMode.INCREMENTAL) {
							delete newStream.cursor_field
						} else if (config.cursorField !== undefined) {
							newStream.cursor_field = config.cursorField
						}
					}
					if (config.appendMode !== undefined)
						newStream.append_mode = config.appendMode
					if (config.upsertType !== undefined)
						newStream.update_type = config.upsertType
					if (
						config.appendMode !== undefined ||
						config.upsertType !== undefined
					)
						withUpsertTypeSynced(
							newStream,
							getDefaultUpsertTypeFor(prev.streams, {
								streamName,
								namespace,
							}),
						)
					if (config.normalization !== undefined)
						newStream.normalization = config.normalization
					if (config.partitionRegex !== undefined)
						newStream.partition_regex = config.partitionRegex
					if (config.useSourceColumnNames !== undefined)
						newStream.use_source_column_names = config.useSourceColumnNames
					newStream.disabled = false

					// Only touch filter fields if filter was explicitly included in the config
					if (
						config.filterValue !== undefined ||
						config.filterConfig !== undefined
					) {
						// Clear both first to ensure only one is ever set at a time
						delete newStream.filter
						delete newStream.filter_config

						if (config.filterConfig) {
							newStream.filter_config = config.filterConfig
						} else if (config.filterValue) {
							newStream.filter = config.filterValue
						}

						const streamKey = `${namespace}_${streamName}`
						updatedFilterStates[streamKey] = !!(
							config.filterConfig || config.filterValue
						)
					}

					updatedSelected[namespace] = [
						...streamList.slice(0, streamIndex),
						newStream,
						...streamList.slice(streamIndex + 1),
					]
					changed = true
				}
			})

			return changed
				? {
						streamsData: {
							...prev,
							selected_streams: updatedSelected,
						},
						streamFilterStates: updatedFilterStates,
						bulkApplyVersion: state.bulkApplyVersion + 1,
					}
				: state
		}),

	updateAllIngestionMode: appendMode =>
		set(state => {
			if (!state.streamsData) return state

			const prev = state.streamsData
			const updatedSelected = Object.fromEntries(
				Object.entries(prev.selected_streams).map(([ns, streams]) => [
					ns,
					streams.map(s =>
						withUpsertTypeSynced(
							{ ...s, append_mode: appendMode },
							getDefaultUpsertTypeFor(prev.streams, {
								streamName: s.stream_name,
								namespace: ns,
							}),
						),
					),
				]),
			)

			return {
				streamsData: { ...prev, selected_streams: updatedSelected },
			}
		}),

	updateDestinationDatabase: (format, databaseName) =>
		set(state => {
			if (!state.streamsData || state.streamsData.streams.length === 0) {
				return state
			}

			const prev = state.streamsData
			const firstStreamDestDb = Object.values(prev.selected_streams).flat()[0]
				?.destination_database
			const hasColonFormat =
				firstStreamDestDb && firstStreamDestDb.includes(":")

			// Applies to every entry, deselected ones included, so a stream enabled
			// later lands in the chosen database.
			const updatedSelected = Object.fromEntries(
				Object.entries(prev.selected_streams).map(([ns, streams]) => [
					ns,
					streams.map(stream => {
						const currentDestDb = stream.destination_database

						if (format === "dynamic") {
							if (hasColonFormat && currentDestDb) {
								// "a:b" → "databaseName:b"
								const parts = currentDestDb.split(":")
								return {
									...stream,
									destination_database: `${databaseName}:${parts[1]}`,
								}
							} else {
								// No colon — derive namespace from initialStreamsSnapshot
								const initialStream =
									state.initialStreamsSnapshot?.streams.find(
										s =>
											s.stream.name === stream.stream_name &&
											(s.stream.namespace || "") === ns,
									)
								const namespace = extractNamespaceFromDestination(
									initialStream?.stream.destination_database,
									ns,
								)
								return {
									...stream,
									destination_database: `${databaseName}:${namespace}`,
								}
							}
						} else {
							return { ...stream, destination_database: databaseName }
						}
					}),
				]),
			)

			return {
				streamsData: { ...prev, selected_streams: updatedSelected },
			}
		}),

	updateSelectedColumns: (stream, columns) =>
		set(state => {
			if (!state.streamsData) return state
			const { streamName, namespace } = stream

			const prev = state.streamsData
			return {
				streamsData: {
					...prev,
					selected_streams: {
						...prev.selected_streams,
						[namespace]: prev.selected_streams[namespace]?.map(s =>
							s.stream_name === streamName
								? { ...s, selected_columns: columns }
								: s,
						),
					},
				},
			}
		}),

	setStreamFilterState: (streamKey, value) =>
		set(state => ({
			streamFilterStates: {
				...state.streamFilterStates,
				[streamKey]: value,
			},
		})),

	setActiveStreamKey: key => set({ activeStreamKey: key }),

	reset: () => set(initialState),
}))

// Narrow selectors for optimized subscriptions (avoid full-store re-renders).
export const selectSelectedStreams = (
	state: StreamSelectionState,
): SelectedStreamsByNamespace => state.streamsData?.selected_streams ?? {}
export const selectStreamsData = (state: StreamSelectionState) =>
	state.streamsData
export const selectIsDiscovering = (state: StreamSelectionState) =>
	state.isDiscovering
// Delete formats the catalog's target query engines can read. Discover computes
// this once and writes the identical list to every stream, so reading it off any
// one stream is enough — no need to scan or intersect the rest.
// Undefined on catalogs discovered before target query engines existed.
export const selectAvailableUpdateTypes = (state: StreamSelectionState) =>
	state.streamsData?.streams?.[0]?.stream.available_update_types
export const selectInitialStreamsSnapshot = (state: StreamSelectionState) =>
	state.initialStreamsSnapshot
export const selectActiveStreamKey = (state: StreamSelectionState) =>
	state.activeStreamKey
export const selectStreamFilterState =
	(streamKey: string) => (state: StreamSelectionState) =>
		state.streamFilterStates[streamKey] ?? false

// Returns the StreamData entry for the currently active stream.
export const selectActiveStreamData = (
	state: StreamSelectionState,
): StreamData | null => {
	if (!state.activeStreamKey || !state.streamsData?.streams) return null
	return (
		state.streamsData.streams.find(
			s =>
				s.stream.namespace === state.activeStreamKey!.namespace &&
				s.stream.name === state.activeStreamKey!.streamName,
		) ?? null
	)
}

// Returns the SelectedStream entry for the currently active stream.
export const selectActiveSelectedStream = (
	state: StreamSelectionState,
): SelectedStream | null => {
	if (!state.activeStreamKey || !state.streamsData?.selected_streams)
		return null
	return (
		state.streamsData.selected_streams[state.activeStreamKey.namespace]?.find(
			s => s.stream_name === state.activeStreamKey!.streamName,
		) ?? null
	)
}

// Derives destination database display values from the first stream entry.
export const selectDestinationDatabase = (
	state: StreamSelectionState,
): { display: string | null; forModal: string | null } => {
	const firstStream = Object.values(
		state.streamsData?.selected_streams ?? {},
	).flat()[0]
	const destDb = firstStream?.destination_database

	if (!destDb) return { display: null, forModal: null }

	if (destDb.includes(":")) {
		const parts = destDb.split(":")
		return {
			display: `${parts[0]}_${"${source_namespace}"}`,
			forModal: parts[0],
		}
	}

	return { display: destDb, forModal: destDb }
}

export const selectUseFilterConfig = (state: StreamSelectionState) =>
	state.useFilterConfig

// Returns true when the given stream is selected (not disabled).
export const selectIsStreamEnabled = (
	state: StreamSelectionState,
	streamData: StreamData | null,
): boolean => {
	if (!streamData) return false

	const stream = state.streamsData?.selected_streams[
		streamData.stream.namespace || ""
	]?.find(s => s.stream_name === streamData.stream.name)

	if (!stream) return false
	return !stream.disabled
}
