import { useEffect, useRef, useState } from "react"
import clsx from "clsx"
import { Button, Divider, Input, message, Select, Switch, Tooltip } from "antd"
import { LightningIcon, PlusIcon, XIcon, InfoIcon } from "@phosphor-icons/react"

import {
	FilterCondition,
	FilterOperator,
	LogicalOperator,
	MultiFilterCondition,
} from "../../types"
import { CARD_STYLE } from "../../constants"
import { operatorOptions } from "../../constants"

import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores"

const DataFilterSection = () => {
	const store = useStreamSelectionStore()
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, stream),
	)

	const [fullLoadFilter, setFullLoadFilter] = useState<boolean>(false)
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

	// Guard to prevent prop-driven effect from clobbering local edits
	const isLocalFilterUpdateRef = useRef(false)

	if (!stream || !selectedStream) return null

	// Unique stream key to differentiate a stream with same name and different namespace
	const streamKey = `${stream.stream.namespace || ""}_${stream.stream.name}`

	// Filter parsing effect to parse the filter string and set the filter state
	const currentFilter = selectedStream.filter || ""
	useEffect(() => {
		// Skip when change originated from local user action
		if (isLocalFilterUpdateRef.current) {
			isLocalFilterUpdateRef.current = false
			return
		}
		// Parse initial filter if exists
		if (currentFilter) {
			const conditions: FilterCondition[] = []
			let logicalOperator: LogicalOperator = "and"
			// Check for AND/OR operator
			const parts = currentFilter.toLowerCase().includes(" and ")
				? currentFilter.split(" and ")
				: currentFilter.split(" or ")

			if (parts.length > 1) {
				logicalOperator = currentFilter.toLowerCase().includes(" and ")
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
				setMultiFilterCondition({ conditions, logicalOperator })
				setFullLoadFilter(true)
				// Persist the filter enabled state for this stream
				store.setStreamFilterState(streamKey, true)
			}
		} else {
			setMultiFilterCondition({
				conditions: [{ columnName: "", operator: "=", value: "" }],
				logicalOperator: "and",
			})
			// Restore filter state for this stream or default to false
			const savedState = store.streamFilterStates[streamKey] || false
			setFullLoadFilter(savedState)
		}
	}, [currentFilter])

	// get columns based on primary keys and cursor fields and their properties
	const getColumnOptions = () => {
		const properties = stream.stream.type_schema?.properties || {}
		const primaryKeys = (stream.stream.source_defined_primary_key ||
			[]) as string[]
		const cursorFields = (stream.stream.available_cursor_fields ||
			[]) as string[]

		return cursorFields
			.filter(key => properties[key])
			.sort((a, b) => {
				const aIsPK = primaryKeys.includes(a)
				const bIsPK = primaryKeys.includes(b)
				if (aIsPK && !bIsPK) return -1
				if (!aIsPK && bIsPK) return 1
				return a.localeCompare(b)
			})
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

	// when the type is either string or timestamp we wrap the value in quotes
	const formatFilterValue = (columnName: string, value: string) => {
		const properties = stream.stream.type_schema?.properties || {}
		const columnType = properties[columnName]?.type
		const primaryType = Array.isArray(columnType)
			? columnType.find(t => t !== "null") || columnType[0]
			: columnType

		if (
			primaryType === "string" ||
			primaryType === "timestamp" ||
			primaryType === "timestamp_micro" ||
			primaryType === "timestamp_nano" ||
			primaryType === "timestamp_milli"
		) {
			// Check if value is already wrapped in quotes
			if (!value.startsWith('"') && !value.endsWith('"')) {
				return `"${value}"`
			}
		}
		return value
	}

	// Handlers
	const handleFullLoadFilterChange = (checked: boolean) => {
		setFullLoadFilter(checked)
		// Persist the filter state for this stream
		store.setStreamFilterState(streamKey, checked)

		setMultiFilterCondition({
			conditions: [{ columnName: "", operator: "=", value: "" }],
			logicalOperator: "and",
		})
		isLocalFilterUpdateRef.current = true
		// If toggled on insert empty condition
		store.updateFilter(
			stream.stream.name,
			stream.stream.namespace || "",
			checked ? "=" : "",
		)
	}

	const handleFilterConditionChange = (
		index: number,
		field: keyof FilterCondition,
		value: string,
	) => {
		const newConditions = [...multiFilterCondition.conditions]
		newConditions[index] = { ...newConditions[index], [field]: value }

		const newMultiCondition = {
			...multiFilterCondition,
			conditions: newConditions,
		}
		setMultiFilterCondition(newMultiCondition)

		const filterString = newConditions
			.map(
				cond =>
					`${cond.columnName} ${cond.operator} ${formatFilterValue(cond.columnName, cond.value)}`,
			)
			.join(` ${newMultiCondition.logicalOperator} `)

		isLocalFilterUpdateRef.current = true
		store.updateFilter(
			stream.stream.name,
			stream.stream.namespace || "",
			filterString,
		)
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

			isLocalFilterUpdateRef.current = true
			store.updateFilter(
				stream.stream.name,
				stream.stream.namespace || "",
				filterString,
			)
		}
	}

	const handleAddFilter = () => {
		const { conditions } = multiFilterCondition

		if (conditions.length >= 2) return

		const firstCondition = conditions[0]
		if (
			!firstCondition.columnName ||
			!firstCondition.operator ||
			!firstCondition.value
		) {
			message.error("Please complete the first filter before applying another.")
			return
		}

		setMultiFilterCondition({
			...multiFilterCondition,
			conditions: [...conditions, { columnName: "", operator: "=", value: "" }],
		})

		// insert empty condition in the filter string
		const filterString =
			conditions
				.map(
					cond =>
						`${cond.columnName} ${cond.operator} ${formatFilterValue(cond.columnName, cond.value)}`,
				)
				.join(` ${multiFilterCondition.logicalOperator} `) + " = "

		isLocalFilterUpdateRef.current = true
		store.updateFilter(
			stream.stream.name,
			stream.stream.namespace || "",
			filterString,
		)
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
				isLocalFilterUpdateRef.current = true
				store.updateFilter(
					stream.stream.name,
					stream.stream.namespace || "",
					filterString,
				)
			} else {
				isLocalFilterUpdateRef.current = true
				store.updateFilter(
					stream.stream.name,
					stream.stream.namespace || "",
					"",
				)
			}
		}
	}

	return (
		<>
			<div
				className={clsx(
					!isSelected ? "font-normal text-text-disabled" : "font-medium",
					CARD_STYLE,
					"!p-0",
				)}
			>
				<div className="flex items-center justify-between !p-3">
					<label>Data Filter</label>
					<Switch
						checked={fullLoadFilter}
						onChange={handleFullLoadFilterChange}
						disabled={!isSelected}
					/>
				</div>
				{fullLoadFilter && (
					<>
						<Divider className="my-0 p-0" />
						<div className="flex flex-col gap-4 !p-3">
							{multiFilterCondition.conditions.map((condition, index) => (
								<div key={index}>
									{index > 0 && (
										<div className="mb-4 flex items-center justify-between">
											<div className="flex rounded-md bg-primary-100 p-1">
												<button
													type="button"
													onClick={() => handleLogicalOperatorChange("and")}
													className={clsx(
														"rounded px-3 py-1 text-sm font-medium transition-colors",
														multiFilterCondition.logicalOperator === "and"
															? "bg-white text-gray-800 shadow-sm"
															: "bg-transparent text-gray-600",
													)}
													disabled={!isSelected}
												>
													AND
												</button>
												<button
													type="button"
													onClick={() => handleLogicalOperatorChange("or")}
													className={clsx(
														"rounded px-3 py-1 text-sm font-medium transition-colors",
														multiFilterCondition.logicalOperator === "or"
															? "bg-white text-gray-800 shadow-sm"
															: "bg-transparent text-gray-600",
													)}
													disabled={!isSelected}
												>
													OR
												</button>
											</div>
											<Button
												type="text"
												danger
												icon={<XIcon className="size-4" />}
												onClick={() => handleRemoveFilter(index)}
												disabled={!isSelected}
											>
												Remove
											</Button>
										</div>
									)}
									<div className="mb-4">
										<div className="mb-2 text-sm font-medium text-neutral-text">
											Column {index === 0 ? "I" : "II"}
										</div>
										{index === 0 && (
											<div className="mb-4 flex items-center gap-1 rounded-lg bg-warning-light p-2 text-warning-light">
												<LightningIcon className="size-4 font-bold text-warning" />
												<div className="text-warning-dark">
													Selecting indexed columns will enhance performance
												</div>
											</div>
										)}
									</div>
									<div className="grid grid-cols-[50%_15%_30%] gap-4">
										<div>
											<label className="mb-2 block text-sm text-neutral-text">
												Column Name
											</label>
											<Select
												className="w-full"
												placeholder="Select Column"
												value={condition.columnName}
												onChange={value =>
													handleFilterConditionChange(
														index,
														"columnName",
														value,
													)
												}
												options={getColumnOptions()}
												labelInValue={false}
												optionLabelProp="value"
												disabled={!isSelected}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm text-neutral-text">
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
												disabled={!isSelected}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm text-gray-600">
												Value
											</label>
											<Input
												placeholder="Enter value"
												value={condition.value}
												onChange={e =>
													handleFilterConditionChange(
														index,
														"value",
														e.target.value,
													)
												}
												disabled={!isSelected}
											/>
										</div>
									</div>
								</div>
							))}
							{multiFilterCondition.conditions.length < 2 && (
								<Button
									type="default"
									icon={<PlusIcon className="size-4" />}
									onClick={handleAddFilter}
									className="w-fit"
									disabled={!isSelected}
								>
									New Column filter
								</Button>
							)}
						</div>
					</>
				)}
			</div>
			{!isSelected && (
				<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4" />
					Select the stream to configure Data Filter
				</div>
			)}
		</>
	)
}

export default DataFilterSection
