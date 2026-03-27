import {
	ArrowSquareOutIcon,
	InfoIcon,
	PencilSimpleIcon,
} from "@phosphor-icons/react"
import { useIsFetching } from "@tanstack/react-query"
import { Input, Empty, Spin, Tooltip, Button } from "antd"
import clsx from "clsx"
import React, { useEffect, useState, useMemo } from "react"
import { useShallow } from "zustand/react/shallow"

import StepTitle from "@/modules/ingestion/common/components/StepTitle"
import { DESTINATION_INTERNAL_TYPES } from "@/modules/ingestion/common/constants"
import {
	StreamData,
	StreamsDataStructure,
} from "@/modules/ingestion/common/types"
import { useDiscoverSourceStreams } from "@/modules/ingestion/features/sources/hooks"

import { DESTINATATION_DATABASE_TOOLTIP_TEXT, jobsKeys } from "../constants"
import { useClearDestinationStatus } from "../hooks"
import { useJobStore, useStreamSelectionStore } from "../stores"
import {
	selectActiveStreamData,
	selectDestinationDatabase,
	selectStreamsData,
	selectIsDiscovering,
	selectInitialStreamsSnapshot,
} from "../stores/streamSelectionStore"
import { SchemaConfigurationProps } from "../types"
import FilterButton from "./FilterButton"
import DestinationDatabaseModal from "./modals/DestinationDatabaseModal"
import StreamConfiguration from "./streams/StreamConfiguration"
import StreamsCollapsibleList from "./streams/StreamsCollapsibleList"
import { getStreamsDataFromSourceStreamsResponse } from "../utils/streams"

const STREAM_FILTERS = ["All tables", "Selected", "Not Selected"]

