import { ArrowSquareOutIcon, InfoIcon } from "@phosphor-icons/react"
import { Checkbox, Input, Switch, Tooltip } from "antd"
import { CheckboxChangeEvent } from "antd/es/checkbox/Checkbox"
import { useState, useEffect } from "react"

import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	useStreamSelectionStore,
} from "../../stores"
import { isColumnSelectionSupported, isColumnEnabled } from "../../utils"
import RenderTypeItems from "../RenderTypeItems"

const StreamsSchema = () => {
	const streamData = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const updateSelectedColumns = useStreamSelectionStore(
		state => state.updateSelectedColumns,
	)

	const typeSchemaProperties = streamData?.stream.type_schema?.properties || {}

	const [columnsToDisplay, setColumnsToDisplay] =
		useState<Record<string, any>>(typeSchemaProperties)

	// Re-sync columns to display when the user switches to a different stream
	useEffect(() => {
		setColumnsToDisplay(typeSchemaProperties)
	}, [streamData?.stream.name, streamData?.stream.namespace])

	if (!streamData || !selectedStream) return null

	const isOlakeColumn = (name: string) =>
		typeSchemaProperties[name]?.olake_column === true

	const columnSelectionSupported = isColumnSelectionSupported(selectedStream)

	// Column selection is editable only if it supported and stream is enabled
	const isEditable = columnSelectionSupported && !selectedStream.disabled

	const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const query = event.target.value
		const props = typeSchemaProperties
		if (!props) return
		const filtered = Object.entries(props).filter(([key]) =>
			key.toLowerCase().includes(query.toLowerCase()),
		)
		setColumnsToDisplay(Object.fromEntries(filtered))
	}

	const handleSelectAll = (e: CheckboxChangeEvent) => {
		if (!isEditable || !selectedStream.selected_columns) return
		const selectAll = e.target.checked
		const current = selectedStream.selected_columns
		const visibleColumnNames = Object.keys(columnsToDisplay)

		let newColumns: string[]

		if (selectAll) {
			newColumns = [...new Set([...current.columns, ...visibleColumnNames])]
		} else {
			newColumns = current.columns.filter(
				c => !visibleColumnNames.includes(c) || isOlakeColumn(c),
			)
		}

		updateSelectedColumns(
			streamData.stream.name,
			streamData.stream.namespace || "",
			{ ...current, columns: newColumns },
		)
	}

	const handleColumnSelect = (columnName: string, checked: boolean) => {
		if (
			!isEditable ||
			isOlakeColumn(columnName) ||
			!selectedStream.selected_columns
		)
			return

		const current = selectedStream.selected_columns
		const newColumns = checked
			? [...new Set([...current.columns, columnName])]
			: current.columns.filter(c => c !== columnName)

		updateSelectedColumns(
			streamData.stream.name,
			streamData.stream.namespace || "",
			{ ...current, columns: newColumns },
		)
	}

	const handleSyncNewColumnsChange = (checked: boolean) => {
		if (!isEditable || !selectedStream.selected_columns) return
		updateSelectedColumns(
			streamData.stream.name,
			streamData.stream.namespace || "",
			{
				...selectedStream.selected_columns,
				sync_new_columns: checked,
			},
		)
	}

	const visibleNonLocked = Object.keys(columnsToDisplay).filter(
		name => !isOlakeColumn(name),
	)
	const isAllSelected =
		visibleNonLocked.length > 0 &&
		visibleNonLocked.every(name => isColumnEnabled(name, selectedStream))

	const hasDestinationColumns = Object.values(columnsToDisplay || {}).some(
		columnSchema => columnSchema?.destination_column_name,
	)

	return (
		<div className="rounded-xl border border-[#E3E3E3] bg-white p-4">
			{columnSelectionSupported && (
				<div className="mb-3 flex items-center justify-between gap-x-1 rounded-lg border border-[#E3E3E3] px-3 py-2">
					<div className="space-y-1">
						<div className="flex items-center gap-x-2 text-sm font-medium text-neutral-text">
							Sync new columns automatically
							<Tooltip
								title="View Documentation"
								className="border-l px-2"
							>
								<a
									href="https://olake.io/docs/understanding/terminologies/olake/#sync-new-columns-automatically"
									target="_blank"
									rel="noopener noreferrer"
									className="flex items-center text-gray-600 transition-colors hover:text-primary"
								>
									<ArrowSquareOutIcon className="size-4" />
								</a>
							</Tooltip>
						</div>
						<p className="text-xs text-gray-500">
							When enabled, newly added columns in the source will be synced
							automatically.
						</p>
					</div>
					<Switch
						checked={!!selectedStream.selected_columns?.sync_new_columns}
						onChange={handleSyncNewColumnsChange}
						disabled={!isEditable}
					/>
				</div>
			)}

			{/* Search */}
			<div className="mb-3 flex items-center gap-x-2">
				<div className="w-full border-r pr-3">
					<Input.Search
						className="custom-search-input w-full"
						placeholder="Search Columns"
						allowClear
						onChange={handleSearchChange}
					/>
				</div>
				<div className="flex h-full items-center gap-x-2">
					<Tooltip title="Enable only specific columns">
						<InfoIcon className="size-5 cursor-help items-center pt-1 text-gray-500" />
					</Tooltip>
					<Tooltip title="View Documentation">
						<a
							href="https://olake.io/docs/understanding/terminologies/olake/#4-schema"
							target="_blank"
							rel="noopener noreferrer"
							className="flex items-center text-primary hover:text-primary/80"
						>
							<ArrowSquareOutIcon className="size-5" />
						</a>
					</Tooltip>
				</div>
			</div>

			<div className="max-h-[400px] overflow-auto rounded border border-[#d9d9d9]">
				{/* Table Header */}
				<div className="flex items-center border-b border-gray-400 bg-gray-50 px-4 py-3">
					<div className="flex w-16 items-center justify-center">
						<Checkbox
							checked={isAllSelected}
							onChange={handleSelectAll}
							disabled={!isEditable}
						/>
					</div>
					<div className="flex-1 px-2 text-left font-medium text-gray-700">
						Column name
					</div>
					{hasDestinationColumns && (
						<div className="flex-1 px-2 text-left font-medium text-gray-700">
							Destination Column
						</div>
					)}
					<div className="flex-1 px-2 text-left font-medium text-gray-700">
						Source Data type
					</div>
				</div>

				{/* Data Rows */}
				{Object.keys(columnsToDisplay || {}).map(item => {
					const columnSchema = columnsToDisplay[item]
					const destinationColumnName =
						columnSchema?.destination_column_name || item
					const checked = columnSelectionSupported
						? isColumnEnabled(item, selectedStream)
						: true
					// Disabled when stream is unselected, driver is legacy, or column is locked (olake column)
					const checkboxDisabled =
						!isEditable || columnSchema?.olake_column === true
					const checkbox = (
						<Checkbox
							checked={checked}
							onChange={e => handleColumnSelect(item, e.target.checked)}
							disabled={checkboxDisabled}
						/>
					)

					return (
						<div
							key={item}
							className="flex items-center border-b border-gray-400 px-4 py-3 last:border-b-0 hover:bg-background-primary"
						>
							<div className="flex w-16 items-center justify-center">
								{isOlakeColumn(item) ? (
									<Tooltip title="OLake generated column. It is mandatory and cannot be deselected.">
										{checkbox}
									</Tooltip>
								) : (
									checkbox
								)}
							</div>
							<div className="flex-1 px-2 text-left">
								<Tooltip title={item}>
									<span className="block">
										{item.length > 13 ? `${item.substring(0, 13)}...` : item}
									</span>
								</Tooltip>
							</div>
							{hasDestinationColumns && (
								<div className="flex-1 px-2 text-left">
									<Tooltip title={destinationColumnName}>
										<span className="block">
											{destinationColumnName.length > 13
												? `${destinationColumnName.substring(0, 13)}...`
												: destinationColumnName}
										</span>
									</Tooltip>
								</div>
							)}
							<div className="flex-1 px-2 text-left">
								<RenderTypeItems
									initialList={typeSchemaProperties}
									item={item}
								/>
							</div>
						</div>
					)
				})}
			</div>
		</div>
	)
}

export default StreamsSchema
