import {
	GetSourceStreamsResponse,
	SelectedStreamsByNamespace,
	StreamsDataStructure,
	StreamData,
} from "../../../types"

/**
 * Processes the raw GetSourceStreamsResponse into the
 * StreamsDataStructure expected by the UI.
 */
export const getStreamsDataFromGetSourceStreamsResponse = (
	response: GetSourceStreamsResponse,
): StreamsDataStructure => {
	const hasDefaultStreams =
		!!response.default_streams &&
		Object.keys(response.default_streams).length > 0

	// Backward compatibility for older OLake versions without default_streams
	if (!hasDefaultStreams) {
		const mergedSelectedStreams: SelectedStreamsByNamespace =
			response.selected_streams || {}

		response.streams.forEach((stream: StreamData) => {
			const namespace = stream.stream.namespace || ""
			if (!mergedSelectedStreams[namespace]) {
				mergedSelectedStreams[namespace] = []
			}
		})

		return {
			streams: response.streams,
			selected_streams: mergedSelectedStreams,
		}
	}

	const mergedSelectedStreams: SelectedStreamsByNamespace = {}

	Object.entries(response.default_streams || {}).forEach(
		([namespace, defaultNamespaceStreams]) => {
			const selectedNamespaceStreams =
				response.selected_streams?.[namespace] || []

			mergedSelectedStreams[namespace] = defaultNamespaceStreams.map(
				defaultStream => {
					const matchingSelectedStream = selectedNamespaceStreams.find(
						s => s.stream_name === defaultStream.stream_name,
					)

					// Use selected stream when it exists
					if (matchingSelectedStream) {
						return {
							...defaultStream,
							...matchingSelectedStream,
							disabled: false,
						}
					}

					// Otherwise fall back to default
					return {
						...defaultStream,
						disabled: true,
					}
				},
			)
		},
	)

	return {
		streams: response.streams,
		selected_streams: mergedSelectedStreams,
	}
}
