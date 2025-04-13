import { Input, Tooltip } from "antd"
import RenderTypeItems from "../../../common/components/RenderTypeItems"
import { Checkbox } from "antd/es"
import { useState, useEffect } from "react"

const StreamsSchema = ({ initialData, onColumnsChange }: any) => {
	const [columnsToDisplay, setColumnsToDisplay] = useState(initialData)
	const [selectedColumns, setSelectedColumns] = useState<string[]>(
		Object.keys(initialData || {}),
	)

	useEffect(() => {
		setColumnsToDisplay(initialData)
		setSelectedColumns(Object.keys(initialData || {}))
	}, [initialData])

	const handleSearch = (query: string) => {
		const asArray = Object.entries(initialData)
		const filtered = asArray.filter(([key]) => key.includes(query))
		setColumnsToDisplay(Object.fromEntries(filtered))
	}

	const handleSearchValueClear = (
		event: React.ChangeEvent<HTMLInputElement>,
	) => {
		if (event.target.value === "")
			setTimeout(() => setColumnsToDisplay(initialData), 0)
	}

	const handleSelectAll = (checked: boolean) => {
		const allColumns = Object.keys(initialData || {})
		setSelectedColumns(checked ? allColumns : [])
		onColumnsChange?.(checked ? allColumns : [])
	}

	const handleColumnSelect = (column: string, checked: boolean) => {
		const newSelectedColumns = checked
			? [...selectedColumns, column]
			: selectedColumns.filter(col => col !== column)
		setSelectedColumns(newSelectedColumns)
		onColumnsChange?.(newSelectedColumns)
	}

	const isAllSelected =
		Object.keys(columnsToDisplay || {}).length === selectedColumns.length

	return (
		<>
			<div className={`flex flex-col items-center gap-2 p-5`}>
				<Input.Search
					className="w-full flex-auto"
					placeholder="Search Schema"
					onSearch={handleSearch}
					onChange={handleSearchValueClear}
				/>
				<div className="flex w-full items-center">
					<div className="flex w-full flex-col items-center justify-between truncate rounded-[6px] border border-solid border-[#d9d9d9]">
						<div className="flex w-full items-center border-b border-solid border-[#d9d9d9] px-6 py-4">
							<Checkbox
								checked={isAllSelected}
								onChange={e => handleSelectAll(e.target.checked)}
								className="font-medium"
							>
								Select all
							</Checkbox>
						</div>
						<div className="max-h-[400px] w-full overflow-auto">
							{Object.keys(columnsToDisplay)?.map(item => (
								<div
									key={item}
									className={`flex w-full items-center justify-between truncate border border-l-0 border-r-0 border-t-0 border-solid border-[#d9d9d9] px-6 py-4 last:border-b-0`}
								>
									<div className="flex items-center gap-4 overflow-hidden">
										<Checkbox
											checked={selectedColumns.includes(item)}
											onChange={e => handleColumnSelect(item, e.target.checked)}
											className="ml-auto select-none rounded-lg"
										>
											<Tooltip title={item}>
												<span className="truncate font-medium">{item}</span>
											</Tooltip>
										</Checkbox>
									</div>
									<RenderTypeItems
										initialList={initialData}
										item={item}
									/>
								</div>
							))}
						</div>
					</div>
				</div>
			</div>
		</>
	)
}

export default StreamsSchema
