import { Checkbox, Empty } from "antd"
// import { StreamsCollapsibleListProps } from "../../../../types" // Removed unused import
import { useEffect, useState } from "react"
import StreamPanel from "./StreamPanel"
import { StreamData } from "../../../../types"

interface GroupedStreamsCollapsibleListProps {
	groupedStreams: { [namespace: string]: StreamData[] }
	selectedStreams: {
		[namespace: string]: {
			stream_name: string
			partition_regex: string
			split_column: string
		}[]
	}
	setActiveStreamData: (stream: StreamData) => void
	activeStreamData: StreamData | null
	onStreamSelect: (
		streamName: string,
		checked: boolean,
		namespace: string,
	) => void
	setSelectedStreams: React.Dispatch<
		React.SetStateAction<
			| string[]
			| {
					[namespace: string]: {
						stream_name: string
						partition_regex: string
						split_column: string
					}[]
			  }
			| {
					selected_streams: {
						[namespace: string]: {
							stream_name: string
							partition_regex: string
							split_column: string
						}[]
					}
					streams: StreamData[]
			  }
		>
	>
	selectedStreamsFromAPI?: { [namespace: string]: { stream_name: string }[] }
}

// Add type guard functions at the top of the component
const isCombinedStreamsData = (
	obj: any,
): obj is {
	selected_streams: {
		[namespace: string]: {
			stream_name: string
			partition_regex: string
			split_column: string
		}[]
	}
	streams: StreamData[]
} => {
	return (
		obj &&
		typeof obj === "object" &&
		"selected_streams" in obj &&
		"streams" in obj
	)
}

const isStreamMapping = (
	obj: any,
): obj is {
	[namespace: string]: {
		stream_name: string
		partition_regex: string
		split_column: string
	}[]
} => {
	return (
		obj && typeof obj === "object" && !Array.isArray(obj) && !("streams" in obj)
	)
}

