import {
	CheckIcon,
	FadersHorizontalIcon,
	RowsIcon,
	TableIcon,
	XIcon,
} from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import clsx from "clsx"
import { useEffect, useMemo, useState } from "react"

import {
	FilterConfig,
	StreamIdentifier,
	SyncMode,
} from "@/modules/ingestion/common/types"
import { BulkConfigureStreamsModalProps } from "@/modules/ingestion/features/jobs/types"

import { useStreamSelectionStore } from "../../stores"
import { buildBulkStreamsData } from "../../utils/streams"
import BulkStreamSelectorList from "../streams/BulkStreamSelectorList"
import DataFilterSection from "../streams/DataFilterSection"
import IngestionModeSection from "../streams/IngestionModeSection"
import NormalizationSection from "../streams/NormalizationSection"
import PartitionRegexSection from "../streams/PartitionRegexSection"
import SyncModeSection from "../streams/SyncModeSection"

type BulkConfigureStep = "select-streams" | "apply-configurations" | "success"
type BulkConfigurationTab = "config" | "partitioning"

type BulkConfig = {
	syncMode: string
	cursorField: string | undefined
	appendMode: boolean
	normalization: boolean
	filter: string
	filterConfig: FilterConfig | undefined
	partitionRegex: string
}

enum BulkDirtyFieldKey {
	SyncMode = "syncMode",
	AppendMode = "appendMode",
	Normalization = "normalization",
	Filter = "filter",
	PartitionRegex = "partitionRegex",
}

type BulkDirtyFields = Record<BulkDirtyFieldKey, boolean>

const INITIAL_DIRTY_FIELDS: BulkDirtyFields = {
	[BulkDirtyFieldKey.SyncMode]: false,
	[BulkDirtyFieldKey.AppendMode]: false,
	[BulkDirtyFieldKey.Normalization]: false,
	[BulkDirtyFieldKey.Filter]: false,
	[BulkDirtyFieldKey.PartitionRegex]: false,
}

const INITIAL_BULK_CONFIG: BulkConfig = {
	syncMode: SyncMode.FULL_REFRESH,
	cursorField: undefined,
	appendMode: false,
	normalization: false,
	filter: "",
	filterConfig: undefined,
	partitionRegex: "",
}

const CLOSE_COUNTDOWN = 3

const sortCursorFields = (
	availableCursors: string[],
	primaryKeys: string[],
): string[] =>
	[...availableCursors].sort((a, b) => {
		const aIsPK = primaryKeys.includes(a)
		const bIsPK = primaryKeys.includes(b)
		if (aIsPK && !bIsPK) return -1
		if (!aIsPK && bIsPK) return 1
		return a.localeCompare(b)
	})

