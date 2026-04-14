import {
	ArrowSquareOutIcon,
	InfoIcon,
	PlusIcon,
	WarningIcon,
} from "@phosphor-icons/react"
import { Radio, Select, Tooltip, Button } from "antd"
import { useEffect, useState } from "react"

import { SyncMode, StreamData } from "@/modules/ingestion/common/types"

import { SYNC_MODE_MAP } from "../../constants"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	useStreamSelectionStore,
	noopNullSelector,
} from "../../stores"
import { getCursorFieldValues } from "../../utils/streams"

export interface SyncModeSectionProps {
	isBulkMode?: boolean
	isDirty?: boolean
	bulkStream?: StreamData
	bulkSyncMode?: string
	bulkCursorField?: string
	onBulkSyncModeChange?: (syncMode: string, cursorField?: string) => void
}

const SyncModeSection = ({
	isBulkMode,
	isDirty,
	bulkStream,
	bulkSyncMode,
	bulkCursorField,
	onBulkSyncModeChange,
}: SyncModeSectionProps = {}) => {
	const updateSyncMode = useStreamSelectionStore(state => state.updateSyncMode)
	// don't subsribe to store if in bulkMode
	const storeStream = useStreamSelectionStore(
		isBulkMode ? noopNullSelector : selectActiveStreamData,
	)
	const storeSelectedStream = useStreamSelectionStore(
		isBulkMode ? noopNullSelector : selectActiveSelectedStream,
	)

	const stream = isBulkMode ? bulkStream : storeStream
	const selectedStream = isBulkMode ? null : storeSelectedStream

	const [storeSyncMode, setStoreSyncMode] = useState(
		storeStream?.stream.sync_mode,
	)
	const [storeCursorField, setStoreCursorField] = useState<string | undefined>(
		storeStream?.stream.cursor_field,
	)

	const syncMode = isBulkMode ? bulkSyncMode : storeSyncMode
	const cursorField = isBulkMode ? bulkCursorField : storeCursorField

	const [showFallbackSelector, setShowFallbackSelector] = useState(false)
	const [fallBackCursorField, setFallBackCursorField] = useState<string>("")

	// Re-sync sync/cursor state when the active stream changes
	useEffect(() => {
		// Parse cursor field for default value
		// cursor field and default will be in a:b form where a is the cursor field and b is the default field
		const activeCursorField = isBulkMode
			? bulkCursorField
			: storeStream?.stream.cursor_field
		const { fallback: initialFallbackCursor } =
			getCursorFieldValues(activeCursorField)
		setFallBackCursorField(initialFallbackCursor)
		setShowFallbackSelector(!!initialFallbackCursor)

		if (isBulkMode) return

		if (!storeStream || !storeSelectedStream) return

		const initialApiSyncMode = storeStream.stream.sync_mode

		setStoreSyncMode(initialApiSyncMode ?? "full_refresh")
		setStoreCursorField(activeCursorField)
		// Auto-select first available cursor field if default sync mode is incremental and no cursor field is set
		if (initialApiSyncMode === "incremental" && !activeCursorField) {
			const cursor = getColumnOptionsForCursor()[0]?.value
			if (cursor) {
				setStoreCursorField(getCursorFieldValues(cursor).primary)
				setFallBackCursorField(getCursorFieldValues(cursor).fallback)
				updateSyncMode(
					storeStream.stream.name,
					storeStream.stream.namespace || "",
					SyncMode.INCREMENTAL,
					cursor,
				)
			}
		}
	}, [
		isBulkMode,
		bulkCursorField,
		storeStream?.stream.name,
		storeStream?.stream.namespace,
	])

	if (!stream || (!isBulkMode && !selectedStream)) return null

	const isSyncModeSupported = (mode: string): boolean =>
		stream.stream.supported_sync_modes?.some(m => m === mode) ?? false

	const dispatchUpdate = (mode: SyncMode, cf?: string) => {
		if (isBulkMode) {
			onBulkSyncModeChange?.(mode, cf)
		} else {
			setStoreSyncMode(mode)
			setStoreCursorField(cf)
			setFallBackCursorField(getCursorFieldValues(cf).fallback)
			if (storeStream) {
				updateSyncMode(
					storeStream.stream.name,
					storeStream.stream.namespace || "",
					mode,
					cf,
				)
			}
		}
	}

	const handleSyncModeChange = (selectedRadioValue: string) => {
		const newApiSyncMode = (
			Object.entries(SYNC_MODE_MAP).find(
				([, value]) => value === selectedRadioValue,
			)?.[0] || ""
		).toLowerCase() as SyncMode

		// Auto-select first available cursor field for incremental mode
		if (selectedRadioValue === "incremental") {
			const cursor = cursorField || getColumnOptionsForCursor()[0]?.value
			dispatchUpdate(SyncMode.INCREMENTAL, cursor)
		} else {
			dispatchUpdate(newApiSyncMode, undefined)
		}
	}

	const handleCursorChange = (value: string) => {
		const newCursorField = fallBackCursorField
			? `${value}:${fallBackCursorField}`
			: value
		dispatchUpdate(SyncMode.INCREMENTAL, newCursorField)
	}

	const handleFallbackCursorChange = (value: string) => {
		const { primary } = getCursorFieldValues(cursorField)
		const newCursorField = value ? `${primary}:${value}` : primary
		dispatchUpdate(SyncMode.INCREMENTAL, newCursorField)
	}

	const handleFallbackCursorClear = () => {
		setShowFallbackSelector(false)
		dispatchUpdate(
			SyncMode.INCREMENTAL,
			getCursorFieldValues(cursorField).primary,
		)
	}

	const getColumnOptionsForCursor = (
		isFallback: boolean = false,
	): { label: React.ReactNode; value: string }[] => {
		const availableCursorFields = stream.stream.available_cursor_fields || []
		const selectedField = getCursorFieldValues(cursorField).primary

		return [...availableCursorFields]
			.filter(field => !isFallback || field !== selectedField)
			.sort((a, b) => {
				const aIsPK =
					stream.stream.source_defined_primary_key?.includes(a) || false
				const bIsPK =
					stream.stream.source_defined_primary_key?.includes(b) || false
				if (aIsPK && !bIsPK) return -1
				if (!aIsPK && bIsPK) return 1
				return a.localeCompare(b)
			})
			.map((field: string) => ({
				label: (
					<div className="flex items-center justify-between">
						<span>{field}</span>
						{stream.stream.source_defined_primary_key?.includes(field) && (
							<span className="text-primary">PK</span>
						)}
					</div>
				),
				value: field,
			}))
	}

	const cursorFieldValues = getCursorFieldValues(cursorField)

	return (
		<>
			<div className="mb-4">
				<div className="mb-3 flex w-full items-center gap-1 font-medium text-neutral-text">
					{isDirty && <WarningIcon className="size-4 text-orange-500" />}
					<label>Sync mode:</label>
					<a
						href="https://olake.io/docs/understanding/terminologies/olake/#2-sync-modes"
						target="_blank"
						rel="noopener noreferrer"
						aria-label="Open sync modes docs"
						className="inline-flex text-text-tertiary hover:text-primary"
					>
						<ArrowSquareOutIcon size={14} />
					</a>
				</div>
				<Radio.Group
					className="mb-4 grid grid-cols-2 gap-4"
					value={syncMode}
					onChange={e => handleSyncModeChange(e.target.value)}
				>
					<Radio
						value="full_refresh"
						disabled={!isSyncModeSupported(SyncMode.FULL_REFRESH)}
					>
						Full Refresh
					</Radio>
					<Radio
						value="incremental"
						disabled={
							!isSyncModeSupported(SyncMode.INCREMENTAL) ||
							(isBulkMode &&
								(stream.stream.available_cursor_fields || []).length === 0)
						}
					>
						<Tooltip
							title={
								isBulkMode &&
								(stream.stream.available_cursor_fields || []).length === 0
									? "No common cursor fields across selected streams"
									: ""
							}
						>
							Full Refresh + Incremental
						</Tooltip>
					</Radio>
					<Radio
						value="cdc"
						disabled={!isSyncModeSupported(SyncMode.CDC)}
					>
						Full Refresh + CDC
					</Radio>
					<Radio
						value="strict_cdc"
						disabled={!isSyncModeSupported(SyncMode.STRICT_CDC)}
					>
						CDC Only
					</Radio>
				</Radio.Group>
				{syncMode === "incremental" &&
					stream.stream.available_cursor_fields && (
						<div className="mb-4 mr-2">
							<div className="flex w-full gap-4">
								<div className="flex w-1/2 flex-col">
									<label className="mb-1 flex items-center gap-1 font-medium text-neutral-text">
										Cursor field:
										<Tooltip title="Column for identifying new/updated records ">
											<InfoIcon className="size-3.5 cursor-pointer" />
										</Tooltip>
									</label>
									<Select
										placeholder="Select cursor field"
										value={cursorFieldValues.primary || undefined}
										onChange={handleCursorChange}
										optionLabelProp="label"
									>
										{getColumnOptionsForCursor().map(option => (
											<Select.Option
												key={option.value}
												value={option.value}
												label={option.value}
											>
												{option.label}
											</Select.Option>
										))}
									</Select>
								</div>
								{cursorField &&
									!showFallbackSelector &&
									!fallBackCursorField && (
										<div className="flex w-1/2 items-end">
											<Tooltip title="Alternative cursor column in case cursor column encounters null values">
												<Button
													type="default"
													icon={<PlusIcon className="size-4" />}
													onClick={() => setShowFallbackSelector(true)}
													className="mb-[2px] flex items-center gap-1"
												>
													Add Fallback Cursor
												</Button>
											</Tooltip>
										</div>
									)}
								{cursorField &&
									(showFallbackSelector || fallBackCursorField) && (
										<div className="flex w-1/2 flex-col">
											<label className="mb-1 flex items-center gap-1 font-medium text-neutral-text">
												Fallback Cursor:
												<Tooltip title="Alternative cursor column in case cursor column encounters null values">
													<InfoIcon className="size-3.5 cursor-pointer text-neutral-text" />
												</Tooltip>
											</label>
											<Select
												placeholder="Select default"
												value={fallBackCursorField}
												onChange={handleFallbackCursorChange}
												allowClear
												onClear={handleFallbackCursorClear}
												optionLabelProp="label"
											>
												{getColumnOptionsForCursor(true).map(option => (
													<Select.Option
														key={option.value}
														value={option.value}
														label={option.value}
													>
														{option.label}
													</Select.Option>
												))}
											</Select>
										</div>
									)}
							</div>
						</div>
					)}
			</div>
		</>
	)
}

export default SyncModeSection