const StreamsCollapsibleList = ({
	groupedStreams,
	selectedStreams,
	setActiveStreamData,
	activeStreamData,
	onStreamSelect,
	setSelectedStreams,
	selectedStreamsFromAPI,
}: GroupedStreamsCollapsibleListProps) => {
	const [openNamespaces, setOpenNamespaces] = useState<{
		[ns: string]: boolean
	}>({})

	// Open all by default
	useEffect(() => {
		const allOpen: { [ns: string]: boolean } = {}
		Object.keys(groupedStreams).forEach((ns: string) => {
			allOpen[ns] = true
		})
		setOpenNamespaces(allOpen)
	}, [groupedStreams])

	// Initialize selected streams from API response
	useEffect(() => {
		if (selectedStreamsFromAPI) {
			const updated: {
				[namespace: string]: {
					stream_name: string
					partition_regex: string
					split_column: string
				}[]
			} = {}

			let hasChanges = false

			Object.entries(selectedStreamsFromAPI).forEach(([ns, streams]) => {
				updated[ns] = streams.map(stream => ({
					stream_name: stream.stream_name,
					partition_regex: "",
					split_column: "",
				}))
				hasChanges = true
			})

			// Only update if there are actual changes
			if (hasChanges) {
				setSelectedStreams(updated)
			}
		}
	}, [selectedStreamsFromAPI]) // Remove setSelectedStreams from dependencies

	const handleToggleNamespace = (ns: string) => {
		setOpenNamespaces(prev => ({ ...prev, [ns]: !prev[ns] }))
	}

	const handleNamespaceSyncAll = (ns: string, checked: boolean) => {
		if (!checked) {
			// When unchecking a database, remove all its streams from selected_streams
			setSelectedStreams(prev => {
				// Handle different types of prev structure
				if (Array.isArray(prev)) {
					return prev // Can't modify string[] structure
				} else if (isCombinedStreamsData(prev)) {
					// CombinedStreamsData structure
					const updated = { ...prev }
					delete updated.selected_streams[ns]
					return updated
				} else if (isStreamMapping(prev)) {
					// Original object structure
					const updated = { ...prev }
					delete updated[ns]
					return updated
				}
				return prev // Default fallback
			})
		} else {
			// When checking a database, add all its streams to selected_streams
			const streamNames = groupedStreams[ns].map(s => s.stream.name)

			setSelectedStreams(prev => {
				// Handle different types of prev structure
				if (Array.isArray(prev)) {
					return prev // Can't modify string[] structure
				} else if (isCombinedStreamsData(prev)) {
					// CombinedStreamsData structure
					const updated = { ...prev }
					updated.selected_streams[ns] = streamNames.map(stream_name => ({
						stream_name,
						partition_regex: "",
						split_column: "",
					}))
					return updated
				} else if (isStreamMapping(prev)) {
					// Original object structure
					const updated = { ...prev }
					updated[ns] = streamNames.map(stream_name => ({
						stream_name,
						partition_regex: "",
						split_column: "",
					}))
					return updated
				}
				return prev // Default fallback
			})
		}
	}

	const handleGlobalSyncAll = (checked: boolean) => {
		setSelectedStreams(prev => {
			// Handle different types of prev structure
			if (Array.isArray(prev)) {
				return prev // Can't modify string[] structure
			} else if (isCombinedStreamsData(prev)) {
				// CombinedStreamsData structure
				if (!checked) {
					const updated = { ...prev }
					updated.selected_streams = {}
					return updated
				}

				const updated = { ...prev }
				Object.entries(groupedStreams).forEach(([ns, streams]) => {
					updated.selected_streams[ns] = streams.map(s => ({
						stream_name: s.stream.name,
						partition_regex: "",
						split_column: "",
					}))
				})
				return updated
			} else if (isStreamMapping(prev)) {
				// Original object structure
				if (!checked) {
					return {}
				}

				const updated: {
					[namespace: string]: {
						stream_name: string
						partition_regex: string
						split_column: string
					}[]
				} = {}

				Object.entries(groupedStreams).forEach(([ns, streams]) => {
					updated[ns] = streams.map(s => ({
						stream_name: s.stream.name,
						partition_regex: "",
						split_column: "",
					}))
				})
				return updated
			}
			return prev // Default fallback
		})
	}

	const allStreamsSelected = Object.entries(groupedStreams).every(
		([ns, streams]) =>
			streams.every(s => {
				// Check if the stream is selected in the correct data structure
				if ("selected_streams" in selectedStreams) {
					// CombinedStreamsData format
					return (selectedStreams as any).selected_streams[ns]?.some(
						(selected: { stream_name: string }) =>
							selected.stream_name === s.stream.name,
					)
				} else {
					// Regular mapping format
					return (selectedStreams as any)[ns]?.some(
						(selected: { stream_name: string }) =>
							selected.stream_name === s.stream.name,
					)
				}
			}),
	)

	return (
		<div className="flex h-full flex-col">
			{Object.keys(groupedStreams).length === 0 ? (
				<Empty className="pt-10" />
			) : (
				<>
					<div className="mb-4 flex items-center">
						<Checkbox
							checked={allStreamsSelected}
							onChange={e => handleGlobalSyncAll(e.target.checked)}
						>
							Sync all
						</Checkbox>
					</div>
					{Object.entries(groupedStreams).map(([ns, streams]) => {
						const allChecked = streams.every(s => {
							// Check if the stream is selected in the correct data structure
							if ("selected_streams" in selectedStreams) {
								// CombinedStreamsData format
								return (selectedStreams as any).selected_streams[ns]?.some(
									(selected: { stream_name: string }) =>
										selected.stream_name === s.stream.name,
								)
							} else {
								// Regular mapping format
								return (selectedStreams as any)[ns]?.some(
									(selected: { stream_name: string }) =>
										selected.stream_name === s.stream.name,
								)
							}
						})
						return (
							<div
								key={ns}
								className="mb-2 border border-solid border-[#e5e7eb]"
							>
								<div
									className="flex cursor-pointer items-center border-b border-solid border-[#e5e7eb] bg-[#f5f5f5] p-3"
									onClick={() => handleToggleNamespace(ns)}
								>
									<Checkbox
										checked={allChecked}
										onChange={e => handleNamespaceSyncAll(ns, e.target.checked)}
										onClick={e => e.stopPropagation()}
										className="mr-2"
									/>
									<span className="font-semibold">{ns}</span>
									<span className="ml-auto">
										{openNamespaces[ns] ? "▼" : "►"}
									</span>
								</div>
								{openNamespaces[ns] && (
									<div className="w-full">
										{streams.map(streamData => (
											<StreamPanel
												stream={streamData}
												key={streamData?.stream?.name}
												activeStreamData={activeStreamData}
												setActiveStreamData={setActiveStreamData}
												onStreamSelect={(streamName, checked) =>
													onStreamSelect(streamName, checked, ns)
												}
												isSelected={
													"selected_streams" in selectedStreams
														? (selectedStreams as any).selected_streams[
																ns
															]?.some(
																(s: { stream_name: string }) =>
																	s.stream_name === streamData.stream.name,
															)
														: (selectedStreams as any)[ns]?.some(
																(s: { stream_name: string }) =>
																	s.stream_name === streamData.stream.name,
															)
												}
											/>
										))}
									</div>
								)}
							</div>
						)
					})}
				</>
			)}
		</div>
	)
}

export default StreamsCollapsibleList
