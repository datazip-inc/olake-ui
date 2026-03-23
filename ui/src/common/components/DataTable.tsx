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

	/** The expected number of rows per page, used to fix the table's minimum height. Defaults to 6. */
	pageSize?: number

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
	const getVisiblePages = () => {
		if (totalPages <= 7)
			return Array.from({ length: totalPages }, (_, i) => i + 1)

		if (currentPage <= 4) {
			return [1, 2, 3, 4, 5, "...", totalPages]
		}

		if (currentPage >= totalPages - 3) {
			return [
				1,
				"...",
				totalPages - 4,
				totalPages - 3,
				totalPages - 2,
				totalPages - 1,
				totalPages,
			]
		}

		return [
			1,
			"...",
			currentPage - 1,
			currentPage,
			currentPage + 1,
			"...",
			totalPages,
		]
	}

	const visiblePages = getVisiblePages()

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

			{visiblePages.map((page, index) => {
				const isEllipsis = page === "..."
				return (
					<button
						key={index}
						type="button"
						className={`h-6 min-w-[24px] rounded-md border text-sm leading-5 ${
							isEllipsis
								? "border-transparent bg-transparent text-olake-body-secondary"
								: currentPage === page
									? "border-olake-border bg-olake-surface-muted text-olake-body-secondary"
									: "border-olake-border bg-white text-olake-body-secondary hover:bg-gray-50"
						}`}
						onClick={() =>
							!isEllipsis ? onPageChange(page as number) : undefined
						}
						disabled={isEllipsis}
					>
						{page}
					</button>
				)
			})}

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
	loadingRowCount = 10,
	emptyState,
	pagination,
	pageSize = 10,
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

	// Calculate fixed height: 48px header + (pageSize * 56px per row) + 2px borders
	const tableMinHeight = 48 + pageSize * 56 + 2

	return (
		<div className="flex flex-col">
			<div style={{ minHeight: `${tableMinHeight}px` }}>
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
					{!loading &&
						rows.length === 0 &&
						(emptyState ?? <DefaultEmptyState />)}
				</div>
			</div>

			{/* Pagination */}
			{pagination && (
				<div className="mt-4 flex justify-end">
					<TablePagination {...pagination} />
				</div>
			)}
		</div>
	)
}

export default DataTable
