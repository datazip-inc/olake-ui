import { MagnifyingGlassIcon } from "@phosphor-icons/react"
import { Input } from "antd"

import { TABLE_FILTER_OPTIONS } from "../constants"
import type { FilterKey } from "../types"
import FilterChips from "./FilterChips"

type Props = {
	searchTerm: string
	onSearchChange: (value: string) => void
	activeFilter: FilterKey
	onFilterChange: (filter: FilterKey) => void
}

const TableFilterBar: React.FC<Props> = ({
	searchTerm,
	onSearchChange,
	activeFilter,
	onFilterChange,
}) => (
	<div className="flex items-center gap-6">
		<div className="flex h-8 w-[479px] overflow-hidden rounded-md border border-olake-border">
			<Input
				value={searchTerm}
				onChange={e => onSearchChange(e.target.value)}
				placeholder="Search Tables"
				className="h-8 border-0"
			/>
			<button
				type="button"
				className="flex h-8 w-8 items-center justify-center border-l border-olake-border"
			>
				<MagnifyingGlassIcon size={16} />
			</button>
		</div>

		<FilterChips
			options={TABLE_FILTER_OPTIONS}
			value={activeFilter}
			onChange={onFilterChange}
		/>
	</div>
)

export default TableFilterBar
