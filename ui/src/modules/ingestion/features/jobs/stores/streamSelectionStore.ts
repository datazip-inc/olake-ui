import { create } from "zustand"

import {
	StreamsDataStructure,
	StreamData,
	SelectedStream,
	SelectedStreamsByNamespace,
	SelectedColumns,
	SyncMode,
	FilterConfig,
} from "@/modules/ingestion/common/types"
import { IngestionMode } from "@/modules/ingestion/features/jobs/enums"

import { STREAM_DEFAULTS } from "../constants"
import { extractNamespaceFromDestination } from "../utils"

interface StreamSelectionState {
	streamsData: StreamsDataStructure | null

	// Frozen snapshot from initial discovery; used by DestinationDatabaseModal.
	initialStreamsSnapshot: StreamsDataStructure | null
	isDiscovering: boolean
	discoverError: string | null
	activeStreamKey: { name: string; namespace: string } | null

	// Per-stream filter toggle keyed by `${namespace}_${name}`.
	streamFilterStates: Record<string, boolean>

	// Whether to use structured filter_config (new) vs legacy filter string.
	useFilterConfig: boolean

	initializeFromDiscovery: (data: StreamsDataStructure) => void
	setDiscovering: (loading: boolean) => void
	setDiscoverError: (message: string | null) => void

	// Toggles a stream on (disabled:false) or off (disabled:true).
	// Inserts a default entry if the stream has never been in selected_streams.
	toggleStream: (
		streamName: string,
		namespace: string,
		checked: boolean,
		ingestionMode: IngestionMode,
	) => void

	// Updates sync_mode and optional cursor_field; no-op if unchanged.
	updateSyncMode: (
		streamName: string,
		namespace: string,
		syncMode: SyncMode,
		cursorField?: string,
	) => void

	updateNormalization: (
		streamName: string,
		namespace: string,
		normalization: boolean,
	) => void

	updatePartitionRegex: (
		streamName: string,
		namespace: string,
		regex: string,
	) => void

	// Empty string removes the `filter` key entirely.
	updateFilter: (
		streamName: string,
		namespace: string,
		filterValue: string,
	) => void

	updateFilterConfig: (
		streamName: string,
		namespace: string,
		filterConfig: FilterConfig | undefined,
	) => void

	setUseFilterConfig: (value: boolean) => void

	updateIngestionMode: (
		streamName: string,
		namespace: string,
		appendMode: boolean,
	) => void

	// Applies append_mode to every stream in selected_streams.
	updateAllIngestionMode: (appendMode: boolean) => void

	// Updates destination_database on all streams.
	updateDestinationDatabase: (format: string, databaseName: string) => void

	updateSelectedColumns: (
		streamName: string,
		namespace: string,
		columns: SelectedColumns,
	) => void

	setStreamFilterState: (streamKey: string, value: boolean) => void
	setActiveStreamKey: (key: { name: string; namespace: string } | null) => void
	reset: () => void
}

const initialState = {
	streamsData: null,
	initialStreamsSnapshot: null,
	isDiscovering: false,
	discoverError: null,
	activeStreamKey: null,
	streamFilterStates: {} as Record<string, boolean>,
	useFilterConfig: false,
}

