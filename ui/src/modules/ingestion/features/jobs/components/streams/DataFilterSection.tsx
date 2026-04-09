import {
	ArrowSquareOutIcon,
	InfoIcon,
	LightningIcon,
	PlusIcon,
	XIcon,
} from "@phosphor-icons/react"
import { Button, Divider, Input, message, Select, Switch, Tooltip } from "antd"
import clsx from "clsx"
import { useEffect, useRef, useState } from "react"

import {
	FilterConfig,
	FilterConfigCondition,
	FilterOperator,
	LogicalOperator,
	MultiFilterCondition,
} from "@/modules/ingestion/common/types"

import { CARD_STYLE, operatorOptions } from "../../constants"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
	selectStreamFilterState,
	selectUseFilterConfig,
	useStreamSelectionStore,
} from "../../stores"

const DataFilterSection = () => {
	const updateFilter = useStreamSelectionStore(state => state.updateFilter)
	const updateFilterConfig = useStreamSelectionStore(
		state => state.updateFilterConfig,
	)
	const setStreamFilterState = useStreamSelectionStore(
		state => state.setStreamFilterState,
	)
	const useFilterConfig = useStreamSelectionStore(selectUseFilterConfig)
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, stream),
	)

	// Unique stream key to differentiate a stream with same name and different namespace
	const streamKey = stream
		? `${stream.stream.namespace || ""}_${stream.stream.name}`
		: ""
	const streamFilterState = useStreamSelectionStore(
		selectStreamFilterState(streamKey),
	)

	const [isFilterEnabled, setIsFilterEnabled] = useState<boolean>(false)
	const [multiFilterCondition, setMultiFilterCondition] =
		useState<MultiFilterCondition>({
			conditions: [
				{
					column: "",
					operator: "=",
					value: "",
				},
			],
			logicalOperator: "and",
		})

	// Guard to prevent prop-driven effect from clobbering local edits
	const isLocalFilterUpdateRef = useRef(false)

	// Filter parsing effect — re-runs when the active stream changes or its filter/filter_config changes
	const currentFilter = selectedStream?.filter || ""
	const currentFilterConfig = selectedStream?.filter_config

	useEffect(() => {
		// Skip when change originated from local user action
		if (isLocalFilterUpdateRef.current) {
			isLocalFilterUpdateRef.current = false
			return
		}

		if (useFilterConfig) {
			if (currentFilterConfig && currentFilterConfig.conditions.length > 0) {
				setMultiFilterCondition({
					conditions: currentFilterConfig.conditions,
					logicalOperator: currentFilterConfig.logical_operator,
				})
				setIsFilterEnabled(true)
				setStreamFilterState(streamKey, true)
			} else {
				setMultiFilterCondition({
					conditions: [{ column: "", operator: "=", value: null }],
					logicalOperator: "and",
				})
				const savedFilterState = streamFilterState || false
				setIsFilterEnabled(savedFilterState)
			}
			return
		}

		// Legacy filter string path
		if (currentFilter) {
			const conditions: FilterConfigCondition[] = []
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
						column: columnName.trim(),
						operator,
						value: cleanValue,
					})
				}
			})

			if (conditions.length > 0) {
				setMultiFilterCondition({ conditions, logicalOperator })
				setIsFilterEnabled(true)
				// Persist the filter enabled state for this stream
				setStreamFilterState(streamKey, true)
			}
		} else {
			setMultiFilterCondition({
				conditions: [{ column: "", operator: "=", value: "" }],
				logicalOperator: "and",
			})
			// Restore filter state for this stream or default to false
			setIsFilterEnabled(streamFilterState)
		}
	}, [currentFilter, currentFilterConfig])

	if (!stream || !selectedStream) return null

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
				const primaryType = types.find(t => t !== "null") || types[0]
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
		if (!value) return value ?? ""

		const properties = stream.stream.type_schema?.properties || {}
		const columnType = properties[columnName]?.type
		const primaryType = columnType?.find(t => t !== "null") || columnType?.[0]

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
	const handleFilterEnabledChange = (checked: boolean) => {
		setIsFilterEnabled(checked)
		// Persist the filter state for this stream
		setStreamFilterState(streamKey, checked)

		setMultiFilterCondition({
			conditions: [
				{
					column: "",
					operator: "=",
					value: null,
				},
			],
			logicalOperator: "and",
		})
		isLocalFilterUpdateRef.current = true

		if (useFilterConfig) {
			updateFilterConfig(
				stream.stream.name,
				stream.stream.namespace || "",
				checked ? { logical_operator: "and", conditions: [] } : undefined,
			)
		} else {
			updateFilter(
				stream.stream.name,
				stream.stream.namespace || "",
				checked ? "=" : "",
			)
		}
	}

	const handleFilterConditionChange = (
		index: number,
		field: keyof FilterConfigCondition,
		value: string,
	) => {
		const newConditions = [...multiFilterCondition.conditions]
		newConditions[index] = { ...newConditions[index], [field]: value }

		const newMultiCondition = {
			...multiFilterCondition,
			conditions: newConditions,
		}
		setMultiFilterCondition(newMultiCondition)
		isLocalFilterUpdateRef.current = true

		if (useFilterConfig) {
			const filterConfig: FilterConfig = {
				logical_operator: newMultiCondition.logicalOperator,
				conditions: newConditions,
			}
			updateFilterConfig(
				stream.stream.name,
				stream.stream.namespace || "",
				filterConfig,
			)
		} else {
			const filterString = newConditions
				.map(
					cond =>
						`${cond.column} ${cond.operator} ${formatFilterValue(cond.column, cond.value as string)}`,
				)
				.join(` ${newMultiCondition.logicalOperator} `)
			updateFilter(
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
			cond => cond.column && cond.operator,
		)

		if (filledConditions.length > 1) {
			isLocalFilterUpdateRef.current = true
			if (useFilterConfig) {
				const filterConfig: FilterConfig = {
					logical_operator: value,
					conditions: filledConditions,
				}
				updateFilterConfig(
					stream.stream.name,
					stream.stream.namespace || "",
					filterConfig,
				)
			} else {
				const filterString = filledConditions
					.map(
						cond =>
							`${cond.column} ${cond.operator} ${formatFilterValue(cond.column, cond.value as string)}`,
					)
					.join(` ${value} `)
				updateFilter(
					stream.stream.name,
					stream.stream.namespace || "",
					filterString,
				)
			}
		}
	}

	const handleAddFilter = () => {
		const { conditions } = multiFilterCondition

		if (conditions.length >= 2) return

		const firstCondition = conditions[0]
		if (!firstCondition.column || !firstCondition.operator) {
			message.error("Please complete the first filter before applying another.")
			return
		}

		const newConditions = [
			...conditions,
			{ column: "", operator: "=" as FilterOperator, value: null },
		]
		setMultiFilterCondition({
			...multiFilterCondition,
			conditions: newConditions,
		})

		isLocalFilterUpdateRef.current = true
		if (useFilterConfig) {
			const filterConfig: FilterConfig = {
				logical_operator: multiFilterCondition.logicalOperator,
				conditions: newConditions,
			}
			updateFilterConfig(
				stream.stream.name,
				stream.stream.namespace || "",
				filterConfig,
			)
		} else {
			const filterString =
				conditions
					.map(
						cond =>
							`${cond.column} ${cond.operator} ${formatFilterValue(cond.column, cond.value as string)}`,
					)
					.join(` ${multiFilterCondition.logicalOperator} `) + " = "
			updateFilter(
				stream.stream.name,
				stream.stream.namespace || "",
				filterString,
			)
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
			isLocalFilterUpdateRef.current = true

			if (useFilterConfig) {
				if (condition.column && condition.operator) {
					const filterConfig: FilterConfig = {
						logical_operator: newMultiCondition.logicalOperator,
						conditions: newConditions,
					}
					updateFilterConfig(
						stream.stream.name,
						stream.stream.namespace || "",
						filterConfig,
					)
				} else {
					// Remaining condition is incomplete — clear filter_config
					updateFilterConfig(
						stream.stream.name,
						stream.stream.namespace || "",
						undefined,
					)
				}
			} else {
				if (condition.column && condition.operator) {
					const filterString = `${condition.column} ${condition.operator} ${formatFilterValue(condition.column, condition.value as string)}`
					updateFilter(
						stream.stream.name,
						stream.stream.namespace || "",
						filterString,
					)
				} else {
					updateFilter(stream.stream.name, stream.stream.namespace || "", "")
				}
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
					<div className="flex items-center gap-1">
						<label>Data Filter</label>
						<Tooltip title="Filters the stream to include only records that match conditions on specific columns.">
							<InfoIcon
								size={14}
								className="cursor-help text-text-tertiary"
							/>
						</Tooltip>
						<a
							href="https://olake.io/docs/understanding/terminologies/olake/#3-data-filter"
							target="_blank"
							rel="noopener noreferrer"
							aria-label="Open data filter docs"
							className="inline-flex text-text-tertiary hover:text-primary"
						>
							<ArrowSquareOutIcon size={14} />
						</a>
					</div>
					<Switch
						checked={isFilterEnabled}
						onChange={handleFilterEnabledChange}
						disabled={!isSelected}
					/>
				</div>
				{isFilterEnabled && (
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
												value={condition.column}
												onChange={value =>
													handleFilterConditionChange(index, "column", value)
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
												placeholder={condition.value === null ? "<null>" : ""}
												value={condition.value === null ? "" : condition.value}
												onFocus={() => {
													if (condition.value === null) {
														handleFilterConditionChange(index, "value", "")
													}
												}}
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
