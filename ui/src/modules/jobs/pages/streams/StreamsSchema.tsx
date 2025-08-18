import { useState, useEffect, useMemo } from "react"
import { Checkbox, Input, Tooltip } from "antd"
import { CheckboxChangeEvent } from "antd/es/checkbox/Checkbox"

import { StreamSchemaProps } from "../../../../types"
import RenderTypeItems from "../../../common/components/RenderTypeItems"

const StreamsSchema = ({ initialData, onColumnsChange }: StreamSchemaProps) => {
	const [columnsToDisplay, setColumnsToDisplay] = useState<Record<string, any>>(
		initialData.stream.type_schema?.properties || {},
	)
	const [selectedColumns, setSelectedColumns] = useState<string[]>(
		Object.keys(initialData.stream?.type_schema?.properties || {}),
	)
	const [isDisabled] = useState(true)

	useEffect(() => {
		if (initialData.stream.type_schema?.properties) {
			setColumnsToDisplay(initialData.stream.type_schema.properties)
			setSelectedColumns(Object.keys(initialData.stream.type_schema.properties))
		}
	}, [initialData])

	const handleSearch = useMemo(
		() => (query: string) => {
			if (!initialData.stream.type_schema?.properties) return
			const asArray = Object.entries(initialData.stream.type_schema.properties)
			const filtered = asArray.filter(([key]) =>
				key.toLowerCase().includes(query.toLowerCase()),
			)
			const filteredObject = Object.fromEntries(filtered)
			setColumnsToDisplay(filteredObject as Record<string, any>)
		},
		[initialData],
	)

	const handleSearchValueClear = useMemo(
		() => (event: React.ChangeEvent<HTMLInputElement>) => {
			if (
				event.target.value === "" &&
				initialData.stream.type_schema?.properties
			) {
				setTimeout(
					() =>
						setColumnsToDisplay(
							initialData?.stream?.type_schema?.properties || {},
						),
					0,
				)
			}
		},
		[initialData],
	)

	const handleSelectAll = useMemo(
		() => (e: CheckboxChangeEvent) => {
			if (!initialData.stream.type_schema?.properties) return
			const allColumns = Object.keys(initialData.stream.type_schema.properties)
			setSelectedColumns(e.target.checked ? allColumns : [])
			onColumnsChange?.(e.target.checked ? allColumns : [])
		},
		[initialData, onColumnsChange],
	)

	const handleColumnSelect = useMemo(
		() => (column: string, checked: boolean) => {
			const newSelectedColumns = checked
				? [...selectedColumns, column]
				: selectedColumns.filter(col => col !== column)
			setSelectedColumns(newSelectedColumns)
			onColumnsChange?.(newSelectedColumns)
		},
		[selectedColumns, onColumnsChange],
	)

	const isAllSelected = useMemo(
		() =>
			initialData.stream.type_schema?.properties
				? Object.keys(columnsToDisplay).length === selectedColumns.length
				: false,
		[initialData, columnsToDisplay, selectedColumns],
	)

	return (
		<div className="rounded-xl border border-[#E3E3E3] bg-white p-4">
			<div className="mb-3">
				<Input.Search
					className="custom-search-input w-full"
					placeholder="Search Streams"
					allowClear
					onSearch={handleSearch}
					onChange={handleSearchValueClear}
				/>
			</div>
			<div className="max-h-[400px] overflow-auto rounded border border-[#d9d9d9]">
				<div className="flex items-center border-b border-gray-400 p-3 last:border-b-0 hover:bg-background-primary">
					<Checkbox
						checked={isAllSelected}
						onChange={handleSelectAll}
						className="font-medium"
						disabled={isDisabled}
					>
						Select all
					</Checkbox>
				</div>
				{Object.keys(columnsToDisplay || {}).map(item => (
					<div
						key={item}
						className="flex items-center justify-between border-b border-gray-400 p-3 last:border-b-0 hover:bg-background-primary"
					>
						<div className="flex items-center gap-2">
							<Checkbox
								checked={selectedColumns.includes(item)}
								onChange={e => handleColumnSelect(item, e.target.checked)}
								disabled={isDisabled}
							/>
							<Tooltip title={item}>
								<span className="truncate font-medium">{item}</span>
							</Tooltip>
						</div>
						<RenderTypeItems
							initialList={initialData.stream.type_schema?.properties}
							item={item}
						/>
					</div>
				))}
			</div>
		</div>
	)
}

export default StreamsSchema
