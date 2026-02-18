import React, { useEffect, useState, useMemo, useRef } from "react"
import { Input, Empty, Spin, Tooltip } from "antd"
import clsx from "clsx"

import { sourceService } from "../../../api"
import { useAppStore } from "../../../store"
import {
	IngestionMode,
	SchemaConfigurationProps,
	StreamData,
	SyncMode,
	StreamsDataStructure,
	SelectedColumns,
} from "../../../types"
import FilterButton from "../components/FilterButton"
import StepTitle from "../../common/components/StepTitle"
import StreamsCollapsibleList from "./streams/StreamsCollapsibleList"
import StreamConfiguration from "./streams/StreamConfiguration"
import {
	ArrowSquareOutIcon,
	InfoIcon,
	PencilSimpleIcon,
} from "@phosphor-icons/react"
import {
	DESTINATION_INTERNAL_TYPES,
	DESTINATATION_DATABASE_TOOLTIP_TEXT,
	STREAM_DEFAULTS,
} from "../../../utils/constants"
import { extractNamespaceFromDestination } from "../../../utils/destination-database"
import DestinationDatabaseModal from "../../common/Modals/DestinationDatabaseModal"
import { getStreamsDataFromSourceStreamsResponse } from "../utils/streams"

const STREAM_FILTERS = ["All tables", "Selected", "Not Selected"]

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
	jobName,
	onLoadingChange,
}) => {
	const prevSourceConfig = useRef(sourceConfig)
	const {
		isClearDestinationStatusLoading,
		setShowDestinationDatabaseModal,
		ingestionMode,
		setIngestionMode,
	} = useAppStore()
	const [searchText, setSearchText] = useState("")
	const [selectedFilters, setSelectedFilters] = useState<string[]>([
		"All tables",
	])
	const [activeStreamKey, setActiveStreamKey] = useState<{
		name: string
		namespace: string
	} | null>(null)

	const [apiResponse, setApiResponse] = useState<StreamsDataStructure | null>(
		initialStreamsData || null,
	)

	// Derive activeStreamData from apiResponse to always get fresh data
	const activeStreamData = useMemo(() => {
		if (!activeStreamKey || !apiResponse?.streams) return null
		return (
			apiResponse.streams.find(
				s =>
					s.stream.name === activeStreamKey.name &&
					s.stream.namespace === activeStreamKey.namespace,
			) || null
		)
	}, [activeStreamKey, apiResponse?.streams])

	const activeSelectedStream = useMemo(() => {
		if (!activeStreamKey || !apiResponse?.selected_streams) return null
		return (
			apiResponse.selected_streams[activeStreamKey.namespace || ""]?.find(
				s => s.stream_name === activeStreamKey.name,
			) || null
		)
	}, [activeStreamKey, apiResponse?.selected_streams])

	const [isStreamsLoading, setIsStreamsLoading] = useState(!initialStreamsData)
	// Store initial streams data for reference
	const [initialStreamsState, setInitialStreamsState] =
		useState(initialStreamsData)

	// Use ref to track if we've initialized to prevent double updates
	const initialized = useRef(false)

	const isStreamEnabled = (streamData: StreamData | null) => {
		if (streamData === null) return false

		const stream = apiResponse?.selected_streams[
			streamData.stream.namespace || ""
		]?.find(s => s.stream_name === streamData.stream.name)

		if (!stream) return false

		return !stream?.disabled
	}

	const isLoading = isStreamsLoading || isClearDestinationStatusLoading

	// Check if first stream has destination_database and compute values
	const { destinationDatabase, destinationDatabaseForModal } = useMemo(() => {
		if (!apiResponse?.streams || apiResponse.streams.length === 0) {
			return { destinationDatabase: null, destinationDatabaseForModal: null }
		}

		const firstStream = apiResponse.streams[0]
		const destDb = firstStream.stream?.destination_database

		if (!destDb) {
			return { destinationDatabase: null, destinationDatabaseForModal: null }
		}

		// If it's in "a:b" format
		if (destDb.includes(":")) {
			const parts = destDb.split(":")
			return {
				destinationDatabase: `${parts[0]}_${"${source_namespace}"}`, // For display
				destinationDatabaseForModal: parts[0], // For modal (just the prefix)
			}
		}

		// Otherwise use full value for both
		return {
			destinationDatabase: destDb,
			destinationDatabaseForModal: destDb,
		}
	}, [apiResponse?.streams])

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
			setIsStreamsLoading(false)
			onLoadingChange?.(false)
			initialized.current = true

			return
		}

		const fetchSourceStreams = async () => {
			if (initialized.current) return

			onLoadingChange?.(true)
			setIsStreamsLoading(true)
			try {
				const response = await sourceService.getSourceStreams(
					sourceName,
					sourceConnector,
					sourceVersion,
					sourceConfig,
					jobName,
					fromJobEditFlow ? jobId : -1,
				)

				const streamsData: StreamsDataStructure =
					getStreamsDataFromSourceStreamsResponse(
						response,
						destinationType,
						sourceConnector,
					)

				setApiResponse(streamsData)
				setSelectedStreams(streamsData)
				setInitialStreamsState(streamsData)

				initialized.current = true
			} catch (error) {
				console.error("Error fetching source streams:", error)
			} finally {
				setIsStreamsLoading(false)
				onLoadingChange?.(false)
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

	const handleStreamSyncModeChange = (
		streamName: string,
		namespace: string,
		newSyncMode: SyncMode,
		cursorField?: string,
	) => {
		setApiResponse(prev => {
			if (!prev) return prev

			const streamIndex = prev.streams.findIndex(
				s => s.stream.name === streamName && s.stream.namespace === namespace,
			)

			if (
				streamIndex !== -1 &&
				prev.streams[streamIndex].stream.sync_mode === newSyncMode &&
				(prev.streams[streamIndex].stream.cursor_field || "") ===
					(cursorField || "")
			) {
				return prev
			}

			const updated = { ...prev }

			if (streamIndex !== -1) {
				updated.streams = [...prev.streams]
				const nextStream = {
					...updated.streams[streamIndex],
					stream: {
						...updated.streams[streamIndex].stream,
						sync_mode: newSyncMode,
					},
				}

				if (cursorField !== undefined && newSyncMode === SyncMode.INCREMENTAL) {
					nextStream.stream.cursor_field = cursorField
				}
				if (newSyncMode !== SyncMode.INCREMENTAL) {
					delete nextStream.stream.cursor_field
				}

				updated.streams[streamIndex] = nextStream
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

			const existingStream = updated.selected_streams[namespace]?.find(
				s => s.stream_name === streamName,
			)
			if (checked) {
				if (!updated.selected_streams[namespace]) {
					updated.selected_streams[namespace] = []
				}
				// TODO: remove this as this case will never get executed as we are already setting defaults in streams.ts
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
				[namespace]: prev.selected_streams[namespace].map(s => {
					if (s.stream_name === streamName) {
						if (filterValue === "") {
							// Remove the 'filter' if filterValue is empty
							const updatedStream = { ...s }
							delete updatedStream.filter
							return updatedStream
						} else {
							return { ...s, filter: filterValue }
						}
					}
					return s
				}),
			}

			const updated = {
				...prev,
				selected_streams: updatedSelectedStreams,
			}

			setSelectedStreams(updated)
			return updated
		})
	}

	const handleIngestionModeChange = (
		streamName: string,
		namespace: string,
		appendMode: boolean,
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
					s.stream_name === streamName ? { ...s, append_mode: appendMode } : s,
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

	const handleAllIngestionModeChange = (ingestionMode: IngestionMode) => {
		const appendMode = ingestionMode === IngestionMode.APPEND
		setIngestionMode(ingestionMode)
		setApiResponse(prev => {
			if (!prev) return prev

			// Update all streams with the same append mode
			const updateSelectedStreams = Object.fromEntries(
				Object.entries(prev.selected_streams).map(([namespace, streams]) => [
					namespace,
					streams.map(stream => ({
						...stream,
						append_mode: appendMode,
					})),
				]),
			)

			const updated = {
				...prev,
				selected_streams: updateSelectedStreams,
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

		const showSelected = selectedFilters.includes("Selected")
		const showNotSelected = selectedFilters.includes("Not Selected")

		return tempFilteredStreams.filter(stream => {
			const ns = stream.stream.namespace || ""
			const matchingSelectedStream = apiResponse.selected_streams[ns]?.find(
				s => s.stream_name === stream.stream.name,
			)
			const isSelected =
				matchingSelectedStream && !matchingSelectedStream?.disabled
			if (showSelected && showNotSelected) return true
			if (showSelected) return isSelected
			if (showNotSelected) return !isSelected
			return false
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

	useEffect(() => {
		if (selectedFilters.length === 0) {
			setSelectedFilters(["All tables"])
		}
	}, [selectedFilters])

	// Handler for destination database modal save
	const handleDestinationDatabaseSave = (
		format: string,
		databaseName: string,
	) => {
		setApiResponse(prev => {
			if (!prev || prev.streams.length === 0) return prev

			// Check first stream to determine format for all streams
			const firstStreamDestDb = prev.streams[0].stream.destination_database
			const hasColonFormat =
				firstStreamDestDb && firstStreamDestDb.includes(":")

			const updatedStreams = prev.streams.map(stream => {
				const currentDestDb = stream.stream.destination_database
				const currentNamespace = stream.stream.namespace

				if (format === "dynamic") {
					// Dynamic format: preserve the suffix part
					if (hasColonFormat && currentDestDb) {
						// If format is "a:b", change to "c:b" (databaseName:suffix)
						const parts = currentDestDb.split(":")
						return {
							...stream,
							stream: {
								...stream.stream,
								destination_database: `${databaseName}:${parts[1]}`,
							},
						}
					} else {
						// If no ":", set to databaseName only
						// Find the stream in initial streams data to get its original namespace
						const initialStream = initialStreamsState?.streams.find(
							s =>
								s.stream.name === stream.stream.name &&
								s.stream.namespace === stream.stream.namespace,
						)

						// Get namespace from initial destination_database if it exists
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
					// Custom format: set all to databaseName
					return {
						...stream,
						stream: {
							...stream.stream,
							destination_database: databaseName,
						},
					}
				}
			})

			const updated = {
				...prev,
				streams: updatedStreams,
			}

			// Update parent component
			setSelectedStreams(updated)

			return updated
		})
	}

	const handleSelectedColumnChange = (
		streamName: string,
		namespace: string,
		selected_columns: SelectedColumns,
	) => {
		setApiResponse(prev => {
			if (!prev) return prev

			const updatedSelectedStreams = {
				...prev.selected_streams,
				[namespace]: prev.selected_streams[namespace]?.map(s =>
					s.stream_name === streamName ? { ...s, selected_columns } : s,
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

	const { Search } = Input

	return (
		<div className="mb-4 p-6">
			{stepNumber && stepTitle && (
				<StepTitle
					stepNumber={stepNumber}
					stepTitle={stepTitle}
				/>
			)}

			<div className="mb-4 mr-4 flex justify-between gap-4">
				<div className="flex w-2/6 items-center">
					<Search
						placeholder="Search Streams"
						allowClear
						className="custom-search-input w-full"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
				</div>
				<div className="flex w-4/5 justify-between gap-2">
					{destinationDatabase && (
						<div className="flex w-1/2 items-center justify-start gap-1">
							<div className="group relative rounded-md border border-neutral-disabled bg-white p-2.5 shadow-sm transition-all duration-200">
								<div className="absolute -right-2 -top-2">
									<Tooltip title={DESTINATATION_DATABASE_TOOLTIP_TEXT}>
										<div className="rounded-full bg-white p-1 shadow-sm ring-1 ring-gray-100">
											<InfoIcon className="size-4 cursor-help text-primary" />
										</div>
									</Tooltip>
								</div>

								<div className="flex items-center">
									<div className="font-medium text-gray-700">
										{destinationType === DESTINATION_INTERNAL_TYPES.S3
											? "S3 Folder"
											: "Iceberg DB"}
									</div>

									<span className="px-1">:</span>

									<div className="text-gray-600">{destinationDatabase}</div>

									<div className="ml-1 flex items-center space-x-1 border-l border-gray-200 pl-1">
										<Tooltip
											title="Edit"
											placement="top"
										>
											<PencilSimpleIcon
												className="size-4 cursor-pointer text-gray-600 transition-colors hover:text-primary"
												onClick={() => setShowDestinationDatabaseModal(true)}
											/>
										</Tooltip>
										<Tooltip title="View Documentation">
											<a
												href="https://olake.io/docs/understanding/terminologies/olake#7-tablecolumn-normalization--destination-database-creation"
												target="_blank"
												rel="noopener noreferrer"
												className="flex items-center text-gray-600 transition-colors hover:text-primary"
											>
												<ArrowSquareOutIcon className="size-4" />
											</a>
										</Tooltip>
									</div>
								</div>
							</div>
						</div>
					)}
					<div
						className={clsx(
							"flex w-1/2 flex-wrap gap-2",
							destinationDatabase ? "justify-end" : "justify-start",
						)}
					>
						{STREAM_FILTERS.map(filter => (
							<FilterButton
								key={filter}
								filter={filter}
								selectedFilters={selectedFilters}
								setSelectedFilters={setSelectedFilters}
							/>
						))}
					</div>
				</div>
			</div>

			<div className={clsx("flex", !isLoading && "rounded-[4px] border")}>
				<div
					className={clsx(
						activeStreamData ? "w-1/2" : "w-full",
						"max-h-[calc(100vh-250px)] overflow-y-auto",
					)}
				>
					{!isLoading && apiResponse?.streams ? (
						<StreamsCollapsibleList
							groupedStreams={groupedFilteredStreams}
							selectedStreams={apiResponse.selected_streams}
							setActiveStreamData={(stream: StreamData) => {
								setActiveStreamKey({
									name: stream.stream.name,
									namespace: stream.stream.namespace || "",
								})
							}}
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
								setSelectedStreams(fullData as StreamsDataStructure)
							}}
							onIngestionModeChange={handleAllIngestionModeChange}
							sourceType={sourceConnector}
							destinationType={destinationType}
						/>
					) : isLoading ? (
						<div className="flex h-[calc(100vh-250px)] items-center justify-center">
							<Spin size="large"></Spin>
						</div>
					) : (
						<Empty className="flex h-full flex-col items-center justify-center" />
					)}
				</div>

				<div
					className={clsx(
						"sticky top-0 flex w-1/2 flex-col rounded-[4px] bg-white p-4 transition-all duration-150 ease-linear",
						!isLoading && "border-l",
					)}
				>
					{activeStreamData && activeSelectedStream ? (
						<StreamConfiguration
							stream={activeStreamData}
							onUpdate={() => {
								// Update the stream config in the local state
								// Implementation will be added later if needed
							}}
							onSyncModeChange={(
								streamName: string,
								namespace: string,
								syncMode: SyncMode,
								cursorField?: string,
							) => {
								handleStreamSyncModeChange(
									streamName,
									namespace,
									syncMode,
									cursorField,
								)
							}}
							useDirectForms={useDirectForms}
							isSelected={isStreamEnabled(activeStreamData)}
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
							onIngestionModeChange={handleIngestionModeChange}
							sourceType={sourceConnector}
							onSelectedColumnChange={handleSelectedColumnChange}
							selectedStream={activeSelectedStream}
						/>
					) : null}
				</div>
			</div>

			{/* Destination Database Modal */}
			<DestinationDatabaseModal
				destinationType={destinationType || ""}
				destinationDatabase={destinationDatabaseForModal}
				allStreams={apiResponse}
				onSave={handleDestinationDatabaseSave}
				originalDatabase={destinationDatabase || ""}
				initialStreams={initialStreamsState || null}
			/>
		</div>
	)
}

export default SchemaConfiguration
