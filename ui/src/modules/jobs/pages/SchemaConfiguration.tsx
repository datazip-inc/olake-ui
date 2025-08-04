import { useEffect, useState, useMemo, useRef } from "react"
import { Input, Empty, Spin } from "antd"
import FilterButton from "../components/FilterButton"
import StreamsCollapsibleList from "./streams/StreamsCollapsibleList"
import {
	CombinedStreamsData,
	SchemaConfigurationProps,
	StreamData,
} from "../../../types"
import StreamConfiguration from "./streams/StreamConfiguration"
import StepTitle from "../../common/components/StepTitle"
import { sourceService } from "../../../api"
import React from "react"

const SchemaConfiguration: React.FC<SchemaConfigurationProps> = ({
	setSelectedStreams,
	stepNumber = 3,
	stepTitle = "Streams Selection",
	useDirectForms = false,
	sourceName,
	sourceConnector,
	sourceVersion,
	sourceConfig,
	initialStreamsData,
	fromJobEditFlow = false,
	jobId = -1,
	destinationType,
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
				normalization: boolean
				filter?: string
			}[]
		}
		streams: StreamData[]
	} | null>(initialStreamsData || null)
	const [loading, setLoading] = useState(!initialStreamsData)

	// Use ref to track if we've initialized to prevent double updates
	const initialized = useRef(false)

	useEffect(() => {
		// Reset initialized ref when source config changes
		if (sourceConfig !== prevSourceConfig.current) {
			initialized.current = false
			prevSourceConfig.current = sourceConfig
		}

		if (
			initialStreamsData &&
			initialStreamsData.selected_streams &&
			Object.keys(initialStreamsData.selected_streams).length > 0
		) {
			setApiResponse(initialStreamsData)
			setSelectedStreams(initialStreamsData)
			setLoading(false)
			initialized.current = true

			// Select first stream if no stream is currently active
			if (!activeStreamData && initialStreamsData.streams.length > 0) {
				setActiveStreamData(initialStreamsData.streams[0])
			}
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
					fromJobEditFlow ? jobId : -1,
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
					const namespace = stream.stream.namespace || ""
					if (!processedResponseData.selected_streams[namespace]) {
						processedResponseData.selected_streams[namespace] = []
					}
				})

				setApiResponse(processedResponseData)
				setSelectedStreams(processedResponseData)

				// Always select first stream if no stream is currently active
				if (processedResponseData.streams.length > 0 && !activeStreamData) {
					setActiveStreamData(processedResponseData.streams[0])
				}

				initialized.current = true
			} catch (error) {
				console.error("Error fetching source streams:", error)
			} finally {
				setLoading(false)
			}
		}

		if (!initialized.current && sourceConfig && sourceConnector) {
			fetchSourceStreams()
		}
	}, [
		sourceName,
		sourceConnector,
		sourceVersion,
		sourceConfig,
		initialStreamsData,
	])

	const prevSourceConfig = useRef(sourceConfig)

	const handleStreamSyncModeChange = (
		streamName: string,
		namespace: string,
		newSyncMode: "full_refresh" | "cdc" | "incremental" | "strict_cdc",
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

	const handleNormalizationChange = (
		streamName: string,
		namespace: string,
		normalization: boolean,
	) => {
		setApiResponse(prev => {
			if (!prev) return prev

			const streamExistsInSelected = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)

			if (!streamExistsInSelected) return prev

			const updatedSelectedStreams = {
				...prev.selected_streams,
				[namespace]: prev.selected_streams[namespace].map(s =>
					s.stream_name === streamName ? { ...s, normalization } : s,
				),
			}

			const updated = {
				...prev,
				selected_streams: updatedSelectedStreams,
			}

			setSelectedStreams(updated)
			return updated
		})
	}

	const handlePartitionRegexChange = (
		streamName: string,
		namespace: string,
		partitionRegex: string,
	) => {
		setApiResponse(prev => {
			if (!prev) return prev

			const streamExistsInSelected = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)

			if (!streamExistsInSelected) return prev // Should not happen if UI is correct

			const updatedSelectedStreams = {
				...prev.selected_streams,
				[namespace]: prev.selected_streams[namespace].map(s =>
					s.stream_name === streamName
						? { ...s, partition_regex: partitionRegex }
						: s,
				),
			}

			const updated = {
				...prev,
				selected_streams: updatedSelectedStreams,
			}

			setSelectedStreams(updated)
			return updated
		})
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
							normalization: false,
							filter: "",
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

	const handleFullLoadFilterChange = (
		streamName: string,
		namespace: string,
		filterValue: string,
	) => {
		setApiResponse(prev => {
			if (!prev) return prev

			const streamExistsInSelected = prev.selected_streams[namespace]?.some(
				s => s.stream_name === streamName,
			)

			if (!streamExistsInSelected) return prev

			const updatedSelectedStreams = {
				...prev.selected_streams,
				[namespace]: prev.selected_streams[namespace].map(s =>
					s.stream_name === streamName ? { ...s, filter: filterValue } : s,
				),
			}

			const updated = {
				...prev,
				selected_streams: updatedSelectedStreams,
			}

			setSelectedStreams(updated)
			return updated
		})
	}

	const filteredStreams = useMemo(() => {
		if (!apiResponse?.streams) return []
		let tempFilteredStreams = [...apiResponse.streams]

		if (searchText) {
			tempFilteredStreams = tempFilteredStreams.filter(stream =>
				stream.stream.name.toLowerCase().includes(searchText.toLowerCase()),
			)
		}

		if (selectedFilters.includes("All tables")) {
			return tempFilteredStreams
		}

		return tempFilteredStreams.filter(stream => {
			const cdcIsActive = selectedFilters.includes("CDC")
			const frIsActive = selectedFilters.includes("Full refresh")
			const hasSelectedFilter = selectedFilters.includes("Selected")
			const hasNotSelectedFilter = selectedFilters.includes("Not selected")

			// Sync mode filtering
			let passesSyncModeFilter = true
			if (cdcIsActive && frIsActive) {
				passesSyncModeFilter =
					stream.stream.sync_mode === "cdc" ||
					stream.stream.sync_mode === "full_refresh"
			} else if (cdcIsActive) {
				passesSyncModeFilter = stream.stream.sync_mode === "cdc"
			} else if (frIsActive) {
				passesSyncModeFilter = stream.stream.sync_mode === "full_refresh"
			}

			if (!passesSyncModeFilter) {
				return false
			}

			// Selection status filtering
			const isSelected = apiResponse.selected_streams[
				stream.stream.namespace || ""
			]?.some(s => s.stream_name === stream.stream.name)

			if (hasSelectedFilter && hasNotSelectedFilter) {
				// No filtering based on selection status if both are selected
			} else if (hasSelectedFilter) {
				if (!isSelected) return false
			} else if (hasNotSelectedFilter) {
				if (isSelected) return false
			}

			return true
		})
	}, [apiResponse, searchText, selectedFilters])

	const groupedFilteredStreams = useMemo(() => {
		const grouped: { [namespace: string]: StreamData[] } = {}
		filteredStreams.forEach(stream => {
			const ns = stream.stream.namespace || ""
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
				<div
					className={`${activeStreamData ? "w-1/2" : "w-full"} max-h-[calc(100vh-250px)] overflow-y-auto`}
				>
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
						/>
					) : loading ? (
						<div className="flex h-[calc(100vh-250px)] items-center justify-center">
							<Spin size="large"></Spin>
						</div>
					) : (
						<Empty className="flex h-full flex-col items-center justify-center" />
					)}
				</div>

				<div
					className={`sticky top-0 mx-4 flex w-1/2 flex-col rounded-xl ${!loading ? "border" : ""} bg-[#ffffff] p-4 transition-all duration-150 ease-linear`}
				>
					{activeStreamData ? (
						<StreamConfiguration
							stream={activeStreamData}
							onUpdate={() => {
								// Update the stream config in the local state
								// Implementation will be added later if needed
							}}
							onSyncModeChange={(
								streamName: string,
								namespace: string,
								syncMode: "full_refresh" | "cdc" | "incremental" | "strict_cdc",
							) => {
								handleStreamSyncModeChange(streamName, namespace, syncMode)
							}}
							useDirectForms={useDirectForms}
							isSelected={
								!!apiResponse?.selected_streams[
									activeStreamData.stream.namespace || ""
								]?.some(s => s.stream_name === activeStreamData.stream.name)
							}
							initialNormalization={
								apiResponse?.selected_streams[
									activeStreamData.stream.namespace || ""
								]?.find(s => s.stream_name === activeStreamData.stream.name)
									?.normalization || false
							}
							onNormalizationChange={handleNormalizationChange}
							initialPartitionRegex={
								apiResponse?.selected_streams[
									activeStreamData.stream.namespace || ""
								]?.find(s => s.stream_name === activeStreamData.stream.name)
									?.partition_regex || ""
							}
							onPartitionRegexChange={handlePartitionRegexChange}
							initialFullLoadFilter={
								apiResponse?.selected_streams[
									activeStreamData.stream.namespace || ""
								]?.find(s => s.stream_name === activeStreamData.stream.name)
									?.filter || ""
							}
							onFullLoadFilterChange={handleFullLoadFilterChange}
							fromJobEditFlow={fromJobEditFlow}
							initialSelectedStreams={apiResponse || undefined}
							destinationType={destinationType}
						/>
					) : null}
				</div>
			</div>
		</div>
	)
}

export default SchemaConfiguration
