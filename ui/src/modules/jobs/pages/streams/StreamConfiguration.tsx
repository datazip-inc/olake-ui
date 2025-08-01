import { useEffect, useState } from "react"
import {
	ExtendedStreamConfigurationProps,
	FilterCondition,
	FilterOperator,
	LogicalOperator,
	MultiFilterCondition,
	CombinedStreamsData,
} from "../../../../types"
import { Button, Divider, Input, Radio, Select, Switch, Tooltip } from "antd"
import StreamsSchema from "./StreamsSchema"
import {
	ColumnsPlusRight,
	GridFour,
	Info,
	Lightning,
	Plus,
	SlidersHorizontal,
	X,
	ArrowSquareOut,
} from "@phosphor-icons/react"
import {
	CARD_STYLE,
	PartioningRegexTooltip,
	TAB_STYLES,
} from "../../../../utils/constants"
import { operatorOptions } from "../../../../utils/utils"

const StreamConfiguration = ({
	stream,
	onSyncModeChange,
	isSelected,
	initialNormalization,
	initialPartitionRegex,
	onNormalizationChange,
	onPartitionRegexChange,
	initialFullLoadFilter = "",
	onFullLoadFilterChange,
	fromJobEditFlow = false,
	initialSelectedStreams,
	destinationType = "s3",
}: ExtendedStreamConfigurationProps) => {
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState(
		stream.stream.sync_mode === "full_refresh"
			? "full"
			: stream.stream.sync_mode === "incremental"
				? "incremental"
				: "cdc",
	)
	const [normalisation, setNormalisation] =
		useState<boolean>(initialNormalization)
	const [fullLoadFilter, setFullLoadFilter] = useState<boolean>(false)
	const [partitionRegex, setPartitionRegex] = useState("")
	const [showFallbackSelector, setShowFallbackSelector] = useState(false)
	const [fallBackCursorField, setFallBackCursorField] = useState<string>("")
	const [activePartitionRegex, setActivePartitionRegex] = useState(
		initialPartitionRegex || "",
	)
	const [multiFilterCondition, setMultiFilterCondition] =
		useState<MultiFilterCondition>({
			conditions: [
				{
					columnName: "",
					operator: "=",
					value: "",
				},
			],
			logicalOperator: "and",
		})
	const [formData, setFormData] = useState<any>({
		sync_mode: stream.stream.sync_mode,
		backfill: false,
		partition_regex: initialPartitionRegex || "",
	})

	const [initialJobStreams, setInitialJobStreams] = useState<
		CombinedStreamsData | undefined
	>(undefined)

	useEffect(() => {
		// Set initial streams only once when component mounts
		if (fromJobEditFlow && initialSelectedStreams && !initialJobStreams) {
			setInitialJobStreams(initialSelectedStreams)
		}
	}, [fromJobEditFlow, initialSelectedStreams])

	// Check if this stream was in the initial job streams
	const isStreamInInitialSelection =
		fromJobEditFlow &&
		initialJobStreams?.selected_streams?.[stream.stream.namespace || ""]?.some(
			(s: { stream_name: string }) => s.stream_name === stream.stream.name,
		)

	useEffect(() => {
		setActiveTab("config")
		const initialApiSyncMode = stream.stream.sync_mode
		let initialEnableBackfillForSwitch = true

		// Parse cursor field for default value
		if (
			stream.stream.cursor_field &&
			stream.stream.cursor_field.includes(":")
		) {
			const [, defaultField] = stream.stream.cursor_field.split(":")
			setFallBackCursorField(defaultField)
			setShowFallbackSelector(true)
		} else {
			setFallBackCursorField("")
			setShowFallbackSelector(false)
		}

		if (initialApiSyncMode === "full_refresh") {
			setSyncMode("full")
		} else if (initialApiSyncMode === "cdc") {
			setSyncMode("cdc")
		} else if (initialApiSyncMode === "strict_cdc") {
			setSyncMode("cdc")
			initialEnableBackfillForSwitch = false
		} else if (initialApiSyncMode === "incremental") {
			setSyncMode("incremental")
		}
		setNormalisation(initialNormalization)
		setActivePartitionRegex(initialPartitionRegex || "")
		setPartitionRegex("")

		// Parse initial filter if exists
		if (initialFullLoadFilter) {
			const conditions: FilterCondition[] = []
			let logicalOperator: LogicalOperator = "and"

			// Check for AND/OR operator
			const parts = initialFullLoadFilter.toLowerCase().includes(" and ")
				? initialFullLoadFilter.split(" and ")
				: initialFullLoadFilter.split(" or ")

			if (parts.length > 1) {
				logicalOperator = initialFullLoadFilter.toLowerCase().includes(" and ")
					? "and"
					: "or"
			}

			parts.forEach(part => {
				const operatorMatch = part.match(/(>=|<=|=|!=|>|<)/)
				if (operatorMatch) {
					const operator = operatorMatch[0] as FilterOperator
					const [columnName, value] = part.split(operator)
					// Remove quotes if present in the value
					const cleanValue = value.trim().replace(/^"(.*)"$/, "$1")
					conditions.push({
						columnName: columnName.trim(),
						operator,
						value: cleanValue,
					})
				}
			})

			if (conditions.length > 0) {
				setMultiFilterCondition({
					conditions,
					logicalOperator,
				})
				setFullLoadFilter(true)
			}
		} else {
			setMultiFilterCondition({
				conditions: [
					{
						columnName: "",
						operator: "=",
						value: "",
					},
				],
				logicalOperator: "and",
			})
			setFullLoadFilter(false)
		}

		setFormData((prevFormData: any) => ({
			...prevFormData,
			sync_mode: initialApiSyncMode,
			backfill: initialEnableBackfillForSwitch,
			partition_regex: initialPartitionRegex || "",
		}))
	}, [stream, initialNormalization, initialFullLoadFilter])

	// Add helper function for checking supported sync modes
	const isSyncModeSupported = (mode: string): boolean => {
		return (
			stream.stream.supported_sync_modes?.some(
				supportedMode => supportedMode === mode,
			) ?? false
		)
	}

	// Handlers
	const handleSyncModeChange = (selectedRadioValue: string) => {
		setSyncMode(selectedRadioValue)
		let newApiSyncMode: "full_refresh" | "cdc" | "incremental" | "strict_cdc"
		let newEnableBackfillState = true
		if (selectedRadioValue === "full") {
			newApiSyncMode = "full_refresh"
		} else if (selectedRadioValue === "incremental") {
			newApiSyncMode = "incremental"
			// Auto-select first available cursor field if none is selected
			const availableCursorFields = stream.stream.available_cursor_fields || []
			if (!stream.stream.cursor_field && availableCursorFields.length > 0) {
				const firstCursorField = availableCursorFields[0]
				stream.stream.cursor_field = firstCursorField
			}
		} else if (selectedRadioValue === "cdc") {
			newApiSyncMode = "cdc"
		} else {
			newApiSyncMode = "strict_cdc"
		}

		stream.stream.sync_mode = newApiSyncMode
		onSyncModeChange?.(
			stream.stream.name,
			stream.stream.namespace || "",
			newApiSyncMode,
		)

		setFormData({
			...formData,
			sync_mode: newApiSyncMode,
			backfill: newEnableBackfillState,
		})
	}

	const handleNormalizationChange = (checked: boolean) => {
		setNormalisation(checked)
		onNormalizationChange(
			stream.stream.name,
			stream.stream.namespace || "",
			checked,
		)
		setFormData({
			...formData,
			normalization: checked,
		})
	}

	const handleSetPartitionRegex = () => {
		if (partitionRegex) {
			setActivePartitionRegex(partitionRegex)
			setPartitionRegex("")
			onPartitionRegexChange(
				stream.stream.name,
				stream.stream.namespace || "",
				partitionRegex,
			)
			setFormData({
				...formData,
				partition_regex: partitionRegex,
			})
		}
	}

	const handleClearPartitionRegex = () => {
		setActivePartitionRegex("")
		setPartitionRegex("")
		onPartitionRegexChange(
			stream.stream.name,
			stream.stream.namespace || "",
			"",
		)
		setFormData({
			...formData,
			partition_regex: "",
		})
	}

	const handleFullLoadFilterChange = (checked: boolean) => {
		setFullLoadFilter(checked)
		if (!checked) {
			setMultiFilterCondition({
				conditions: [
					{
						columnName: "",
						operator: "=",
						value: "",
					},
				],
				logicalOperator: "and",
			})
			onFullLoadFilterChange?.(
				stream.stream.name,
				stream.stream.namespace || "",
				"",
			)
		}
	}

	const handleFilterConditionChange = (
		index: number,
		field: keyof FilterCondition,
		value: string,
	) => {
		const newConditions = [...multiFilterCondition.conditions]
		newConditions[index] = {
			...newConditions[index],
			[field]: value,
		}

		const newMultiCondition = {
			...multiFilterCondition,
			conditions: newConditions,
		}
		setMultiFilterCondition(newMultiCondition)

		// Generate filter string if all fields in any condition are filled
		const filledConditions = newConditions.filter(
			cond => cond.columnName && cond.operator && cond.value,
		)

		if (filledConditions.length > 0) {
			const filterString = filledConditions
				.map(
					cond =>
						`${cond.columnName} ${cond.operator} ${formatFilterValue(cond.columnName, cond.value)}`,
				)
				.join(` ${multiFilterCondition.logicalOperator} `)

			onFullLoadFilterChange?.(
				stream.stream.name,
				stream.stream.namespace || "",
				filterString,
			)
		}
	}

	const handleLogicalOperatorChange = (value: LogicalOperator) => {
		const newMultiCondition = {
			...multiFilterCondition,
			logicalOperator: value,
		}
		setMultiFilterCondition(newMultiCondition)

		// Regenerate filter string if conditions exist
		const filledConditions = multiFilterCondition.conditions.filter(
			cond => cond.columnName && cond.operator && cond.value,
		)

		if (filledConditions.length > 1) {
			const filterString = filledConditions
				.map(
					cond =>
						`${cond.columnName} ${cond.operator} ${formatFilterValue(cond.columnName, cond.value)}`,
				)
				.join(` ${value} `)

			onFullLoadFilterChange?.(
				stream.stream.name,
				stream.stream.namespace || "",
				filterString,
			)
		}
	}

	const handleAddFilter = () => {
		if (multiFilterCondition.conditions.length < 2) {
			setMultiFilterCondition({
				...multiFilterCondition,
				conditions: [
					...multiFilterCondition.conditions,
					{
						columnName: "",
						operator: "=",
						value: "",
					},
				],
			})
		}
	}

	const handleRemoveFilter = (index: number) => {
		const newConditions = multiFilterCondition.conditions.filter(
			(_, i) => i !== index,
		)
		const newMultiCondition = {
			...multiFilterCondition,
			conditions: newConditions,
		}
		setMultiFilterCondition(newMultiCondition)

		// If removing leaves us with one condition, update the filter string
		if (newConditions.length === 1) {
			const condition = newConditions[0]
			if (condition.columnName && condition.operator && condition.value) {
				const filterString = `${condition.columnName} ${condition.operator} ${formatFilterValue(condition.columnName, condition.value)}`
				onFullLoadFilterChange?.(
					stream.stream.name,
					stream.stream.namespace || "",
					filterString,
				)
			} else {
				onFullLoadFilterChange?.(
					stream.stream.name,
					stream.stream.namespace || "",
					"",
				)
			}
		}
	}

	const getColumnOptions = () => {
		const properties = stream.stream.type_schema?.properties || {}
		const primaryKeys = (stream.stream.source_defined_primary_key ||
			[]) as string[]
		const cursorFields = (stream.stream.available_cursor_fields ||
			[]) as string[]

		return cursorFields
			.filter(key => properties[key.toLowerCase()])
			.sort((a, b) => {
				const aIsPK = primaryKeys.includes(a)
				const bIsPK = primaryKeys.includes(b)
				if (aIsPK && !bIsPK) return -1
				if (!aIsPK && bIsPK) return 1
				return a.localeCompare(b)
			})
			.map(key => {
				const types = properties[key.toLowerCase()].type
				// Get the first non-null type as primary type
				const primaryType = Array.isArray(types)
					? types.find(t => t !== "null") || types[0]
					: types

				const isPrimaryKey = primaryKeys.includes(key)

				return {
					label: (
						<div className="flex w-full items-center justify-between whitespace-nowrap">
							<Tooltip title={key}>
								<span className="truncate">{key}</span>
							</Tooltip>
							<div className="flex shrink-0 items-center gap-2">
								{isPrimaryKey && (
									<span className="rounded bg-blue-100 px-1 py-0.5 text-xs text-blue-700">
										PK
									</span>
								)}
								<span className="rounded border border-gray-200 px-2 py-0.5 text-xs text-gray-600">
									{primaryType}
								</span>
							</div>
						</div>
					),
					value: key,
				}
			})
	}

	const formatFilterValue = (columnName: string, value: string) => {
		const properties = stream.stream.type_schema?.properties || {}
		const columnType = properties[columnName]?.type
		const primaryType = Array.isArray(columnType)
			? columnType.find(t => t !== "null") || columnType[0]
			: columnType

		if (primaryType === "string" || primaryType === "timestamp") {
			// Check if value is already wrapped in quotes
			if (!value.startsWith('"') && !value.endsWith('"')) {
				return `"${value}"`
			}
		}
		return value
	}

	const getColumnOptionsForCursor = (
		isFallback: boolean = false,
	): { label: React.ReactNode; value: string }[] => {
		const availableCursorFields = stream.stream.available_cursor_fields || []
		const selectedField = stream.stream.cursor_field?.split(":")[0]

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
							<span className="text-[#203FDD]">PK</span>
						)}
					</div>
				),
				value: field,
			}))
	}

	// Tab button component
	const TabButton = ({
		id,
		label,
		icon,
	}: {
		id: string
		label: string
		icon: React.ReactNode
	}) => {
		const tabStyle =
			activeTab === id
				? TAB_STYLES.active
				: `${TAB_STYLES.inactive} ${TAB_STYLES.hover}`

		return (
			<button
				className={`${tabStyle} flex items-center justify-center gap-1 text-xs`}
				style={{ fontWeight: 500, height: "28px", width: "100%" }}
				onClick={() => setActiveTab(id)}
				type="button"
			>
				<span className="flex items-center">{icon}</span>
				{label}
			</button>
		)
	}

	// Content rendering components
	const renderConfigContent = () => {
		return (
			<div className="flex flex-col gap-4">
				<div className={CARD_STYLE}>
					<div className="mb-4">
						<label className="mb-3 block w-full font-medium text-[#575757]">
							Sync mode:
						</label>
						<Radio.Group
							className="mb-4 grid grid-cols-2 gap-4"
							value={syncMode}
							onChange={e => handleSyncModeChange(e.target.value)}
						>
							<Radio
								value="full"
								disabled={!isSyncModeSupported("full_refresh")}
							>
								Full Refresh
							</Radio>
							<Radio
								value="incremental"
								disabled={!isSyncModeSupported("incremental")}
							>
								Full Refresh + Incremental
							</Radio>
							<Radio
								value="cdc"
								disabled={!isSyncModeSupported("cdc")}
							>
								Full Refresh + CDC
							</Radio>
							<Radio
								value="strict_cdc"
								disabled={!isSyncModeSupported("strict_cdc")}
							>
								CDC Only
							</Radio>
						</Radio.Group>
						{syncMode === "incremental" &&
							stream.stream.available_cursor_fields && (
								<div className="mb-4 mr-2">
									<div className="flex w-full gap-4">
										<div className="flex w-1/2 flex-col">
											<label className="mb-1 flex items-center gap-1 font-medium text-[#575757]">
												Cursor field:
												<Tooltip title="Column for identifying new/updated records ">
													<Info className="size-3.5 cursor-pointer" />
												</Tooltip>
											</label>
											<Select
												placeholder="Select cursor field"
												value={stream.stream.cursor_field?.split(":")[0]}
												onChange={(value: string) => {
													const newCursorField = fallBackCursorField
														? `${value}:${fallBackCursorField}`
														: value
													stream.stream.cursor_field = newCursorField
													setFallBackCursorField("")
													onSyncModeChange?.(
														stream.stream.name,
														stream.stream.namespace || "",
														"incremental",
													)
												}}
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
										{stream.stream.cursor_field &&
											!showFallbackSelector &&
											!fallBackCursorField && (
												<div className="flex w-1/2 items-end">
													<Tooltip title="Alternative cursor column in case cursor column encounters null values">
														<Button
															type="default"
															icon={<Plus className="size-4" />}
															onClick={() => setShowFallbackSelector(true)}
															className="mb-[2px] flex items-center gap-1"
														>
															Add Fallback Cursor
														</Button>
													</Tooltip>
												</div>
											)}

										{stream.stream.cursor_field &&
											(showFallbackSelector || fallBackCursorField) && (
												<div className="flex w-1/2 flex-col">
													<label className="mb-1 flex items-center gap-1 font-medium text-[#575757]">
														Fallback Cursor:
														<Tooltip title="Alternative cursor column in case cursor column encounters null values">
															<Info className="size-3.5 cursor-pointer text-[#575757]" />
														</Tooltip>
													</label>
													<Select
														placeholder="Select default"
														value={fallBackCursorField}
														onChange={(value: string) => {
															const newCursorField = value
																? `${stream.stream.cursor_field}:${value}`
																: stream.stream.cursor_field
															stream.stream.cursor_field = newCursorField
															setFallBackCursorField(value)
															onSyncModeChange?.(
																stream.stream.name,
																stream.stream.namespace || "",
																"incremental",
															)
														}}
														allowClear
														onClear={() => {
															setShowFallbackSelector(false)
															setFallBackCursorField("")
															stream.stream.cursor_field =
																stream.stream.cursor_field?.split(":")[0]
															onSyncModeChange?.(
																stream.stream.name,
																stream.stream.namespace || "",
																"incremental",
															)
														}}
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
				</div>

				<div
					className={`${!isSelected ? "font-normal text-[#c1c1c1]" : "font-medium"} ${CARD_STYLE}`}
				>
					<div className="flex items-center justify-between">
						<label>Normalisation</label>
						<Switch
							checked={normalisation}
							onChange={handleNormalizationChange}
							disabled={!isSelected || isStreamInInitialSelection}
						/>
					</div>
				</div>
				{!isSelected && (
					<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
						<Info className="size-4" />
						Select the stream to configure Normalisation
					</div>
				)}

				<div
					className={`${!isSelected ? "font-normal text-[#c1c1c1]" : "font-medium"} ${CARD_STYLE} !p-0`}
				>
					<div className="flex items-center justify-between !p-3">
						<label className="">Data Filter</label>
						<Switch
							checked={fullLoadFilter}
							onChange={handleFullLoadFilterChange}
							disabled={!isSelected || isStreamInInitialSelection}
						/>
					</div>
					{fullLoadFilter && isSelected && (
						<>
							<Divider className="my-0 p-0" />
							{renderFilterContent()}
						</>
					)}
				</div>
				{!isSelected && (
					<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
						<Info className="size-4" />
						Select the stream to configure Data Filter
					</div>
				)}
			</div>
		)
	}

	const renderPartitioningContent = () => (
		<div className="flex flex-col gap-4">
			{renderPartitioningRegexContent()}
		</div>
	)

	const renderPartitioningRegexContent = () => (
		<>
			<div className="flex items-center gap-0.5">
				<div className="text-[#575757]">Partitioning regex:</div>

				<Tooltip title={PartioningRegexTooltip}>
					<Info className="size-5 cursor-help items-center pt-1 text-gray-400" />
				</Tooltip>
				<a
					href={
						destinationType === "s3"
							? "https://olake.io/docs/writers/parquet/partitioning"
							: "https://olake.io/docs/writers/iceberg/partitioning"
					}
					target="_blank"
					rel="noopener noreferrer"
					className="flex items-center text-[#203FDD] hover:text-[#203FDD]/80"
				>
					<ArrowSquareOut className="size-5" />
				</a>
			</div>
			{isSelected ? (
				<>
					<Input
						placeholder="Enter your partition regex"
						className="w-full"
						value={partitionRegex}
						onChange={e => setPartitionRegex(e.target.value)}
						disabled={!!activePartitionRegex || isStreamInInitialSelection}
					/>
					{!activePartitionRegex ? (
						<Button
							className="mt-2 w-fit bg-[#203FDD] px-1 py-3 font-light text-white"
							onClick={handleSetPartitionRegex}
							disabled={!partitionRegex || isStreamInInitialSelection}
						>
							Set Partition
						</Button>
					) : (
						<div className="mt-4">
							<div className="text-sm text-[#575757]">
								Active partition regex:
							</div>
							<div className="mt-2 flex items-center justify-between text-sm">
								<span>{activePartitionRegex}</span>
								<Button
									type="text"
									danger
									size="small"
									className="rounded-[6px] py-1 text-sm"
									onClick={handleClearPartitionRegex}
									disabled={isStreamInInitialSelection}
								>
									Delete Partition
								</Button>
							</div>
						</div>
					)}
				</>
			) : (
				<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
					<Info className="size-4" />
					Select the stream to configure Partitioning
				</div>
			)}
		</>
	)

	const renderFilterContent = () => (
		<div className="flex flex-col gap-4 !p-3">
			{multiFilterCondition.conditions.map((condition, index) => (
				<div key={index}>
					{index > 0 && (
						<div className="mb-4 flex items-center justify-between">
							<div className="flex rounded-md bg-[#e9ebfc] p-1">
								<button
									type="button"
									onClick={() => handleLogicalOperatorChange("and")}
									className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
										multiFilterCondition.logicalOperator === "and"
											? "bg-white text-gray-800 shadow-sm"
											: "bg-transparent text-gray-600"
									}`}
									disabled={isStreamInInitialSelection}
								>
									AND
								</button>
								<button
									type="button"
									onClick={() => handleLogicalOperatorChange("or")}
									className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
										multiFilterCondition.logicalOperator === "or"
											? "bg-white text-gray-800 shadow-sm"
											: "bg-transparent text-gray-600"
									}`}
									disabled={isStreamInInitialSelection}
								>
									OR
								</button>
							</div>
							<Button
								type="text"
								danger
								icon={<X className="size-4" />}
								onClick={() => handleRemoveFilter(index)}
								disabled={isStreamInInitialSelection}
							>
								Remove
							</Button>
						</div>
					)}
					<div className="mb-4">
						<div className="mb-2 text-sm font-medium text-[#575757]">
							Column {index === 0 ? "I" : "II"}
						</div>
						{index === 0 && (
							<div className="mb-4 flex items-center gap-1 rounded-lg bg-[#FFF7E6] p-2 text-[#FFF6D5]">
								<Lightning className="size-4 font-bold text-[#DAAC06]" />
								<div className="text-[#6E5807]">
									Selecting indexed columns will enhance performance
								</div>
							</div>
						)}
					</div>
					<div className="grid grid-cols-[50%_15%_30%] gap-4">
						<div>
							<label className="mb-2 block text-sm text-[#575757]">
								Column Name
							</label>
							<Select
								className="w-full"
								placeholder="Select Column"
								value={condition.columnName}
								onChange={value =>
									handleFilterConditionChange(index, "columnName", value)
								}
								options={getColumnOptions()}
								labelInValue={false}
								optionLabelProp="value"
								disabled={isStreamInInitialSelection}
							/>
						</div>
						<div>
							<label className="mb-2 block text-sm text-[#575757]">
								Operator
							</label>
							<Select
								className="w-full"
								placeholder="Select"
								value={condition.operator}
								onChange={value =>
									handleFilterConditionChange(index, "operator", value)
								}
								options={operatorOptions}
								disabled={isStreamInInitialSelection}
							/>
						</div>
						<div>
							<label className="mb-2 block text-sm text-gray-600">Value</label>
							<Input
								placeholder="Enter value"
								value={condition.value}
								onChange={e =>
									handleFilterConditionChange(index, "value", e.target.value)
								}
								disabled={isStreamInInitialSelection}
							/>
						</div>
					</div>
				</div>
			))}
			{multiFilterCondition.conditions.length < 2 && (
				<Button
					type="default"
					icon={<Plus className="size-4" />}
					onClick={handleAddFilter}
					className="w-fit"
					disabled={isStreamInInitialSelection}
				>
					New Column filter
				</Button>
			)}
		</div>
	)

	// Main render
	return (
		<div>
			<div className="pb-4 font-medium capitalize">{stream.stream.name}</div>
			<div className="mb-4 w-full">
				<div className="grid grid-cols-3 gap-1 rounded-[6px] bg-[#F5F5F5] p-1">
					<TabButton
						id="config"
						label="Config"
						icon={<SlidersHorizontal className="size-3.5" />}
					/>
					<TabButton
						id="schema"
						label="Schema"
						icon={<ColumnsPlusRight className="size-3.5" />}
					/>
					<TabButton
						id="partitioning"
						label="Partitioning"
						icon={<GridFour className="size-3.5" />}
					/>
				</div>
			</div>

			{activeTab === "config" && renderConfigContent()}
			{activeTab === "schema" && <StreamsSchema initialData={stream} />}
			{activeTab === "partitioning" && renderPartitioningContent()}
		</div>
	)
}

export default StreamConfiguration
