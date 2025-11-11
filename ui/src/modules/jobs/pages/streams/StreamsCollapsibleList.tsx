import { useEffect, useRef, useState } from "react"
import { CaretDownIcon, CaretRightIcon } from "@phosphor-icons/react"
import { Checkbox, Empty } from "antd"
import clsx from "clsx"

import {
	GroupedStreamsCollapsibleListProps,
	StreamData,
} from "../../../../types"
import StreamPanel from "./StreamPanel"
import { useAppStore } from "../../../../store"
import { IngestionMode } from "../../../../types/commonTypes"
import IngestionModeChangeModal from "../../../common/Modals/IngestionModeChangeModal"
import { getIngestionMode } from "../../../../utils/utils"

const StreamsCollapsibleList = ({
	groupedStreams,
	selectedStreams,
	setActiveStreamData,
	activeStreamData,
	onStreamSelect,
	onIngestionModeChange,
}: GroupedStreamsCollapsibleListProps) => {
	const { setShowIngestionModeChangeModal, ingestionMode, setIngestionMode } =
		useAppStore()
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

	useEffect(() => {
		setIngestionMode(getIngestionMode(selectedStreams))
	}, [selectedStreams])

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

	const dataHasChanged = () => {
		const prev = prevGroupedStreams.current
		const current = groupedStreams

		const prevKeys = Object.keys(prev)
		const currentKeys = Object.keys(current)

		if (prevKeys.length !== currentKeys.length) return true

		for (const key of currentKeys) {
			if (!prev[key]) return true
			if (prev[key].length !== current[key].length) return true
		}

		return false
	}

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
				const streamName = stream.stream.name
				let isStreamSelected = false

				if ("selected_streams" in selectedStreams) {
					// CombinedStreamsData format
					const selectedStream = (selectedStreams as any).selected_streams[
						ns
					]?.find(
						(selected: { stream_name: string; disabled?: boolean }) =>
							selected.stream_name === streamName,
					)
					isStreamSelected = !!(selectedStream && !selectedStream.disabled)
				} else {
					// Regular mapping format
					const selectedStream = (selectedStreams as any)[ns]?.find(
						(selected: { stream_name: string; disabled?: boolean }) =>
							selected.stream_name === streamName,
					)
					isStreamSelected = !!(selectedStream && !selectedStream.disabled)
				}

				newCheckedStatus.streams[ns][streamName] = isStreamSelected
			})

			// Check if all streams in this namespace are selected
			const allStreamsInNamespaceSelected = streams.every(
				stream => newCheckedStatus.streams[ns][stream.stream.name],
			)
			newCheckedStatus.namespaces[ns] = allStreamsInNamespaceSelected
		})

		// Check if all namespaces are selected
		const allNamespacesSelected = Object.values(
			newCheckedStatus.namespaces,
		).every(value => value)
		newCheckedStatus.global = allNamespacesSelected

		setCheckedStatus(newCheckedStatus)

		// sort the namespaces and streams inside it alphabetically on the basis of checked and unchecked status
		if (sortedGroupedNamespaces.length === 0 || dataHasChanged()) {
			const sortStreamsByCheckedStatus = (
				streams: StreamData[],
				namespace: string,
			): StreamData[] => {
				const checked: StreamData[] = []
				const unchecked: StreamData[] = []

				streams.forEach(stream => {
					const isChecked =
						newCheckedStatus.streams[namespace]?.[stream.stream.name]
					if (isChecked) checked.push(stream)
					else unchecked.push(stream)
				})

				const sortByStreamName = (a: StreamData, b: StreamData) =>
					a.stream.name.localeCompare(b.stream.name)

				checked.sort(sortByStreamName)
				unchecked.sort(sortByStreamName)

				return [...checked, ...unchecked]
			}

			const namespacesWithCheckedStreams: [string, StreamData[]][] = []
			const namespacesWithoutCheckedStreams: [string, StreamData[]][] = []

			Object.entries(groupedStreams).forEach(([namespace, streams]) => {
				const hasAnySelectedStream = streams.some(
					stream => newCheckedStatus.streams[namespace]?.[stream.stream.name],
				)

				const sortedStreams = sortStreamsByCheckedStatus(streams, namespace)

				if (hasAnySelectedStream)
					namespacesWithCheckedStreams.push([namespace, sortedStreams])
				else namespacesWithoutCheckedStreams.push([namespace, sortedStreams])
			})

			const sortByNamespaceName = (
				a: [string, StreamData[]],
				b: [string, StreamData[]],
			) => a[0].localeCompare(b[0])

			namespacesWithCheckedStreams.sort(sortByNamespaceName)
			namespacesWithoutCheckedStreams.sort(sortByNamespaceName)

			setSortedGroupedNamespaces([
				...namespacesWithCheckedStreams,
				...namespacesWithoutCheckedStreams,
			])

			prevGroupedStreams.current = groupedStreams
		}
	}, [selectedStreams, groupedStreams])

	const handleToggleNamespace = (ns: string) => {
		setOpenNamespaces(prev => ({ ...prev, [ns]: !prev[ns] }))
	}

	const handleNamespaceSyncAll = (ns: string, checked: boolean) => {
		// Force a state update by manipulating the localStorage first
		try {
			// Store the namespace we're trying to modify in localStorage
			localStorage.setItem("__lastToggledNamespace", ns)
			localStorage.setItem("__lastToggledNamespaceState", String(checked))
		} catch {
			// Ignore any localStorage errors
		}

		// Update local checked status
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

		if (!checked) {
			// For unchecking, call the original onStreamSelect for each stream in this namespace
			const streamsInNamespace = groupedStreams[ns] || []

			streamsInNamespace.forEach(streamData => {
				onStreamSelect(streamData.stream.name, false, ns)
			})
		} else {
			const streamsInNamespace = groupedStreams[ns] || []

			streamsInNamespace.forEach(streamData => {
				onStreamSelect(streamData.stream.name, true, ns)
			})
		}
	}

	const handleGlobalSyncAll = (checked: boolean) => {
		try {
			localStorage.setItem("__globalSync", String(checked))
		} catch {}

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

		if (!checked) {
			Object.keys(groupedStreams).forEach(ns => {
				handleNamespaceSyncAll(ns, false)
			})
		} else {
			Object.entries(groupedStreams).forEach(([ns, streams]) => {
				streams.forEach(streamData => {
					onStreamSelect(streamData.stream.name, true, ns)
				})
			})
		}
	}

	const handleStreamSelect = (
		streamName: string,
		checked: boolean,
		ns: string,
	) => {
		setCheckedStatus(prev => {
			const updatedStreams = {
				...prev.streams,
				[ns]: {
					...prev.streams[ns],
					[streamName]: checked,
				},
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

		onStreamSelect(streamName, checked, ns)
	}

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

							<div className="relative flex rounded-[4px] bg-[#F5F5F5] text-sm text-black">
								{/* Sliding background */}
								<div
									className={clsx(
										"absolute inset-y-0.5 w-[calc(34%)] rounded-sm bg-primary-100 shadow-sm transition-transform duration-300 ease-in-out",
										{
											"translate-x-0.5": ingestionMode === IngestionMode.UPSERT,
											"translate-x-[calc(100%+0px)]":
												ingestionMode === IngestionMode.APPEND,
											"translate-x-[calc(200%-2px)]":
												ingestionMode === IngestionMode.CUSTOM,
										},
									)}
								/>
								<div
									onClick={() => {
										if (ingestionMode !== IngestionMode.UPSERT) {
											setTargetIngestionMode(IngestionMode.UPSERT)
											setShowIngestionModeChangeModal(true)
										}
									}}
									className={`relative z-10 flex cursor-pointer items-center justify-center rounded-sm p-1 px-4 text-center transition-colors duration-300`}
								>
									All Upsert
								</div>
								<div
									onClick={() => {
										if (ingestionMode !== IngestionMode.APPEND) {
											setTargetIngestionMode(IngestionMode.APPEND)
											setShowIngestionModeChangeModal(true)
										}
									}}
									className={`relative z-10 flex cursor-pointer items-center justify-center rounded-sm p-1 px-4 text-center transition-colors duration-300`}
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
						</div>
						{sortedGroupedNamespaces.map(([ns, streams]) => {
							return (
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
											onChange={e =>
												handleNamespaceSyncAll(ns, e.target.checked)
											}
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
											{streams.map(streamData => (
												<StreamPanel
													stream={streamData}
													key={streamData?.stream?.name}
													activeStreamData={activeStreamData}
													setActiveStreamData={setActiveStreamData}
													onStreamSelect={(streamName, checked) =>
														handleStreamSelect(streamName, checked, ns)
													}
													isSelected={
														checkedStatus.streams[ns]?.[
															streamData.stream.name
														] || false
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
			<IngestionModeChangeModal
				ingestionMode={targetIngestionMode}
				onConfirm={onIngestionModeChange}
			/>
		</>
	)
}

export default StreamsCollapsibleList
