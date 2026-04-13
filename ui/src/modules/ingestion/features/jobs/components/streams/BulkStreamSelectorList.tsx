import { CaretDownIcon, CaretRightIcon } from "@phosphor-icons/react"
import { Checkbox, Empty } from "antd"
import { useEffect, useMemo, useRef, useState } from "react"

import {
	StreamData,
	StreamIdentifier,
	StreamsDataStructure,
} from "@/modules/ingestion/common/types"

import {
	hasGroupedStreamsStructureChanged,
	sortGroupedStreamsByCheckedState,
} from "../../utils/streams"

interface BulkStreamSelectorListProps {
	streamsData: StreamsDataStructure | null
	bulkSelectedStreams: StreamIdentifier[]
	onChange: (streams: StreamIdentifier[]) => void
}

const BulkStreamSelectorList = ({
	streamsData,
	bulkSelectedStreams,
	onChange,
}: BulkStreamSelectorListProps) => {
	const [openNamespaces, setOpenNamespaces] = useState<{
		[ns: string]: boolean
	}>({})
	const [sortedGroups, setSortedGroups] = useState<[string, StreamData[]][]>([])
	const prevGroupedStreams = useRef<Record<string, StreamData[]>>({})

	const groupedStreams = useMemo(() => {
		const groups: Record<string, StreamData[]> = {}
		streamsData?.streams?.forEach(streamData => {
			const ns = streamData.stream.namespace || ""
			if (!groups[ns]) groups[ns] = []
			groups[ns].push(streamData)
		})
		return groups
	}, [streamsData?.streams])

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

	const checkedStatus = useMemo(() => {
		const status = {
			global: false,
			namespaces: {} as { [ns: string]: boolean },
			streams: {} as { [ns: string]: { [streamName: string]: boolean } },
		}

		let totalStreams = 0
		let totalCheckedStreams = 0

		Object.entries(groupedStreams).forEach(([ns, streams]) => {
			status.streams[ns] = {}
			let nsCheckedCount = 0

			streams.forEach(stream => {
				totalStreams++
				const isChecked = bulkSelectedStreams.some(
					s => s.namespace === ns && s.streamName === stream.stream.name,
				)
				status.streams[ns][stream.stream.name] = isChecked
				if (isChecked) {
					nsCheckedCount++
					totalCheckedStreams++
				}
			})

			status.namespaces[ns] =
				streams.length > 0 && nsCheckedCount === streams.length
		})

		status.global = totalStreams > 0 && totalCheckedStreams === totalStreams

		return status
	}, [groupedStreams, bulkSelectedStreams])

	useEffect(() => {
		if (
			sortedGroups.length > 0 &&
			!hasGroupedStreamsStructureChanged(
				prevGroupedStreams.current,
				groupedStreams,
			)
		) {
			return
		}

		setSortedGroups(
			sortGroupedStreamsByCheckedState(groupedStreams, checkedStatus.streams),
		)
		prevGroupedStreams.current = groupedStreams
	}, [groupedStreams, checkedStatus.streams, sortedGroups.length])

	const handleToggleNamespace = (ns: string) => {
		setOpenNamespaces(prev => ({ ...prev, [ns]: !prev[ns] }))
	}

	const handleNamespaceSyncAll = (ns: string, checked: boolean) => {
		let updatedSelection = [...bulkSelectedStreams]
		const streamsInNamespace = groupedStreams[ns] || []

		if (checked) {
			streamsInNamespace.forEach(s => {
				const exists = updatedSelection.some(
					sel => sel.namespace === ns && sel.streamName === s.stream.name,
				)
				if (!exists) {
					updatedSelection.push({ namespace: ns, streamName: s.stream.name })
				}
			})
		} else {
			updatedSelection = updatedSelection.filter(sel => sel.namespace !== ns)
		}

		onChange(updatedSelection)
	}

	const handleGlobalSyncAll = (checked: boolean) => {
		if (checked) {
			const updatedSelection: StreamIdentifier[] = []
			Object.entries(groupedStreams).forEach(([ns, streams]) => {
				streams.forEach(s => {
					updatedSelection.push({ namespace: ns, streamName: s.stream.name })
				})
			})
			onChange(updatedSelection)
		} else {
			onChange([])
		}
	}

	const handleStreamSelect = (
		streamName: string,
		ns: string,
		checked: boolean,
	) => {
		let updatedSelection = [...bulkSelectedStreams]
		if (checked) {
			updatedSelection.push({ namespace: ns, streamName })
		} else {
			updatedSelection = updatedSelection.filter(
				sel => !(sel.namespace === ns && sel.streamName === streamName),
			)
		}
		onChange(updatedSelection)
	}

	return (
		<div className="flex h-full flex-col overflow-y-auto rounded">
			{sortedGroups.length === 0 ? (
				<Empty className="pt-10" />
			) : (
				<>
					<div className="sticky top-0 z-10 flex h-14 items-center justify-between border-b border-neutral-border bg-olake-surface px-8 py-4">
						<Checkbox
							checked={checkedStatus.global}
							onChange={e => handleGlobalSyncAll(e.target.checked)}
						>
							Sync all
						</Checkbox>
					</div>

					{sortedGroups.map(([ns, streams]) => (
						<div
							key={ns}
							className="border-b border-neutral-border"
						>
							<div
								className="sticky top-14 z-[9] flex h-14 cursor-pointer items-center bg-background-primary px-8"
								onClick={() => handleToggleNamespace(ns)}
							>
								<Checkbox
									className="mr-2"
									checked={checkedStatus.namespaces[ns]}
									onChange={e => handleNamespaceSyncAll(ns, e.target.checked)}
									onClick={e => e.stopPropagation()}
								/>
								<span className="text-sm leading-5 text-olake-text">{ns}</span>
								<span className="ml-auto text-olake-text-secondary">
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
								<div className="flex w-full flex-col bg-olake-surface pl-2">
									{streams.map(streamData => {
										const isSelected =
											checkedStatus.streams[ns]?.[streamData.stream.name] ||
											false
										return (
											<div
												key={`${ns}__${streamData.stream.name}`}
												className="flex h-14 items-center border-b border-neutral-border bg-olake-surface px-8 py-4"
											>
												<Checkbox
													className="text-olake-text"
													checked={isSelected}
													onChange={e =>
														handleStreamSelect(
															streamData.stream.name,
															ns,
															e.target.checked,
														)
													}
												>
													<span className="inline-block max-w-[680px] overflow-hidden text-ellipsis whitespace-nowrap text-sm leading-5 text-olake-text">
														{streamData.stream.name}
													</span>
												</Checkbox>
											</div>
										)
									})}
								</div>
							)}
						</div>
					))}
				</>
			)}
		</div>
	)
}

export default BulkStreamSelectorList