export const useStreamSelectionStore = create<StreamSelectionState>()(set => ({
	...initialState,

	initializeFromDiscovery: data =>
		set(state => ({
			streamsData: data,
			initialStreamsSnapshot: state.initialStreamsSnapshot ?? data,
			isDiscovering: false,
			discoverError: null,
		})),

	setDiscovering: loading => set({ isDiscovering: loading }),
	setDiscoverError: error => set({ discoverError: error }),

	toggleStream: (streamName, namespace, checked, ingestionMode) =>
		set(state => {
			if (!state.streamsData) return state

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
				if (!existingStream) {
					updated.selected_streams[namespace] = [
						...updated.selected_streams[namespace],
						{
							...STREAM_DEFAULTS,
							stream_name: streamName,
							disabled: false,
							append_mode: ingestionMode === IngestionMode.APPEND,
						},
					]
					changed = true
				} else if (existingStream.disabled) {
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

	updateSyncMode: (streamName, namespace, newSyncMode, cursorField) =>
		set(state => {
			if (!state.streamsData) return state

			const prev = state.streamsData
			const streamIndex = prev.streams.findIndex(
				s => s.stream.name === streamName && s.stream.namespace === namespace,
			)

			if (
				streamIndex !== -1 &&
				prev.streams[streamIndex].stream.sync_mode === newSyncMode &&
				(prev.streams[streamIndex].stream.cursor_field || "") ===
					(cursorField || "")
			) {
				return state
			}

			if (streamIndex === -1) return state

			const updatedStreams = [...prev.streams]
			const nextStream: StreamData = {
				...updatedStreams[streamIndex],
				stream: {
					...updatedStreams[streamIndex].stream,
					sync_mode: newSyncMode,
				},
			}

			if (cursorField !== undefined && newSyncMode === SyncMode.INCREMENTAL) {
				nextStream.stream.cursor_field = cursorField
			}
			if (newSyncMode !== SyncMode.INCREMENTAL) {
				delete nextStream.stream.cursor_field
			}

			updatedStreams[streamIndex] = nextStream

			return {
				streamsData: { ...prev, streams: updatedStreams },
			}
		}),

	updateNormalization: (streamName, namespace, normalization) =>
		set(state => {
			if (!state.streamsData) return state

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

	updatePartitionRegex: (streamName, namespace, regex) =>
		set(state => {
			if (!state.streamsData) return state

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

	updateFilter: (streamName, namespace, filterValue) =>
		set(state => {
			if (!state.streamsData) return state

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

	updateFilterConfig: (streamName, namespace, filterConfig) =>
		set(state => {
			if (!state.streamsData) return state

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

	updateIngestionMode: (streamName, namespace, appendMode) =>
		set(state => {
			if (!state.streamsData) return state

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
								? { ...s, append_mode: appendMode }
								: s,
						),
					},
				},
			}
		}),

	updateAllIngestionMode: appendMode =>
		set(state => {
			if (!state.streamsData) return state

			const prev = state.streamsData
			const updatedSelected = Object.fromEntries(
				Object.entries(prev.selected_streams).map(([ns, streams]) => [
					ns,
					streams.map(s => ({ ...s, append_mode: appendMode })),
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
			const firstStreamDestDb = prev.streams[0].stream.destination_database
			const hasColonFormat =
				firstStreamDestDb && firstStreamDestDb.includes(":")

			const updatedStreams = prev.streams.map(stream => {
				const currentDestDb = stream.stream.destination_database
				const currentNamespace = stream.stream.namespace

				if (format === "dynamic") {
					if (hasColonFormat && currentDestDb) {
						// "a:b" → "databaseName:b"
						const parts = currentDestDb.split(":")
						return {
							...stream,
							stream: {
								...stream.stream,
								destination_database: `${databaseName}:${parts[1]}`,
							},
						}
					} else {
						// No colon — derive namespace from initialStreamsSnapshot
						const initialStream = state.initialStreamsSnapshot?.streams.find(
							s =>
								s.stream.name === stream.stream.name &&
								s.stream.namespace === stream.stream.namespace,
						)
						const namespace = extractNamespaceFromDestination(
							initialStream?.stream.destination_database,
							currentNamespace || "",
						)
						return {
							...stream,
							stream: {
								...stream.stream,
								destination_database: `${databaseName}:${namespace}`,
							},
						}
					}
				} else {
					return {
						...stream,
						stream: {
							...stream.stream,
							destination_database: databaseName,
						},
					}
				}
			})

			return {
				streamsData: { ...prev, streams: updatedStreams },
			}
		}),

	updateSelectedColumns: (streamName, namespace, columns) =>
		set(state => {
			if (!state.streamsData) return state

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
				s.stream.name === state.activeStreamKey!.name,
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
			s => s.stream_name === state.activeStreamKey!.name,
		) ?? null
	)
}

// Derives destination database display values from the first stream.
export const selectDestinationDatabase = (
	state: StreamSelectionState,
): { display: string | null; forModal: string | null } => {
	if (!state.streamsData?.streams || state.streamsData.streams.length === 0) {
		return { display: null, forModal: null }
	}

	const firstStream = state.streamsData.streams[0]
	const destDb = firstStream.stream?.destination_database

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
