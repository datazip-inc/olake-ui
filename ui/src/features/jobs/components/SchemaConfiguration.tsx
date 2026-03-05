import React, { useEffect, useState, useMemo } from "react"
import { Input, Empty, Spin, Tooltip } from "antd"
import clsx from "clsx"

import { useShallow } from "zustand/react/shallow"
import { useJobStore, useStreamSelectionStore } from "../stores"
import {
	selectActiveStreamData,
	selectDestinationDatabase,
} from "../stores/streamSelectionStore"
import { useDiscoverSourceStreams } from "@/features/sources/hooks/mutations/useDiscoverSourceStreams"
import { useClearDestinationStatus } from "../hooks/queries/useJobQueries"
import { StreamData, StreamsDataStructure } from "../../../common/types"
import { SchemaConfigurationProps } from "../types"
import FilterButton from "./FilterButton"
import StepTitle from "@/common/components/StepTitle"
import StreamsCollapsibleList from "./streams/StreamsCollapsibleList"
import StreamConfiguration from "./streams/StreamConfiguration"
import {
	ArrowSquareOutIcon,
	InfoIcon,
	PencilSimpleIcon,
} from "@phosphor-icons/react"
import { DESTINATATION_DATABASE_TOOLTIP_TEXT } from "../constants"
import { DESTINATION_INTERNAL_TYPES } from "@/common/constants/constants"
import DestinationDatabaseModal from "@/features/jobs/components/modals/DestinationDatabaseModal"
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

	const store = useStreamSelectionStore()

	const discoverMutation = useDiscoverSourceStreams()

	const { data: clearDestStatus, isLoading: isClearDestinationStatusLoading } =
		useClearDestinationStatus(jobId >= 0 ? jobId.toString() : "", {
			staleTime: Infinity,
		})
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

	// Discovery / initialization effect
	useEffect(() => {
		if (!sourceConfig || !sourceConnector || discoverMutation.isPending) return

		store.setDiscovering(true)

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
					const streamsData: StreamsDataStructure =
						getStreamsDataFromSourceStreamsResponse(
							response,
							destinationType,
							sourceConnector,
							sourceVersion,
						)

					store.initializeFromDiscovery(streamsData)
				},
				onError: error => {
					store.setDiscoverError(error)
				},
			},
		)
	}, [sourceName, sourceConnector, sourceVersion, sourceConfig])

	// Reset selectedFilters to "All tables" when all filters are deselected
	useEffect(() => {
		if (selectedFilters.length === 0) {
			setSelectedFilters(["All tables"])
		}
	}, [selectedFilters])

	// Reset store on unmount
	useEffect(() => {
		return () => {
			store.reset()
		}
	}, [])

	const isLoading = store.isDiscovering || isClearDestinationStatusLoading

	const activeStreamData = useStreamSelectionStore(selectActiveStreamData)
	// useShallow: selector returns a new object literal each time; shallow comparison
	// prevents re-renders when display/forModal values haven't actually changed.
	const {
		display: destinationDatabase,
		forModal: destinationDatabaseForModal,
	} = useStreamSelectionStore(useShallow(selectDestinationDatabase))

	const filteredStreams = useMemo(() => {
		if (!store.streamsData?.streams) return []
		let tempFilteredStreams = [...store.streamsData.streams]

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
			const matchingSelectedStream = store.streamsData!.selected_streams[
				ns
			]?.find(s => s.stream_name === stream.stream.name)
			const isSelected =
				matchingSelectedStream && !matchingSelectedStream?.disabled
			if (showSelected && showNotSelected) return true
			if (showSelected) return isSelected
			if (showNotSelected) return !isSelected
			return false
		})
	}, [store.streamsData, searchText, selectedFilters])

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
					{!isLoading && store.streamsData?.streams ? (
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
						<Empty className="flex h-full flex-col items-center justify-center" />
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
				allStreams={store.streamsData}
				onSave={(format: string, databaseName: string) => {
					store.updateDestinationDatabase(format, databaseName)
				}}
				originalDatabase={destinationDatabase || ""}
				initialStreams={store.initialStreamsSnapshot}
			/>
		</div>
	)
}

export default SchemaConfiguration