const BulkConfigureStreamsModal = ({
	open,
	onClose,
	streamsData,
	sourceType,
	destinationType,
}: BulkConfigureStreamsModalProps) => {
	const [step, setStep] = useState<BulkConfigureStep>("select-streams")
	const [activeTab, setActiveTab] = useState<BulkConfigurationTab>("config")
	const [closeCountdown, setCloseCountdown] = useState(CLOSE_COUNTDOWN)

	const [bulkSelectedStreams, setBulkSelectedStreams] = useState<
		StreamIdentifier[]
	>([])

	// Local bulk config state
	const [bulkConfig, setBulkConfig] = useState<BulkConfig>(INITIAL_BULK_CONFIG)
	const setBulkConfigField = <K extends keyof BulkConfig>(
		key: K,
		value: BulkConfig[K],
	) => setBulkConfig(prev => ({ ...prev, [key]: value }))

	// Tracks which sections the user has explicitly modified.
	// Only dirty sections are included in the apply payload.
	const [dirtyFields, setDirtyFields] =
		useState<BulkDirtyFields>(INITIAL_DIRTY_FIELDS)

	const markDirty = (key: BulkDirtyFieldKey) =>
		setDirtyFields(prev => ({ ...prev, [key]: true }))

	const selectionKey = useMemo(
		() =>
			[...bulkSelectedStreams]
				.map(s => `${s.namespace}__${s.streamName}`)
				.sort()
				.join(","),
		[bulkSelectedStreams],
	)

	const { stream: bulkStream, defaults: bulkStreamDefaults } = useMemo(
		() =>
			buildBulkStreamsData(
				bulkSelectedStreams,
				streamsData,
				sourceType,
				destinationType,
			),
		[selectionKey, streamsData, sourceType, destinationType],
	)

	useEffect(() => {
		if (open) {
			setStep("select-streams")
			setActiveTab("config")
			setCloseCountdown(CLOSE_COUNTDOWN)
			setBulkSelectedStreams([])
		}
	}, [open])

	useEffect(() => {
		if (!open || step !== "success") return

		if (closeCountdown <= 0) {
			onClose()
			return
		}

		const timeoutId = window.setTimeout(() => {
			setCloseCountdown(prev => prev - 1)
		}, 1000)

		return () => window.clearTimeout(timeoutId)
	}, [open, step, closeCountdown, onClose])

	useEffect(() => {
		// Reset all config state to defaults on selection change.
		const syncMode = bulkStream.sync_mode
		const availableCursors = bulkStream.stream.available_cursor_fields ?? []
		const primaryKeys = bulkStream.stream.source_defined_primary_key ?? []
		const sortedCursors = sortCursorFields(availableCursors, primaryKeys)
		setBulkConfig({
			syncMode,
			cursorField:
				syncMode === SyncMode.INCREMENTAL ? sortedCursors[0] : undefined,
			appendMode: bulkStreamDefaults.append_mode ?? false,
			normalization: bulkStreamDefaults.normalization,
			filter: "",
			filterConfig: undefined,
			partitionRegex: "",
		})
		setDirtyFields(INITIAL_DIRTY_FIELDS)
	}, [selectionKey])

	const getStepTitle = () => {
		if (step === "select-streams") return "Select Streams"
		if (step === "apply-configurations") return "Apply Configurations"
		return ""
	}

	const bulkUpdateStreams = useStreamSelectionStore(
		state => state.bulkUpdateStreams,
	)

	const handleApplyChanges = () => {
		bulkUpdateStreams(bulkSelectedStreams, {
			...(dirtyFields[BulkDirtyFieldKey.SyncMode] && {
				syncMode: bulkConfig.syncMode as SyncMode,
				cursorField: bulkConfig.cursorField,
			}),
			...(dirtyFields[BulkDirtyFieldKey.AppendMode] && {
				appendMode: bulkConfig.appendMode,
			}),
			...(dirtyFields[BulkDirtyFieldKey.Normalization] && {
				normalization: bulkConfig.normalization,
			}),
			...(dirtyFields[BulkDirtyFieldKey.Filter] && {
				filterValue: bulkConfig.filter,
				filterConfig: bulkConfig.filterConfig,
			}),
			...(dirtyFields[BulkDirtyFieldKey.PartitionRegex] && {
				partitionRegex: bulkConfig.partitionRegex,
			}),
		})

		setCloseCountdown(CLOSE_COUNTDOWN)
		setStep("success")
	}

	const handleRemoveSelectedStream = (streamToRemove: StreamIdentifier) => {
		setBulkSelectedStreams(prev =>
			prev.filter(
				stream =>
					!(
						stream.namespace === streamToRemove.namespace &&
						stream.streamName === streamToRemove.streamName
					),
			),
		)
	}

	const getFooter = () => {
		if (step === "success") return null

		if (step === "select-streams") {
			return (
				<div className="flex h-20 items-center justify-end gap-3 border-t border-olake-border px-8">
					<Button onClick={onClose}>Cancel</Button>
					<Button
						type="primary"
						onClick={() => setStep("apply-configurations")}
						disabled={bulkSelectedStreams.length === 0}
					>
						Configure Streams
					</Button>
				</div>
			)
		}

		return (
			<div className="flex h-20 items-center justify-between border-t border-olake-border px-8">
				<Button onClick={() => setStep("select-streams")}>Back</Button>
				<div className="flex items-center gap-3">
					<Button onClick={onClose}>Cancel</Button>
					<Button
						type="primary"
						onClick={handleApplyChanges}
						disabled={
							bulkSelectedStreams.length === 0 ||
							!Object.values(dirtyFields).some(Boolean)
						}
					>
						Apply Changes
					</Button>
				</div>
			</div>
		)
	}

	return (
		<Modal
			open={open}
			onCancel={onClose}
			destroyOnHidden
			footer={getFooter()}
			closable={false}
			centered
			width={983}
			classNames={{
				content: "!overflow-hidden !rounded-[20px] !p-0",
				body: "!p-0",
				footer: "!m-0 !p-0",
			}}
		>
			{step === "success" ? (
				<div className="relative h-[808px] bg-background-primary">
					<div className="absolute left-1/2 top-1/2 flex w-[374px] -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-5 text-center">
						<div className="rounded-xl bg-primary-100 p-3">
							<CheckIcon
								weight="bold"
								className="size-8 text-primary"
							/>
						</div>
						<div className="w-full">
							<div className="text-xl font-medium leading-7 text-olake-text">
								{`${bulkSelectedStreams.length} streams configured successfully`}
							</div>
							<div className="mt-1 text-base leading-6 text-olake-text">
								You are free to edit the stream separately if you wish
							</div>
						</div>
					</div>
					<div className="absolute left-1/2 top-[683px] -translate-x-1/2">
						<Button onClick={onClose}>Closing in {closeCountdown}...</Button>
					</div>
				</div>
			) : (
				<div className="flex h-[728px] flex-col">
					<div className="border-b border-olake-border px-8 pb-5 pt-8">
						<h2 className="text-xl font-medium leading-7 text-olake-text">
							Bulk Streams configure
						</h2>
						<p className="mt-2 text-sm leading-5 text-olake-text">
							Select streams you wish to bulk configure
						</p>
					</div>

					<div className="border-b border-olake-border px-8 py-4 text-base font-medium leading-7 text-olake-text">
						{getStepTitle()}
					</div>

					<div className="min-h-0 flex-1">
						{step === "select-streams" ? (
							<div className="h-full">
								<BulkStreamSelectorList
									streamsData={streamsData}
									bulkSelectedStreams={bulkSelectedStreams}
									onChange={setBulkSelectedStreams}
								/>
							</div>
						) : (
							<div className={clsx("h-full", "overflow-y-auto px-8 py-5")}>
								<div className="text-base leading-6 text-olake-text">
									Streams Selected ({bulkSelectedStreams.length})
								</div>
								{bulkSelectedStreams.length === 0 ? (
									<div className="flex h-full items-center justify-center">
										<div className="max-w-[420px] text-center">
											<div className="text-base font-medium text-olake-text">
												No Streams Selected
											</div>
											<div className="mt-2 text-sm leading-5 text-olake-text-secondary">
												No streams are selected for bulk configuration. Please
												Go back and select one or more streams to continue.
											</div>
										</div>
									</div>
								) : (
									<>
										<div className="mt-2 flex flex-wrap gap-2">
											{bulkSelectedStreams.map(stream => {
												const streamId = `${stream.namespace}__${stream.streamName}`
												return (
													<div
														key={streamId}
														className="flex h-7 items-center gap-2 rounded bg-olake-surface-muted px-3.5 py-0.5"
													>
														<TableIcon className="size-4 text-olake-text" />
														<span className="text-sm text-olake-text">
															{stream.namespace ? `${stream.namespace} / ` : ""}
															{stream.streamName}
														</span>
														<button
															type="button"
															onClick={() => handleRemoveSelectedStream(stream)}
															aria-label={`Remove ${stream.streamName}`}
															className="inline-flex items-center text-olake-text-tertiary hover:text-olake-text"
														>
															<XIcon className="size-4" />
														</button>
													</div>
												)
											})}
										</div>
										<div className="mt-4 flex items-center gap-2 rounded-md border border-olake-border bg-olake-surface-muted px-3 py-2 text-sm text-olake-text-secondary">
											<span className="mr-1 inline-block h-2 w-2 shrink-0 rounded-full bg-warning" />
											<span>
												Only fields marked with this indicator will be applied
												to all selected streams and will override previous
												configuration.
											</span>
										</div>

										<div className="mt-8 rounded-md bg-olake-surface-muted p-0.5">
											<div className="grid grid-cols-2 gap-1">
												<button
													type="button"
													onClick={() => setActiveTab("config")}
													className={clsx(
														"flex h-7 items-center justify-center gap-2 rounded-md border px-3 text-sm leading-5",
														activeTab === "config"
															? "border-primary bg-white text-primary shadow-sm"
															: "border-transparent text-olake-text",
													)}
												>
													<FadersHorizontalIcon className="size-4" />
													Config
												</button>
												<button
													type="button"
													onClick={() => setActiveTab("partitioning")}
													className={clsx(
														"flex h-7 items-center justify-center gap-2 rounded-md border px-3 text-sm leading-5",
														activeTab === "partitioning"
															? "border-primary bg-white text-primary shadow-sm"
															: "border-transparent text-olake-text",
													)}
												>
													<RowsIcon className="size-4" />
													Partitioning
												</button>
											</div>
										</div>
										<div className="mt-4">
											{activeTab === "config" ? (
												<div
													key={selectionKey}
													className="flex flex-col gap-4"
												>
													{/* Sync Mode and Ingestion Mode Sections */}
													<div className="rounded border border-neutral-disabled bg-white p-4">
														<SyncModeSection
															isBulkMode
															isDirty={dirtyFields[BulkDirtyFieldKey.SyncMode]}
															bulkStream={bulkStream}
															bulkSyncMode={bulkConfig.syncMode}
															bulkCursorField={bulkConfig.cursorField}
															onBulkSyncModeChange={(mode, cursor) => {
																setBulkConfigField("syncMode", mode)
																setBulkConfigField("cursorField", cursor)
																markDirty(BulkDirtyFieldKey.SyncMode)
															}}
														/>
														<IngestionModeSection
															isBulkMode
															isDirty={
																dirtyFields[BulkDirtyFieldKey.AppendMode]
															}
															sourceType={sourceType}
															destinationType={destinationType}
															bulkAppendMode={bulkConfig.appendMode}
															onBulkIngestionModeChange={value => {
																setBulkConfigField("appendMode", value)
																markDirty(BulkDirtyFieldKey.AppendMode)
															}}
														/>
													</div>

													{/* Normalization Section */}
													<NormalizationSection
														isBulkMode
														isDirty={
															dirtyFields[BulkDirtyFieldKey.Normalization]
														}
														bulkNormalization={bulkConfig.normalization}
														onBulkNormalizationChange={(value: boolean) => {
															setBulkConfigField("normalization", value)
															markDirty(BulkDirtyFieldKey.Normalization)
														}}
													/>

													{/* Data Filter Section */}
													<DataFilterSection
														isBulkMode
														isDirty={dirtyFields[BulkDirtyFieldKey.Filter]}
														bulkStream={bulkStream}
														bulkFilter={bulkConfig.filter}
														onBulkFilterChange={value => {
															setBulkConfigField("filter", value)
															markDirty(BulkDirtyFieldKey.Filter)
														}}
														bulkFilterConfig={bulkConfig.filterConfig}
														onBulkFilterConfigChange={value => {
															setBulkConfigField("filterConfig", value)
															markDirty(BulkDirtyFieldKey.Filter)
														}}
													/>
												</div>
											) : (
												<div
													key={selectionKey}
													className="flex flex-col gap-4"
												>
													<PartitionRegexSection
														isBulkMode
														isDirty={
															dirtyFields[BulkDirtyFieldKey.PartitionRegex]
														}
														destinationType={destinationType}
														bulkPartitionRegex={bulkConfig.partitionRegex}
														onBulkPartitionRegexChange={value => {
															setBulkConfigField("partitionRegex", value)
															markDirty(BulkDirtyFieldKey.PartitionRegex)
														}}
													/>
												</div>
											)}
										</div>
									</>
								)}
							</div>
						)}
					</div>
				</div>
			)}
		</Modal>
	)
}

export default BulkConfigureStreamsModal
