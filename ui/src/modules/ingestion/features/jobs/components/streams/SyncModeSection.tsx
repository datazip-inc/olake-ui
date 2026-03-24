import { ArrowSquareOutIcon, InfoIcon, PlusIcon } from "@phosphor-icons/react"
import { Radio, Select, Tooltip, Button } from "antd"
import { useEffect, useState } from "react"

import { SyncMode } from "@/modules/ingestion/common/types"

import { SYNC_MODE_MAP } from "../../constants"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	useStreamSelectionStore,
} from "../../stores"
import { getCursorFieldValues } from "../../utils/streams"

const SyncModeSection = () => {
	const updateSyncMode = useStreamSelectionStore(state => state.updateSyncMode)
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)

	const [syncMode, setSyncMode] = useState(stream?.stream.sync_mode)
	const [cursorField, setCursorField] = useState<string | undefined>(
		stream?.stream.cursor_field,
	)
	const [showFallbackSelector, setShowFallbackSelector] = useState(false)
	const [fallBackCursorField, setFallBackCursorField] = useState<string>("")

	// Re-sync sync/cursor state when the active stream changes
	useEffect(() => {
		if (!stream || !selectedStream) return

		const initialApiSyncMode = stream.stream.sync_mode
		const initialCursorField = stream.stream.cursor_field

		// Parse cursor field for default value
		// cursor field and default will be in a:b form where a is the cursor field and b is the default field
		const { fallback: initialFallbackCursor } =
			getCursorFieldValues(initialCursorField)

		setFallBackCursorField(initialFallbackCursor)
		setShowFallbackSelector(!!initialFallbackCursor)

		setSyncMode(initialApiSyncMode ?? "full_refresh")
		setCursorField(initialCursorField)
		// Auto-select first available cursor field if default sync mode is incremental and no cursor field is set
		if (initialApiSyncMode === "incremental" && !initialCursorField) {
			const availableCursorFields = stream.stream.available_cursor_fields || []
			const cursor = availableCursorFields[0]
			if (cursor) {
				setCursorField(getCursorFieldValues(cursor).primary)
				setFallBackCursorField(getCursorFieldValues(cursor).fallback)
				updateSyncMode(
					stream.stream.name,
					stream.stream.namespace || "",
					SyncMode.INCREMENTAL,
					cursor,
				)
			}
		}
	}, [stream?.stream.name, stream?.stream.namespace])

	if (!stream || !selectedStream) return null

	const isSyncModeSupported = (mode: string): boolean =>
		stream.stream.supported_sync_modes?.some(m => m === mode) ?? false

	const handleSyncModeChange = (selectedRadioValue: string) => {
		setSyncMode(selectedRadioValue)

		const newApiSyncMode = (
			Object.entries(SYNC_MODE_MAP).find(
				([, value]) => value === selectedRadioValue,
			)?.[0] || ""
		).toLowerCase() as SyncMode

		// Auto-select first available cursor field for incremental mode
		if (selectedRadioValue === "incremental") {
			const availableCursorFields = stream.stream.available_cursor_fields || []
			const cursor = cursorField || availableCursorFields[0]
			if (cursor) {
				setCursorField(getCursorFieldValues(cursor).primary)
				setFallBackCursorField(getCursorFieldValues(cursor).fallback)
				updateSyncMode(
					stream.stream.name,
					stream.stream.namespace || "",
					SyncMode.INCREMENTAL,
					cursor,
				)
			}
		} else {
			updateSyncMode(
				stream.stream.name,
				stream.stream.namespace || "",
				newApiSyncMode,
			)
		}
	}

	const handleCursorChange = (value: string) => {
		const newCursorField = fallBackCursorField
			? `${value}:${fallBackCursorField}`
			: value
		setCursorField(newCursorField)
		setFallBackCursorField("")
		updateSyncMode(
			stream.stream.name,
			stream.stream.namespace || "",
			SyncMode.INCREMENTAL,
			newCursorField,
		)
	}

	const handleFallbackCursorChange = (value: string) => {
		const cursorFieldValues = getCursorFieldValues(cursorField)
		const newCursorField = value
			? `${cursorFieldValues.primary}:${value}`
			: cursorFieldValues.primary
		setCursorField(newCursorField)
		setFallBackCursorField(value)
		updateSyncMode(
			stream.stream.name,
			stream.stream.namespace || "",
			SyncMode.INCREMENTAL,
			newCursorField,
		)
	}

	const handleFallbackCursorClear = () => {
		setShowFallbackSelector(false)
		setFallBackCursorField("")
		const cursorFieldValues = getCursorFieldValues(cursorField)
		const newCursorField = cursorFieldValues.primary
		setCursorField(newCursorField)
		updateSyncMode(
			stream.stream.name,
			stream.stream.namespace || "",
			SyncMode.INCREMENTAL,
			newCursorField,
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
						disabled={!isSyncModeSupported(SyncMode.INCREMENTAL)}
					>
						Full Refresh + Incremental
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
