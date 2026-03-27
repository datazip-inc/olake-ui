import {
	ArrowsClockwiseIcon,
	CaretLeftIcon,
	MagnifyingGlassIcon,
} from "@phosphor-icons/react"
import { Button, Input } from "antd"
import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"

import { DataTable } from "@/common/components"
import type { ColumnDef } from "@/common/components"
import { usePaginatedSearch } from "@/common/hooks"

import { RunMetricsSidebar } from "../components"
import { runLogsStatusConfig } from "../constants"
import { useTableRuns } from "../hooks"
import type { RunMetricRow, TableRun } from "../types"

type RunsFilter = "all" | "failed"

const getColumns = (
	handleMetricsClick: (row: TableRun) => void,
	handleLogsClick: (runId: string) => void,
): ColumnDef<TableRun>[] => [
	{
		key: "runId",
		header: "Run ID",
		width: 12,
		render: row => row.runId,
	},
	{
		key: "status",
		header: "Status",
		align: "center",
		width: 16,
		render: row => {
			const cfg = runLogsStatusConfig[row.status]
			return (
				<span
					className={`inline-flex h-5 items-center gap-1 rounded-[20px] px-2 ${cfg.bgClass}`}
				>
					<cfg.Icon
						size={12}
						className={`${cfg.textClass}`}
					/>
					<span className={`text-xs font-medium leading-5 ${cfg.textClass}`}>
						{cfg.label}
					</span>
				</span>
			)
		},
	},
	{
		key: "type",
		header: "Type",
		width: 10,
		render: row => row.type,
	},
	{
		key: "startTime",
		header: "Start Time",
		width: 14,
		render: row => row.startTime,
	},
	{
		key: "duration",
		header: "Duration",
		width: 12,
		render: row => row.duration,
	},
	{
		key: "metrics",
		header: "Metrics",
		width: 10,
		align: "center",
		render: row => (
			<Button
				size="small"
				onClick={() => handleMetricsClick(row)}
			>
				View
			</Button>
		),
	},
	{
		key: "logs",
		header: "Logs",
		width: 10,
		align: "center",
		render: row => (
			<Button
				size="small"
				onClick={() => handleLogsClick(row.runId)}
			>
				View
			</Button>
		),
	},
]

const RunHistory: React.FC = () => {
	const navigate = useNavigate()
	const { catalog, database, tableName } = useParams<{
		catalog: string
		database: string
		tableName: string
	}>()
	const decodedTableName = decodeURIComponent(tableName ?? "")
	const decodedCatalog = decodeURIComponent(catalog ?? "")
	const decodedDatabase = decodeURIComponent(database ?? "")

	const [metricsSidebarOpen, setMetricsSidebarOpen] = useState(false)
	const [metricsRows, setMetricsRows] = useState<RunMetricRow[]>([])
	const [selectedRunId, setSelectedRunId] = useState<string>("")
	const {
		data: runs = [],
		isLoading,
		isFetching,
		refetch,
	} = useTableRuns(decodedCatalog, decodedDatabase, decodedTableName)

	const {
		searchTerm,
		setSearchTerm,
		activeFilter,
		setActiveFilter,
		currentPage,
		setCurrentPage,
		filteredRows,
		paginatedRows,
		totalPages,
	} = usePaginatedSearch<TableRun, RunsFilter>({
		rows: runs,
		initialFilter: "all",
		searchFn: (run, term) => run.runId.toLowerCase().includes(term),
		filterFn: (run, filter) =>
			filter === "all" ? true : run.status === "FAILED",
	})

	const handleMetricsClick = (row: TableRun) => {
		setSelectedRunId(row.runId)
		setMetricsSidebarOpen(true)
		setMetricsRows(row.metrics)
	}
	const getRunLogsPath = (runId: string) =>
		`/maintenance/tables/${encodeURIComponent(decodedCatalog)}/${encodeURIComponent(decodedDatabase)}/${encodeURIComponent(decodedTableName)}/runs/${encodeURIComponent(runId)}/logs`

	const columns = getColumns(handleMetricsClick, runId =>
		navigate(getRunLogsPath(runId)),
	)

	return (
		<div className="min-h-full bg-white px-6 pt-6">
			<button
				type="button"
				onClick={() => {
					if (!decodedCatalog || !decodedDatabase) {
						navigate("/maintenance/tables")
						return
					}

					navigate(
						`/maintenance/tables?catalog=${encodeURIComponent(decodedCatalog)}&database=${encodeURIComponent(decodedDatabase)}`,
					)
				}}
				className="mb-2 inline-flex items-center gap-1.5 text-sm leading-[22px] text-olake-text-secondary"
			>
				<CaretLeftIcon size={12} />
				<span>Back</span>
			</button>

			<h1 className="text-[28px] font-medium leading-[36px] text-olake-text">
				Run History for{" "}
				<span className="text-olake-primary">{decodedTableName}</span>
			</h1>
			<p className="mt-1 text-sm leading-[22px] text-olake-text-secondary">
				View logs, runs &amp; performance metrics
			</p>

			<div className="mt-6 flex items-center gap-4">
				<div className="flex h-8 w-72 overflow-hidden rounded-md border border-olake-border">
					<Input
						value={searchTerm}
						onChange={e => setSearchTerm(e.target.value)}
						placeholder="Search Run ID"
						className="h-8 border-0 text-sm"
					/>
					<button
						type="button"
						className="flex h-8 w-8 items-center justify-center border-l border-olake-border"
					>
						<MagnifyingGlassIcon size={14} />
					</button>
				</div>
				<Button
					icon={<ArrowsClockwiseIcon size={14} />}
					loading={isFetching}
					onClick={() => {
						refetch()
					}}
				/>

				<div className="flex items-center gap-1.5">
					{[
						{ key: "all", label: "All Runs" },
						{ key: "failed", label: "Failed Runs" },
					].map(filter => {
						const active = activeFilter === (filter.key as RunsFilter)
						return (
							<button
								key={filter.key}
								type="button"
								onClick={() => setActiveFilter(filter.key as RunsFilter)}
								className={`h-8 rounded-md border border-olake-border px-3 text-sm leading-[22px] ${
									active
										? "bg-olake-surface-muted text-olake-text-secondary"
										: "bg-white text-olake-text-secondary"
								}`}
							>
								{filter.label}
							</button>
						)
					})}
				</div>
			</div>

			<div className="mt-4">
				<DataTable
					columns={columns}
					rows={paginatedRows}
					rowKey={row => row.id}
					loading={isLoading || isFetching}
					pagination={{
						currentPage,
						totalPages,
						onPageChange: setCurrentPage,
					}}
					pageSize={10}
					emptyState={
						filteredRows.length === 0
							? undefined
							: "No runs on this page. Try another page."
					}
				/>
			</div>

			<RunMetricsSidebar
				open={metricsSidebarOpen}
				onClose={() => setMetricsSidebarOpen(false)}
				rows={metricsRows}
				loading={false}
				runId={selectedRunId}
			/>
		</div>
	)
}

export default RunHistory
