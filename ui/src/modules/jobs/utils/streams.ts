import {
	GetSourceStreamsResponse,
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamData,
	IngestionMode,
} from "../../../types"
import {
	isDestinationIngestionModeSupported,
	isSourceIngestionModeSupported,
} from "../../../utils/utils"

// fallback defaults for streams
const STREAM_DEFAULTS = {
	append_mode: false,
	normalization: false,
	partition_regex: "",
}

/**
 * Processes the raw GetSourceStreamsResponse into the
 * StreamsDataStructure expected by the UI.
 */
export const getStreamsDataFromGetSourceStreamsResponse = (
	response: GetSourceStreamsResponse,
	destinationType?: string,
	sourceType?: string,
): StreamsDataStructure => {
	const mergedSelectedStreams: SelectedStreamsByNamespace = {}
	const streamDefaults = response.stream_defaults

	const isDestUpsertSupported = isDestinationIngestionModeSupported(
		IngestionMode.UPSERT,
		destinationType,
	)

	const isSourceUpsertSupported = isSourceIngestionModeSupported(
		IngestionMode.UPSERT,
		sourceType,
	)

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
			// Stream is not selected, use defaults from stream_defaults
			// Missing properties in stream_defaults are treated as false/empty
			// Backward compatibility: fall back to hardcoded defaults if STREAM_DEFAULTS is not present (older olake versions)
			const defaults = {
				...STREAM_DEFAULTS,
				...streamDefaults,
			}

			mergedSelectedStreams[namespace].push({
				...defaults,
				stream_name: streamName,
				disabled: true,
				append_mode: !isDestUpsertSupported || !isSourceUpsertSupported, // Default to append if either source or destination does not support upsert
			})
		}
	})

	return {
		streams: response.streams,
		selected_streams: mergedSelectedStreams,
	}
}
