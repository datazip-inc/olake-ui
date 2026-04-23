import { CaretDownIcon, CaretRightIcon } from "@phosphor-icons/react"
import { Checkbox, Empty } from "antd"
import clsx from "clsx"
import { useEffect, useRef, useState } from "react"

import { StreamData } from "@/modules/ingestion/common/types"

import StreamPanel from "./StreamPanel"
import { IngestionMode } from "../../enums"
import {
	useJobStore,
	useStreamSelectionStore,
	selectSelectedStreams,
} from "../../stores"
import { GroupedStreamsCollapsibleListProps } from "../../types"
import {
	getIngestionMode,
	hasGroupedStreamsStructureChanged,
	isDestinationIngestionModeSupported,
	isSourceIngestionModeSupported,
	sortGroupedStreamsByCheckedState,
} from "../../utils/streams"
import IngestionModeChangeModal from "../modals/IngestionModeChangeModal"

const StreamsCollapsibleList = ({
	groupedStreams,
	sourceType,
	destinationType,
}: GroupedStreamsCollapsibleListProps) => {
	const selectedStreams = useStreamSelectionStore(selectSelectedStreams)
	const toggleStream = useStreamSelectionStore(state => state.toggleStream)
	const updateAllIngestionMode = useStreamSelectionStore(
		state => state.updateAllIngestionMode,
	)
	const setActiveStreamKey = useStreamSelectionStore(
		state => state.setActiveStreamKey,
	)
	const { setShowIngestionModeChangeModal, ingestionMode, setIngestionMode } =
		useJobStore()
	const [openNamespaces, setOpenNamespaces] = useState<{
		[ns: string]: boolean
	}>({})
	const [checkedStatus, setCheckedStatus] = useState<{
		global: boolean
		namespaces: { [ns: string]: boolean }
		streams: { [ns: string]: { [streamName: string]: boolean } }
	}>({
		global: false,
		namespaces: {},
		streams: {},
	})
	const [targetIngestionMode, setTargetIngestionMode] = useState<IngestionMode>(
		IngestionMode.APPEND,
	)
	const [sortedGroupedNamespaces, setSortedGroupedNamespaces] = useState<
		[string, StreamData[]][]
	>([])

	const prevGroupedStreams = useRef(groupedStreams)
	const prevBulkApplyVersion = useRef(0)

	const bulkApplyVersion = useStreamSelectionStore(
		state => state.bulkApplyVersion,
	)

	useEffect(() => {
		setIngestionMode(getIngestionMode(selectedStreams, sourceType))
	}, [selectedStreams, sourceType])

	// Keep all namespaces expanded by default and automatically open any newly added namespaces.
	useEffect(() => {
		if (Object.keys(openNamespaces).length === 0) {
			const allOpen: { [ns: string]: boolean } = {}
			Object.keys(groupedStreams).forEach((ns: string) => {
				allOpen[ns] = true
			})
			setOpenNamespaces(allOpen)
		} else {
			setOpenNamespaces(prev => {
				const updated = { ...prev }
				Object.keys(groupedStreams).forEach((ns: string) => {
					if (updated[ns] === undefined) {
						updated[ns] = true
					}
				})
				return updated
			})
		}
	}, [groupedStreams])

	// Update local checked status based on selectedStreams
	useEffect(() => {
		const newCheckedStatus = {
			global: false,
			namespaces: {} as { [ns: string]: boolean },
			streams: {} as { [ns: string]: { [streamName: string]: boolean } },
		}

		// Initialize streams checked status
		Object.entries(groupedStreams).forEach(([ns, streams]) => {
			newCheckedStatus.streams[ns] = {}

			// For each stream, check if it is selected
			streams.forEach(stream => {
				const selectedStream = selectedStreams[ns]?.find(
					s => s.stream_name === stream.stream.name,
				)
				newCheckedStatus.streams[ns][stream.stream.name] = !!(
					selectedStream && !selectedStream.disabled
				)
			})

			newCheckedStatus.namespaces[ns] = streams.every(
				s => newCheckedStatus.streams[ns][s.stream.name],
			)
		})

		newCheckedStatus.global = Object.values(newCheckedStatus.namespaces).every(
			Boolean,
		)

		setCheckedStatus(newCheckedStatus)

		// Sort namespaces and streams alphabetically by checked/unchecked status.
		// Re-sort on first render, structure change, or after a bulk apply.
		const isBulkApply = bulkApplyVersion !== prevBulkApplyVersion.current
		if (
			sortedGroupedNamespaces.length === 0 ||
			hasGroupedStreamsStructureChanged(
				prevGroupedStreams.current,
				groupedStreams,
			) ||
			isBulkApply
		) {
			setSortedGroupedNamespaces(
				sortGroupedStreamsByCheckedState(
					groupedStreams,
					newCheckedStatus.streams,
				),
			)
			prevGroupedStreams.current = groupedStreams
			prevBulkApplyVersion.current = bulkApplyVersion
		}
	}, [selectedStreams, groupedStreams, bulkApplyVersion])

	// Auto-select first stream once sortedGroupedNamespaces is populated
	useEffect(() => {
		const activeKey = useStreamSelectionStore.getState().activeStreamKey
		if (
			!activeKey &&
			sortedGroupedNamespaces.length > 0 &&
			sortedGroupedNamespaces[0][1].length > 0
		) {
			const first = sortedGroupedNamespaces[0][1][0]
			setActiveStreamKey({
				streamName: first.stream.name,
				namespace: first.stream.namespace ?? "",
			})
		}
	}, [sortedGroupedNamespaces, setActiveStreamKey])

	const handleToggleNamespace = (ns: string) => {
		setOpenNamespaces(prev => ({ ...prev, [ns]: !prev[ns] }))
	}

	const handleNamespaceSyncAll = (ns: string, checked: boolean) => {
		setCheckedStatus(prev => ({
			...prev,
			namespaces: { ...prev.namespaces, [ns]: checked },
			streams: {
				...prev.streams,
				[ns]: Object.keys(prev.streams[ns] || {}).reduce(
					(acc, streamName) => {
						acc[streamName] = checked
						return acc
					},
					{} as { [streamName: string]: boolean },
				),
			},
		}))

		const streamsInNamespace = groupedStreams[ns] || []
		streamsInNamespace.forEach(streamData => {
			toggleStream(
				{
					streamName: streamData.stream.name,
					namespace: ns,
				},
				checked,
				ingestionMode,
			)
		})
	}

	const handleGlobalSyncAll = (checked: boolean) => {
		setCheckedStatus(prev => {
			const updatedNamespaces = { ...prev.namespaces }
			const updatedStreams = { ...prev.streams }
			Object.keys(groupedStreams).forEach(ns => {
				updatedNamespaces[ns] = checked
				updatedStreams[ns] = groupedStreams[ns].reduce(
					(acc, stream) => {
						acc[stream.stream.name] = checked
						return acc
					},
					{} as { [streamName: string]: boolean },
				)
			})
			return {
				global: checked,
				namespaces: updatedNamespaces,
				streams: updatedStreams,
			}
		})

		Object.entries(groupedStreams).forEach(([ns, streams]) => {
			streams.forEach(streamData => {
				toggleStream(
					{
						streamName: streamData.stream.name,
						namespace: ns,
					},
					checked,
					ingestionMode,
				)
			})
		})
	}

	const handleStreamSelect = (
		streamName: string,
		checked: boolean,
		ns: string,
	) => {
		setCheckedStatus(prev => {
			const updatedStreams = {
				...prev.streams,
				[ns]: { ...prev.streams[ns], [streamName]: checked },
			}

			const allStreamsInNamespaceSelected = groupedStreams[ns].every(
				stream => updatedStreams[ns][stream.stream.name],
			)

			const updatedNamespaces = {
				...prev.namespaces,
				[ns]: allStreamsInNamespaceSelected,
			}

			const allNamespacesSelected = Object.keys(groupedStreams).every(
				namespace => updatedNamespaces[namespace],
			)

			return {
				global: allNamespacesSelected,
				namespaces: updatedNamespaces,
				streams: updatedStreams,
			}
		})

		toggleStream(
			{
				streamName,
				namespace: ns,
			},
			checked,
			ingestionMode,
		)
	}

	const isSourceUpsertModeSupported = isSourceIngestionModeSupported(
		IngestionMode.UPSERT,
		sourceType,
	)

	const isSourceAppendModeSupported = isSourceIngestionModeSupported(
		IngestionMode.APPEND,
		sourceType,
	)

	const isDestUpsertModeSupported = isDestinationIngestionModeSupported(
		IngestionMode.UPSERT,
		destinationType,
	)

	return (
		<>
			<div className="flex h-full flex-col rounded-[4px] border-gray-200">
				{Object.keys(sortedGroupedNamespaces).length === 0 ? (
					<Empty className="pt-10" />
				) : (
					<>
						<div className="flex items-center justify-between rounded-t-[4px] bg-white px-2 py-4">
							<Checkbox
								checked={checkedStatus.global}
								onChange={e => handleGlobalSyncAll(e.target.checked)}
							>
								Sync all
							</Checkbox>

							{isDestUpsertModeSupported && (
								<div className="relative flex rounded-[4px] bg-[#F5F5F5] text-sm text-black">
									{/* Sliding background */}
									<div
										className={clsx(
											"absolute inset-y-0.5 w-[calc(34%)] rounded-sm bg-primary-100 shadow-sm transition-transform duration-300 ease-in-out",
											{
												"translate-x-0.5":
													ingestionMode === IngestionMode.UPSERT,
												"translate-x-[calc(100%+0px)]":
													ingestionMode === IngestionMode.APPEND,
												"translate-x-[calc(200%-2px)]":
													ingestionMode === IngestionMode.CUSTOM,
											},
										)}
									/>
									<div
										onClick={() => {
											if (
												ingestionMode !== IngestionMode.UPSERT &&
												isSourceUpsertModeSupported
											) {
												setTargetIngestionMode(IngestionMode.UPSERT)
												setShowIngestionModeChangeModal(true)
											}
										}}
										className={clsx(
											`relative z-10 flex items-center justify-center rounded-sm p-1 px-4 text-center transition-colors duration-300`,
											isSourceUpsertModeSupported
												? "cursor-pointer"
												: "cursor-not-allowed opacity-40",
										)}
									>
										All Upsert
									</div>
									<div
										onClick={() => {
											if (
												ingestionMode !== IngestionMode.APPEND &&
												isSourceAppendModeSupported
											) {
												setTargetIngestionMode(IngestionMode.APPEND)
												setShowIngestionModeChangeModal(true)
											}
										}}
										className={clsx(
											`relative z-10 flex cursor-pointer items-center justify-center rounded-sm p-1 px-4 text-center transition-colors duration-300`,
											isSourceAppendModeSupported
												? "cursor-pointer"
												: "cursor-not-allowed opacity-40",
										)}
									>
										All Append
									</div>
									<div
										className={clsx(
											"relative z-10 flex items-center justify-center rounded-sm p-1 px-5 text-center transition-colors duration-300",
											ingestionMode === IngestionMode.CUSTOM
												? "cursor-default"
												: "cursor-not-allowed opacity-40",
										)}
									>
										Custom
									</div>
								</div>
							)}
						</div>
						{sortedGroupedNamespaces.map(([ns, streams]) => (
							<div
								key={ns}
								className="border-gray-200"
							>
								<div
									className="flex cursor-pointer items-center border bg-background-primary p-3"
									onClick={() => handleToggleNamespace(ns)}
								>
									<Checkbox
										checked={checkedStatus.namespaces[ns]}
										onChange={e => handleNamespaceSyncAll(ns, e.target.checked)}
										onClick={e => e.stopPropagation()}
										className="mr-2"
									/>
									<span className="font-semibold">{ns}</span>
									<span className="ml-auto">
										{openNamespaces[ns] ? (
											<CaretDownIcon
												className="size-4"
												weight="fill"
											/>
										) : (
											<CaretRightIcon
												className="size-4"
												weight="fill"
											/>
										)}
									</span>
								</div>
								{openNamespaces[ns] && (
									<div className="w-full">
										{streams.map(streamData => {
											const isSelected =
												checkedStatus.streams[ns]?.[streamData.stream.name] ||
												false
											return (
												<StreamPanel
													stream={streamData}
													key={streamData?.stream?.name}
													onStreamSelect={(streamName, checked) =>
														handleStreamSelect(streamName, checked, ns)
													}
													isSelected={isSelected}
												/>
											)
										})}
									</div>
								)}
							</div>
						))}
					</>
				)}
			</div>
			<IngestionModeChangeModal
				ingestionMode={targetIngestionMode}
				onConfirm={(mode: IngestionMode) => {
					const appendMode = mode === IngestionMode.APPEND
					setIngestionMode(mode)
					updateAllIngestionMode(appendMode)
				}}
			/>
		</>
	)
}

export default StreamsCollapsibleList
