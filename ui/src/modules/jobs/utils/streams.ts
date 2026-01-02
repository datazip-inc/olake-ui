import {
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamData,
	IngestionMode,
} from "../../../types"
import {
	isDestinationIngestionModeSupported,
	isSourceIngestionModeSupported,
} from "../../../utils/utils"
import { STREAM_DEFAULTS } from "../../../utils/constants"

/**
 * Processes the raw SourceStreamsResponse into the
 * StreamsDataStructure expected by the UI.
 */
export const getStreamsDataFromSourceStreamsResponse = (
	response: StreamsDataStructure,
	destinationType?: string,
	sourceType?: string,
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
			})
		}
	})

	return {
		streams: response.streams,
		selected_streams: mergedSelectedStreams,
	}
}
