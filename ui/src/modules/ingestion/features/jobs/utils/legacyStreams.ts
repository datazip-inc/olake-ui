// Deprecated: legacy streams.json support for sources below MinStreamsV2Version.
// Backward compatibility only: new v2 fields/features are not carried into the
// legacy format and must be version-gated in the UI. Delete once v2 is the minimum.
import {
	LegacyStreamsConfig,
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamsV2Catalog,
} from "@/modules/ingestion/common/types"

export const fromLegacyCatalog = ({
	streams,
	selected_streams,
}: LegacyStreamsConfig): StreamsV2Catalog => ({
	available_streams: { streams },
	selected_streams: selected_streams ?? {},
})

// Legacy drivers read sync_mode, cursor_field and destination_database from
// streams[], so copy each entry's values back onto its stream.
export const toLegacyStreamsConfig = (
	streamsData: StreamsDataStructure,
	selectedStreams: SelectedStreamsByNamespace,
): string => {
	const streams = streamsData.streams.map(s => {
		const entry = streamsData.selected_streams[s.stream.namespace || ""]?.find(
			e => e.stream_name === s.stream.name,
		)
		if (!entry) return s
		const stream = {
			...s.stream,
			sync_mode: entry.sync_mode,
			destination_database: entry.destination_database,
		}
		if (entry.cursor_field) stream.cursor_field = entry.cursor_field
		else delete stream.cursor_field
		return { ...s, stream }
	})

	return JSON.stringify({ streams, selected_streams: selectedStreams })
}
