import { MagnifyingGlassIcon } from "@phosphor-icons/react"
import { Input } from "antd"
import clsx from "clsx"

import type { FilterKey } from "../types"

const FILTER_OPTIONS: Array<{ key: FilterKey; label: string }> = [
	{ key: "all", label: "All tables" },
	{ key: "olake", label: "OLake Ingested" },
	{ key: "external", label: "Imported Tables" },
]

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
		<div className="flex h-9 w-[479px] overflow-hidden rounded-md border border-olake-border">
			<Input
				value={searchTerm}
				onChange={e => onSearchChange(e.target.value)}
				placeholder="Search Tables"
				className="h-9 border-0"
			/>
			<button
				type="button"
				className="flex h-9 w-8 items-center justify-center border-l border-olake-border"
			>
				<MagnifyingGlassIcon size={16} />
			</button>
		</div>

		<div className="flex items-center gap-2">
			{FILTER_OPTIONS.map(filter => {
				const active = activeFilter === filter.key
				return (
					<button
						key={filter.key}
						type="button"
						onClick={() => onFilterChange(filter.key)}
						className={clsx(
							"h-9 whitespace-nowrap rounded-md border border-olake-border px-3 text-sm leading-5",
							active
								? "bg-olake-surface-muted text-olake-text-secondary"
								: "bg-white text-olake-body-secondary",
						)}
					>
						{filter.label}
					</button>
				)
			})}
		</div>
	</div>
)

export default TableFilterBar
