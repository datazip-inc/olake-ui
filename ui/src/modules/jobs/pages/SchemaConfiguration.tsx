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
		if (initialStreamsData) {
			setApiResponse(initialStreamsData)
			setSelectedStreams(initialStreamsData)
			setLoading(false)
			initialized.current = true
			return
		}

		const fetchSourceStreams = async () => {
			if (initialized.current) return

			setLoading(true)
			try {
				const response = await sourceService.getSourceStreams(
					sourceName,
					sourceConnector,
					sourceVersion,
					sourceConfig,
				)

				const rawApiResponse = response.data as any
				const processedResponseData: CombinedStreamsData = {
					streams: [],
					selected_streams: {},
				}

				if (rawApiResponse && Array.isArray(rawApiResponse.streams)) {
					processedResponseData.streams = rawApiResponse.streams
				}

				if (
					rawApiResponse &&
					typeof rawApiResponse.selected_streams === "object" &&
					rawApiResponse.selected_streams !== null &&
					!Array.isArray(rawApiResponse.selected_streams)
				) {
					for (const ns in rawApiResponse.selected_streams) {
						if (
							Object.prototype.hasOwnProperty.call(
								rawApiResponse.selected_streams,
								ns,
							) &&
							Array.isArray(rawApiResponse.selected_streams[ns])
						) {
							processedResponseData.selected_streams[ns] =
								rawApiResponse.selected_streams[ns]
						}
					}
				}

				processedResponseData.streams.forEach((stream: StreamData) => {
					const namespace = stream.stream.namespace || "default"
					if (!processedResponseData.selected_streams[namespace]) {
						processedResponseData.selected_streams[namespace] = []
					}
				})

				setApiResponse(processedResponseData)
				setSelectedStreams(processedResponseData)
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
		setApiResponse(prev => {
			if (!prev) return prev

			const streamIndex = prev.streams.findIndex(
				s => s.stream.name === streamName && s.stream.namespace === namespace,
			)

			if (
				streamIndex !== -1 &&
				prev.streams[streamIndex].stream.sync_mode === newSyncMode
			) {
				return prev
			}

			const updated = { ...prev }

			if (streamIndex !== -1) {
				updated.streams = [...prev.streams]
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

		setTimeout(() => {
			setApiResponse(current => {
				if (!current) return current

				const updatedData = {
					selected_streams: current.selected_streams,
					streams: current.streams,
				}
				setSelectedStreams(updatedData)
				return current
			})
		}, 0)
	}

	const handleStreamSelect = (
		streamName: string,
		checked: boolean,
		namespace: string,
	) => {
		setApiResponse(prev => {
			if (!prev) return prev

			const updated = {
				...prev,
				selected_streams: { ...prev.selected_streams },
			}
			let changed = false
			if (checked) {
				if (!updated.selected_streams[namespace]) {
					updated.selected_streams[namespace] = []
					changed = true
				}
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

			return changed ? updated : prev
		})

		setTimeout(() => {
			setApiResponse(current => {
				if (!current) return current
				const updatedData = {
					selected_streams: current.selected_streams,
					streams: current.streams,
				}
				setSelectedStreams(updatedData)
				return current
			})
		}, 0)
	}

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