const SchemaConfiguration: React.FC<SchemaConfigurationProps> = ({
	stepNumber = 3,
	stepTitle = "Streams Selection",
	sourceName,
	sourceConnector,
	sourceVersion,
	sourceConfig,
	fromJobEditFlow = false,
	jobId = -1,
	destinationType,
	jobName,
	advancedSettings,
}) => {
	const { setShowDestinationDatabaseModal, setShowStreamEditDisabledModal } =
		useJobStore()

	const streamsData = useStreamSelectionStore(selectStreamsData)
	const isDiscovering = useStreamSelectionStore(selectIsDiscovering)
	const discoverError = useStreamSelectionStore(state => state.discoverError)
	const initialStreamsSnapshot = useStreamSelectionStore(
		selectInitialStreamsSnapshot,
	)

	const discoverMutation = useDiscoverSourceStreams()

	const { data: clearDestStatus, isLoading: isClearDestinationStatusLoading } =
		useClearDestinationStatus(jobId)

	const isJobFetching =
		useIsFetching({ queryKey: jobsKeys.detail(jobId.toString()) }) > 0

	// Show stream-edit-disabled modal when clear destination is running
	useEffect(() => {
		if (clearDestStatus?.running) {
			setShowStreamEditDisabledModal(true)
		}
	}, [clearDestStatus?.running])
	const [searchText, setSearchText] = useState("")
	const [selectedFilters, setSelectedFilters] = useState<string[]>([
		"All tables",
	])

	const triggerStreamsDiscovery = () => {
		if (
			!sourceConfig ||
			!sourceConnector ||
			!sourceVersion ||
			!jobName ||
			discoverMutation.isPending
		)
			return

		const { setDiscovering, initializeFromDiscovery, setDiscoverError } =
			useStreamSelectionStore.getState()
		setDiscovering(true)
		setDiscoverError(null)

		discoverMutation.mutate(
			{
				name: sourceName,
				type: sourceConnector,
				version: sourceVersion,
				config: sourceConfig,
				job_name: jobName,
				job_id: fromJobEditFlow ? jobId : -1,
				max_discover_threads: advancedSettings?.max_discover_threads,
			},
			{
				onSuccess: response => {
					const data: StreamsDataStructure =
						getStreamsDataFromSourceStreamsResponse(
							response,
							destinationType,
							sourceConnector,
							sourceVersion,
						)

					initializeFromDiscovery(data)
				},
				onError: error => {
					setDiscovering(false)
					setDiscoverError(error)
				},
			},
		)
	}

	// Discovery / initialization effect
	useEffect(() => {
		triggerStreamsDiscovery()
	}, [sourceName, sourceConnector, sourceVersion, sourceConfig, jobName])

	// Reset selectedFilters to "All tables" when all filters are deselected
	useEffect(() => {
		if (selectedFilters.length === 0) {
			setSelectedFilters(["All tables"])
		}
	}, [selectedFilters])

	// Reset store on unmount
	useEffect(() => {
		return () => {
			discoverMutation.cancel()
			discoverMutation.reset()
			useStreamSelectionStore.getState().reset()
			setShowStreamEditDisabledModal(false)
		}
	}, [])

	const isLoading =
		isJobFetching ||
		isDiscovering ||
		isClearDestinationStatusLoading ||
		(!!sourceConfig &&
			!!sourceConnector &&
			!!sourceVersion &&
			!!jobName &&
			!discoverError &&
			!streamsData?.streams)

	const activeStreamData = useStreamSelectionStore(selectActiveStreamData)
	// useShallow: selector returns a new object literal each time; shallow comparison
	// prevents re-renders when display/forModal values haven't actually changed.
	const {
		display: destinationDatabase,
		forModal: destinationDatabaseForModal,
	} = useStreamSelectionStore(useShallow(selectDestinationDatabase))

	const filteredStreams = useMemo(() => {
		if (!streamsData?.streams) return []
		let tempFilteredStreams = [...streamsData.streams]

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
			const matchingSelectedStream = streamsData!.selected_streams[ns]?.find(
				s => s.stream_name === stream.stream.name,
			)
			const isSelected =
				matchingSelectedStream && !matchingSelectedStream?.disabled
			if (showSelected && showNotSelected) return true
			if (showSelected) return isSelected
			if (showNotSelected) return !isSelected
			return false
		})
	}, [streamsData, searchText, selectedFilters])

	const groupedFilteredStreams = useMemo(() => {
		const grouped: { [namespace: string]: StreamData[] } = {}
		filteredStreams.forEach(stream => {
			const ns = stream.stream.namespace || ""
			if (!grouped[ns]) grouped[ns] = []
			grouped[ns].push(stream)
		})
		return grouped
	}, [filteredStreams])

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
						<div className="flex w-1/2 min-w-0 items-center justify-start gap-1">
							<div className="group relative w-full min-w-0 rounded-md border border-neutral-disabled bg-white p-2.5 shadow-sm transition-all duration-200">
								<div className="absolute -right-2 -top-2">
									<Tooltip title={DESTINATATION_DATABASE_TOOLTIP_TEXT}>
										<div className="rounded-full bg-white p-1 shadow-sm ring-1 ring-gray-100">
											<InfoIcon className="size-4 cursor-help text-primary" />
										</div>
									</Tooltip>
								</div>

								<div className="flex min-w-0 items-center">
									<div className="shrink-0 whitespace-nowrap font-medium text-gray-700">
										{destinationType === DESTINATION_INTERNAL_TYPES.S3
											? "S3 Folder"
											: "Iceberg DB"}
									</div>

									<span className="shrink-0 px-1">:</span>

									<Tooltip title={destinationDatabase}>
										<div className="min-w-0 truncate text-gray-600">
											{destinationDatabase}
										</div>
									</Tooltip>

									<div className="ml-1 flex shrink-0 items-center space-x-1 border-l border-gray-200 pl-1">
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
					{!isLoading && streamsData?.streams ? (
						<StreamsCollapsibleList
							groupedStreams={groupedFilteredStreams}
							sourceType={sourceConnector}
							destinationType={destinationType}
						/>
					) : isLoading ? (
						<div className="flex h-[calc(100vh-250px)] items-center justify-center">
							<Spin size="large" />
						</div>
					) : (
						<div className="flex h-full flex-col items-center justify-center gap-3 py-8">
							<Empty />
							{!!discoverError && (
								<Button
									type="primary"
									onClick={triggerStreamsDiscovery}
									loading={discoverMutation.isPending}
								>
									Retry
								</Button>
							)}
						</div>
					)}
				</div>

				<div
					className={clsx(
						"sticky top-0 flex w-1/2 flex-col rounded-[4px] bg-white p-4 transition-all duration-150 ease-linear",
						!isLoading && "border-l",
					)}
				>
					{activeStreamData ? (
						<StreamConfiguration
							destinationType={destinationType}
							sourceType={sourceConnector}
						/>
					) : null}
				</div>
			</div>

			{/* Destination Database Modal */}
			<DestinationDatabaseModal
				destinationType={destinationType || ""}
				destinationDatabase={destinationDatabaseForModal}
				allStreams={streamsData}
				onSave={(format: string, databaseName: string) => {
					useStreamSelectionStore
						.getState()
						.updateDestinationDatabase(format, databaseName)
				}}
				originalDatabase={destinationDatabase || ""}
				initialStreams={initialStreamsSnapshot}
			/>
		</div>
	)
}

export default SchemaConfiguration
