import { useEffect, useState } from "react"
import {
	ExtendedStreamConfigurationProps,
	FilterCondition,
	FilterOperator,
	LogicalOperator,
	MultiFilterCondition,
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
} from "@phosphor-icons/react"
import { CARD_STYLE, TAB_STYLES } from "../../../../utils/constants"
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
}: ExtendedStreamConfigurationProps) => {
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState(
		stream.stream.sync_mode === "full_refresh"
			? "full"
			: stream.stream.sync_mode === "incremental"
				? "incremental"
				: "cdc",
	)
	const [enableBackfill, setEnableBackfill] = useState(false)
	const [normalisation, setNormalisation] =
		useState<boolean>(initialNormalization)
	const [fullLoadFilter, setFullLoadFilter] = useState<boolean>(false)
	const [partitionRegex, setPartitionRegex] = useState("")
	const [defaultCursorField, setDefaultCursorField] = useState<string>("")
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
			setDefaultCursorField(defaultField)
		} else {
			setDefaultCursorField("")
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
		setEnableBackfill(initialEnableBackfillForSwitch)
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
	}, [
		stream,
		initialNormalization,
		initialPartitionRegex,
		initialFullLoadFilter,
	])

	// Handlers
	const handleSyncModeChange = (selectedRadioValue: string) => {
		setSyncMode(selectedRadioValue)
		let newApiSyncMode: "full_refresh" | "cdc" | "incremental" | "strict_cdc"
		let newEnableBackfillState = true
		if (selectedRadioValue === "full") {
			newApiSyncMode = "full_refresh"
		} else if (selectedRadioValue === "incremental") {
			newApiSyncMode = "incremental"
		} else {
			newApiSyncMode = "cdc"
		}

		stream.stream.sync_mode = newApiSyncMode
		setEnableBackfill(newEnableBackfillState)
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

	const handleEnableBackfillChange = (checked: boolean) => {
		setEnableBackfill(checked)
		let finalApiSyncMode = stream.stream.sync_mode

		if (syncMode === "cdc") {
			if (checked) {
				finalApiSyncMode = "cdc"
				stream.stream.sync_mode = "cdc"
				onSyncModeChange?.(
					stream.stream.name,
					stream.stream.namespace || "default",
					"cdc",
				)
			} else {
				finalApiSyncMode = "strict_cdc"
				stream.stream.sync_mode = "strict_cdc"
			}
		}

		setFormData({
			...formData,
			backfill: checked,
			sync_mode: finalApiSyncMode,
		})
	}

	const handleNormalizationChange = (checked: boolean) => {
		setNormalisation(checked)
		onNormalizationChange(
			stream.stream.name,
			stream.stream.namespace || "default",
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
				stream.stream.namespace || "default",
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
			stream.stream.namespace || "default",
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
				stream.stream.namespace || "default",
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
				stream.stream.namespace || "default",
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
				stream.stream.namespace || "default",
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
					stream.stream.namespace || "default",
					filterString,
				)
			} else {
				onFullLoadFilterChange?.(
					stream.stream.name,
					stream.stream.namespace || "default",
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

		// Combine fields in priority order, filter out duplicates
		const orderedFields = [
			...primaryKeys,
			...cursorFields,
			...Object.keys(properties),
		]

		// Convert to unique array while preserving order
		return [...new Set(orderedFields)]
			.filter(key => properties[key])
			.map(key => {
				const types = properties[key].type
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

	const getFilteredOperatorOptions = (columnName: string) => {
		const properties = stream.stream.type_schema?.properties || {}
		const columnType = properties[columnName]?.type
		const primaryType = Array.isArray(columnType)
			? columnType.find(t => t !== "null") || columnType[0]
			: columnType

		if (primaryType === "string") {
			return operatorOptions.filter(op => op.value === "=" || op.value === "!=")
		}
		return operatorOptions
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

	const getFieldType = (fieldName: string): string => {
		const properties = stream.stream.type_schema?.properties || {}
		const fieldType = properties[fieldName]?.type
		return Array.isArray(fieldType)
			? fieldType.find(t => t !== "null") || fieldType[0]
			: fieldType
	}

	const getFieldsOfSameType = (selectedField: string): string[] => {
		const selectedType = getFieldType(selectedField)
		const availableCursorFields = stream.stream.available_cursor_fields || []

		return availableCursorFields.filter(field => {
			// Skip the currently selected cursor field
			if (field === selectedField) return false

			// Check if field has the same type
			const fieldType = getFieldType(field)
			return fieldType === selectedType
		})
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
							className="mb-4 flex w-full items-center"
							value={syncMode}
							onChange={e => handleSyncModeChange(e.target.value)}
						>
							<Radio
								value="full"
								className="w-1/3"
							>
								Full refresh
							</Radio>
							<Radio
								value="cdc"
								className="w-1/3"
							>
								CDC
							</Radio>
							<Radio
								value="incremental"
								className="w-1/3"
								disabled={
									!stream.stream.supported_sync_modes ||
									!stream.stream.supported_sync_modes.some(
										mode => mode === "incremental",
									)
								}
							>
								Incremental
							</Radio>
						</Radio.Group>
						{syncMode === "incremental" &&
							stream.stream.available_cursor_fields && (
								<div className="mb-4 mr-2">
									<div className="flex w-full gap-4">
										<div className="flex w-1/2 flex-col">
											<label className="mb-1 font-medium text-[#575757]">
												Cursor field:
											</label>
											<Select
												placeholder="Select cursor field"
												value={stream.stream.cursor_field?.split(":")[0]}
												onChange={(value: string) => {
													const newCursorField = defaultCursorField
														? `${value}:${defaultCursorField}`
														: value
													stream.stream.cursor_field = newCursorField
													setDefaultCursorField("")
													onSyncModeChange?.(
														stream.stream.name,
														stream.stream.namespace || "",
														"incremental",
													)
												}}
												optionLabelProp="label"
											>
												{[...stream.stream.available_cursor_fields]
													.sort((a, b) => {
														const aIsPK =
															stream.stream.source_defined_primary_key?.includes(
																a,
															) || false
														const bIsPK =
															stream.stream.source_defined_primary_key?.includes(
																b,
															) || false
														if (aIsPK && !bIsPK) return -1
														if (!aIsPK && bIsPK) return 1
														return a.localeCompare(b)
													})
													.map((field: string) => (
														<Select.Option
															key={field}
															value={field}
															label={field}
														>
															<div className="flex items-center justify-between">
																<span>{field}</span>
																{stream.stream.source_defined_primary_key?.includes(
																	field,
																) && <span className="text-[#203FDD]">PK</span>}
															</div>
														</Select.Option>
													))}
											</Select>
										</div>
										{stream.stream.cursor_field && (
											<div className="flex w-1/2 flex-col">
												<label className="mb-1 font-medium text-[#575757]">
													Default:
												</label>
												<Select
													placeholder="Select default"
													value={defaultCursorField}
													onChange={(value: string) => {
														const newCursorField = value
															? `${stream.stream.cursor_field}:${value}`
															: stream.stream.cursor_field
														stream.stream.cursor_field = newCursorField
														setDefaultCursorField(value)
														onSyncModeChange?.(
															stream.stream.name,
															stream.stream.namespace || "",
															"incremental",
														)
													}}
													allowClear
													optionLabelProp="label"
												>
													{getFieldsOfSameType(
														stream.stream.cursor_field?.split(":")[0] ||
															stream.stream.cursor_field,
													).map((field: string) => (
														<Select.Option
															key={field}
															value={field}
															label={field}
														>
															<div className="flex items-center justify-between">
																<span>{field}</span>
																{stream.stream.source_defined_primary_key?.includes(
																	field,
																) && <span className="text-[#203FDD]">PK</span>}
															</div>
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
				<div className={CARD_STYLE}>
					<div className="flex items-center justify-between">
						<label className="font-medium">Enable backfill</label>
						<Switch
							className="text-[#c1c1c1]"
							checked={enableBackfill}
							onChange={handleEnableBackfillChange}
							disabled={
								syncMode === "full" ||
								syncMode === "incremental" ||
								fromJobEditFlow
							}
						/>
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
							disabled={!isSelected || fromJobEditFlow}
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
						<label className="">Full Load Filter</label>
						<Switch
							checked={fullLoadFilter}
							onChange={handleFullLoadFilterChange}
							disabled={!isSelected || fromJobEditFlow}
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
						Select the stream to configure Full Load Filter
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
			<div className="text-[#575757]">Partitioning regex:</div>
			{isSelected ? (
				<>
					<Input
						placeholder="Enter your partition regex"
						className="w-full"
						value={partitionRegex}
						onChange={e => setPartitionRegex(e.target.value)}
						disabled={!!activePartitionRegex}
					/>
					{!activePartitionRegex ? (
						<Button
							className="mt-2 w-fit bg-[#203FDD] px-1 py-3 font-light text-white"
							onClick={handleSetPartitionRegex}
							disabled={!partitionRegex}
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
									disabled={fromJobEditFlow}
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
									disabled={fromJobEditFlow}
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
									disabled={fromJobEditFlow}
								>
									OR
								</button>
							</div>
							<Button
								type="text"
								danger
								icon={<X className="size-4" />}
								onClick={() => handleRemoveFilter(index)}
								disabled={fromJobEditFlow}
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
								disabled={fromJobEditFlow}
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
								options={getFilteredOperatorOptions(condition.columnName)}
								disabled={fromJobEditFlow}
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
								disabled={fromJobEditFlow}
							/>
						</div>
					</div>
				</div>
			))}
			{multiFilterCondition.conditions.length < 2 && !fromJobEditFlow && (
				<Button
					type="default"
					icon={<Plus className="size-4" />}
					onClick={handleAddFilter}
					className="w-fit"
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
