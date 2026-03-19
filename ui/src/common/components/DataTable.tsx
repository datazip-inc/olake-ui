import type React from "react"

export type ColumnAlignment = "left" | "center" | "right"

export type ColumnDef<TRow> = {
	key: string
	header: string | React.ReactNode
	/** Percentage of total table width (0–100). Columns without width get 1fr. */
	width?: number
	/** Horizontal alignment for both header and cell content. Defaults to left. */
	align?: ColumnAlignment
	render: (row: TRow) => React.ReactNode
}

export type PaginationConfig = {
	currentPage: number
	totalPages: number
	onPageChange: (page: number) => void
}

export type DataTableProps<TRow> = {
	columns: ColumnDef<TRow>[]
	rows: TRow[]
	rowKey: (row: TRow) => string

	loading?: boolean
	/** Number of skeleton rows to show while loading. Defaults to 6. */
	loadingRowCount?: number

	/** Shown when rows are empty and not loading. Defaults to built-in "No Data". */
	emptyState?: React.ReactNode

	/** Omit to disable pagination. */
	pagination?: PaginationConfig

	className?: string
}

const DefaultEmptyState: React.FC = () => (
	<div className="flex h-24 items-center justify-center px-6 text-sm text-olake-text-tertiary">
		No Data
	</div>
)

const TablePagination: React.FC<PaginationConfig> = ({
	currentPage,
	totalPages,
	onPageChange,
}) => {
	const pageNumbers = Array.from({ length: totalPages }, (_, i) => i + 1)

	return (
		<div className="flex items-center gap-2 py-3 text-sm leading-5 text-olake-body-secondary">
			<button
				type="button"
				className="flex h-6 items-center gap-1 rounded-md border border-olake-border px-2 disabled:opacity-40"
				onClick={() => onPageChange(Math.max(1, currentPage - 1))}
				disabled={currentPage === 1}
			>
				<span>&lt;</span>
				<span>Previous</span>
			</button>

			{pageNumbers.map(page => (
				<button
					key={page}
					type="button"
					className={`h-6 w-6 rounded-md border border-olake-border text-sm leading-5 ${
						currentPage === page
							? "bg-olake-surface-muted text-olake-body-secondary"
							: "bg-white text-olake-body-secondary"
					}`}
					onClick={() => onPageChange(page)}
				>
					{page}
				</button>
			))}

			<button
				type="button"
				className="flex h-6 items-center gap-1 rounded-md border border-olake-border px-2 disabled:opacity-40"
				onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}
				disabled={currentPage === totalPages}
			>
				<span>Next</span>
				<span>&gt;</span>
			</button>
		</div>
	)
}

function DataTable<TRow>({
	columns,
	rows,
	rowKey,
	loading = false,
	loadingRowCount = 6,
	emptyState,
	pagination,
	className,
}: DataTableProps<TRow>) {
	const getAlignmentClass = (align: ColumnAlignment = "left") => {
		if (align === "center") return "text-center"
		if (align === "right") return "text-right"
		return "text-left"
	}

	const gridTemplateColumns = columns
		.map(col => (col.width !== undefined ? `${col.width}%` : "1fr"))
		.join(" ")

	const gridStyle: React.CSSProperties = { gridTemplateColumns }

	return (
		<>
			<div
				className={`overflow-hidden rounded-lg border border-olake-border ${className ?? ""}`}
			>
				{/* Header */}
				<div
					className="grid h-12 items-center gap-4 bg-olake-surface-subtle px-6 text-xs font-medium leading-5 text-olake-text-secondary"
					style={gridStyle}
				>
					{columns.map(col => (
						<div
							key={col.key}
							className={getAlignmentClass(col.align)}
						>
							{col.header}
						</div>
					))}
				</div>

				{/* Loading skeleton */}
				{loading &&
					Array.from({ length: loadingRowCount }).map((_, idx) => (
						<div
							key={idx} // skeleton rows have no data identity; index is acceptable
							className="h-14 border-t border-olake-border bg-white"
						/>
					))}

				{/* Data rows */}
				{!loading &&
					rows.map(row => (
						<div
							key={rowKey(row)}
							className="grid h-14 items-center gap-4 border-t border-olake-border px-6 text-sm leading-[22px] text-olake-text"
							style={gridStyle}
						>
							{columns.map(col => (
								<div
									key={col.key}
									className={getAlignmentClass(col.align)}
								>
									{col.render(row)}
								</div>
							))}
						</div>
					))}

				{/* Empty state */}
				{!loading && rows.length === 0 && (emptyState ?? <DefaultEmptyState />)}
			</div>

			{/* Pagination */}
			{pagination && <TablePagination {...pagination} />}
		</>
	)
}

export default DataTable
