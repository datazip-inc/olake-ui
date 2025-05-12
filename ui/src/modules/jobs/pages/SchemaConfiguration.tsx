import { useEffect, useState, useMemo } from "react"
import { Input, Empty, Spin } from "antd"
import FilterButton from "../components/FilterButton"
import StreamsCollapsibleList from "./streams/StreamsCollapsibleList"
import { StreamData } from "../../../types"
import StreamConfiguration from "./streams/StreamConfiguration"
import StepTitle from "../../common/components/StepTitle"
import { sourceService } from "../../../api"
import React from "react"

interface CombinedStreamsData {
	selected_streams: {
		[namespace: string]: {
			stream_name: string
			partition_regex: string
			split_column: string
		}[]
	}
	streams: StreamData[]
}

interface SchemaConfigurationProps {
	selectedStreams:
		| string[]
		| {
				[namespace: string]: {
					stream_name: string
					partition_regex: string
					split_column: string
				}[]
		  }
		| CombinedStreamsData
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
			| CombinedStreamsData
		>
	>
	stepNumber?: number | string
	stepTitle?: string
	useDirectForms?: boolean
	sourceName: string
	sourceConnector: string
	sourceVersion: string
	sourceConfig: string
	initialStreamsData?: CombinedStreamsData
}

const SchemaConfiguration: React.FC<SchemaConfigurationProps> = ({
	setSelectedStreams,
	stepNumber = 3,
	stepTitle = "Schema evaluation",
	useDirectForms = false,
	sourceName,
	sourceConnector,
	sourceVersion,
	sourceConfig,
	initialStreamsData,
}) => {
	const [searchText, setSearchText] = useState("")
	const [selectedFilters, setSelectedFilters] = useState<string[]>([
		"All tables",
	])
	const [activeStreamData, setActiveStreamData] = useState<StreamData | null>(
		null,
	)
	const [apiResponse, setApiResponse] = useState<{
		selected_streams: {
			[namespace: string]: {
				stream_name: string
				partition_regex: string
				split_column: string
			}[]
		}
		streams: StreamData[]
	} | null>(initialStreamsData || null)
	const [loading, setLoading] = useState(!initialStreamsData)

	// Use ref to track if we've initialized to prevent double updates
	const initialized = React.useRef(!!initialStreamsData)

	useEffect(() => {
		// If initial data is provided, use it and skip fetching
		if (initialStreamsData) {
			setApiResponse(initialStreamsData)
			setSelectedStreams(initialStreamsData)
			setLoading(false)
			initialized.current = true
			return
		}

		const fetchSourceStreams = async () => {
			if (initialized.current) return // Skip if already initialized

			setLoading(true)
			try {
				const response = await sourceService.getSourceStreams(
					sourceName,
					sourceConnector,
					sourceVersion,
					sourceConfig,
				)

				// Fix: Check and ensure default namespace exists
				const responseData = response.data
				if (responseData.streams && responseData.streams.length > 0) {
					// Ensure selected_streams object is initialized properly
					if (!responseData.selected_streams) {
						responseData.selected_streams = {}
					}

					// Check for streams with undefined or 'default' namespace and ensure they're properly handled
					responseData.streams.forEach((stream: StreamData) => {
						const namespace = stream.stream.namespace || "default"
						if (!responseData.selected_streams[namespace]) {
							responseData.selected_streams[namespace] = []
						}
					})
				}

				// First update the API response
				setApiResponse(responseData)

				// Then set the selected streams for the parent component
				setSelectedStreams(responseData)

				// Mark as initialized
				initialized.current = true
			} catch (error) {
				console.error("Error fetching source streams:", error)
			} finally {
				setLoading(false)
			}
		}

		fetchSourceStreams()
	}, [
		sourceName,
		sourceConnector,
		sourceVersion,
		sourceConfig,
		initialStreamsData,
		setSelectedStreams,
	])

	// Update selected streams when sync mode changes
	const handleStreamSyncModeChange = (
		streamName: string,
		namespace: string,
		newSyncMode: "full_refresh" | "cdc",
	) => {
		// First update the API response
		setApiResponse(prev => {
			if (!prev) return prev

			// Check if this is actually a change
			const streamIndex = prev.streams.findIndex(
				s => s.stream.name === streamName && s.stream.namespace === namespace,
			)

			if (
				streamIndex !== -1 &&
				prev.streams[streamIndex].stream.sync_mode === newSyncMode
			) {
				// No change needed, return existing state
				return prev
			}

			// Create a new object to avoid modifying existing state
			const updated = { ...prev }

			if (streamIndex !== -1) {
				updated.streams = [...prev.streams] // Clone the array
				updated.streams[streamIndex] = {
					...updated.streams[streamIndex],
					stream: {
						...updated.streams[streamIndex].stream,
						sync_mode: newSyncMode,
					},
				}
			}

			return updated
		})

		// Then update the parent's selected streams
		setTimeout(() => {
			setApiResponse(current => {
				if (!current) return current

				// Create the combined data structure for the parent
				const updatedData = {
					selected_streams: current.selected_streams,
					streams: current.streams,
				}

				// Update the parent component with the full structure
				setSelectedStreams(updatedData)

				return current // Return current state without modifying it
			})
		}, 0)
	}

	const handleStreamSelect = (
		streamName: string,
		checked: boolean,
		namespace: string,
	) => {
		// First update the API response
		setApiResponse(prev => {
			if (!prev) return prev

			// Create a new object to avoid modifying existing state
			const updated = {
				...prev,
				selected_streams: { ...prev.selected_streams },
			}

			// Determine if we need to make a change
			let changed = false

			if (checked) {
				// Add stream to selected_streams
				if (!updated.selected_streams[namespace]) {
					updated.selected_streams[namespace] = []
					changed = true
				}

				// Check if the stream is already in the selected streams
				if (
					!updated.selected_streams[namespace].some(
						s => s.stream_name === streamName,
					)
				) {
					updated.selected_streams[namespace] = [
						...updated.selected_streams[namespace],
						{
							stream_name: streamName,
							partition_regex: "",
							split_column: "",
						},
					]
					changed = true
				}
			} else {
				// Remove stream from selected_streams
				if (updated.selected_streams[namespace]) {
					const filtered = updated.selected_streams[namespace].filter(
						s => s.stream_name !== streamName,
					)

					if (filtered.length !== updated.selected_streams[namespace].length) {
						updated.selected_streams[namespace] = filtered
						changed = true
					}
				}
			}

			// Only return the updated object if something changed
			return changed ? updated : prev
		})

		// Then update the parent's selected streams
		setTimeout(() => {
			setApiResponse(current => {
				if (!current) return current

				// Create the combined data structure for the parent
				const updatedData = {
					selected_streams: current.selected_streams,
					streams: current.streams,
				}

				// Update the parent component with the full structure
				setSelectedStreams(updatedData)

				return current // Return current state without modifying it
			})
		}, 0)
	}

	// Filter streams based on search and selected filters
	const filteredStreams = useMemo(() => {
		if (!apiResponse?.streams) return []
		let filtered = [...apiResponse.streams]

		if (searchText) {
			filtered = filtered.filter(stream =>
				stream.stream.name.toLowerCase().includes(searchText.toLowerCase()),
			)
		}

		if (selectedFilters.includes("All tables")) return filtered

		return filtered.filter(stream => {
			const isSelected = apiResponse.selected_streams[
				stream.stream.namespace || "default"
			]?.some(s => s.stream_name === stream.stream.name)
			const hasSelectedFilter = selectedFilters.includes("Selected")
			const hasNotSelectedFilter = selectedFilters.includes("Not selected")

			if (hasSelectedFilter && hasNotSelectedFilter) return true
			if (hasSelectedFilter) return isSelected
			if (hasNotSelectedFilter) return !isSelected
			return true
		})
	}, [apiResponse, searchText, selectedFilters])

	// Group filtered streams by namespace
	const groupedFilteredStreams = useMemo(() => {
		const grouped: { [namespace: string]: StreamData[] } = {}
		filteredStreams.forEach(stream => {
			const ns = stream.stream.namespace || "default"
			if (!grouped[ns]) grouped[ns] = []
			grouped[ns].push(stream)
		})
		return grouped
	}, [filteredStreams])

	const filters = [
		"All tables",
		"CDC",
		"Full refresh",
		"Selected",
		"Not selected",
	]

	useEffect(() => {
		if (selectedFilters.length === 0) {
			setSelectedFilters(["All tables"])
		}
	}, [selectedFilters])

	const { Search } = Input

	return (
		<div className="mb-4 p-6">
			{stepNumber && stepTitle && (
				<StepTitle
					stepNumber={stepNumber}
					stepTitle={stepTitle}
				/>
			)}

			<div className="mb-4 mr-4 flex flex-wrap justify-start gap-4">
				<div className="w-full lg:w-[55%] xl:w-[40%]">
					<Search
						placeholder="Search streams"
						allowClear
						className="custom-search-input w-full"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
				</div>
				<div className="flex flex-wrap gap-2">
					{filters.map(filter => (
						<FilterButton
							key={filter}
							filter={filter}
							selectedFilters={selectedFilters}
							setSelectedFilters={setSelectedFilters}
						/>
					))}
				</div>
			</div>

			<div className="flex">
				<div className={`${activeStreamData ? "w-1/2" : "w-full"} `}>
					{!loading && apiResponse?.streams ? (
						<StreamsCollapsibleList
							groupedStreams={groupedFilteredStreams}
							selectedStreams={apiResponse.selected_streams}
							setActiveStreamData={setActiveStreamData}
							activeStreamData={activeStreamData}
							onStreamSelect={handleStreamSelect}
							setSelectedStreams={(updatedSelectedStreams: any) => {
								if (!apiResponse) return

								// Construct the full data structure
								const fullData = {
									selected_streams: updatedSelectedStreams,
									streams: apiResponse.streams,
								}

								// Pass it to the parent component
								setSelectedStreams(fullData as CombinedStreamsData)
							}}
							selectedStreamsFromAPI={apiResponse.selected_streams}
						/>
					) : loading ? (
						<Spin size="large">Loading streams...</Spin>
					) : (
						<Empty className="flex h-full flex-col items-center justify-center" />
					)}
				</div>

				{activeStreamData && (
					<div className="mx-4 flex h-full w-1/2 flex-col rounded-xl border bg-[#ffffff] p-4 transition-all duration-150 ease-linear">
						<StreamConfiguration
							stream={activeStreamData}
							onUpdate={() => {
								// Update the stream config in the local state
								// Implementation will be added later if needed
							}}
							onSyncModeChange={(
								streamName: string,
								namespace: string,
								syncMode: "full_refresh" | "cdc",
							) => {
								handleStreamSyncModeChange(streamName, namespace, syncMode)
							}}
							useDirectForms={useDirectForms}
						/>
					</div>
				)}
			</div>
		</div>
	)
}

export default SchemaConfiguration
