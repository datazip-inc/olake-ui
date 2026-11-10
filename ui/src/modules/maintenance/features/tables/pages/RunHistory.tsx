import { ArrowsClockwiseIcon, CaretLeftIcon } from "@phosphor-icons/react"
import { Button } from "antd"
import clsx from "clsx"
import { useState } from "react"
import { useNavigate, useParams } from "react-router"

import { DataTable, TruncatedText } from "@/common/components"
import type { ColumnDef } from "@/common/components"

import { FilterChips, RunMetricsSidebar } from "../components"
import {
	RUN_LOGS_TYPE_QUERY_PARAM,
	RUN_STATUS_FILTER_OPTIONS,
	RUN_TYPE_FILTER_OPTIONS,
} from "../constants"
import { useRunHistoryFilters, useTableRuns } from "../hooks"
import type { RunMetricRow, TableRun } from "../types"
import { getRunLogsStatusConfig } from "../utils"

const getColumns = (
	handleMetricsClick: (row: TableRun) => void,
	handleLogsClick: (row: TableRun) => void,
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
			const cfg = getRunLogsStatusConfig(row.status)
			return (
				<span
					className={clsx(
						"inline-flex h-5 items-center gap-1 rounded-[20px] px-2",
						cfg.bgClass,
					)}
				>
					<cfg.Icon
						size={12}
						className={clsx(cfg.textClass)}
					/>
					<span
						className={clsx("text-xs font-medium leading-5", cfg.textClass)}
					>
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
				onClick={() => handleLogsClick(row)}
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
		activeFilter,
		activeTypeFilter,
		currentPage,
		setCurrentPage,
		pageSize,
		status,
		runType,
		excludeRunType,
		handleFilterChange,
		handleTypeFilterChange,
	} = useRunHistoryFilters()
	const { data, isLoading, isFetching, refetch } = useTableRuns({
		catalog: decodedCatalog,
		database: decodedDatabase,
		tableName: decodedTableName,
		page: currentPage,
		pageSize,
		status,
		runType,
		excludeRunType,
	})

	const runs = data?.runs ?? []
	const totalRuns = data?.total ?? 0

	const totalPages = Math.max(1, Math.ceil(totalRuns / pageSize))

	const handleMetricsClick = (row: TableRun) => {
		setSelectedRunId(row.runId)
		setMetricsSidebarOpen(true)
		setMetricsRows(row.metrics)
	}
	// The run kind rides along so the logs page can tell an expiry run from a compaction run.
	const getRunLogsPath = (run: TableRun) =>
		`/maintenance/tables/${encodeURIComponent(decodedCatalog)}/${encodeURIComponent(decodedDatabase)}/${encodeURIComponent(decodedTableName)}/runs/${encodeURIComponent(run.runId)}/logs?${RUN_LOGS_TYPE_QUERY_PARAM}=${encodeURIComponent(run.runType)}`

	const columns = getColumns(handleMetricsClick, run =>
		navigate(getRunLogsPath(run)),
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

			<h1 className="flex min-w-0 items-baseline gap-2 text-[28px] font-medium leading-[36px] text-olake-text">
				<span className="shrink-0">Run History for</span>
				<TruncatedText
					text={decodedTableName}
					className="text-olake-primary"
				/>
			</h1>
			<p className="mt-1 text-sm leading-[22px] text-olake-text-secondary">
				View logs, runs &amp; performance metrics
			</p>

			<div className="mt-6 flex items-center gap-4">
				<Button
					icon={<ArrowsClockwiseIcon size={14} />}
					loading={isFetching}
					onClick={() => {
						refetch()
					}}
				/>

				<FilterChips
					options={RUN_STATUS_FILTER_OPTIONS}
					value={activeFilter}
					onChange={handleFilterChange}
				/>

				{/* Run type: compaction (Lite/Medium/Full) vs snapshot expiry */}
				<div className="border-l border-olake-border pl-4">
					<FilterChips
						options={RUN_TYPE_FILTER_OPTIONS}
						value={activeTypeFilter}
						onChange={handleTypeFilterChange}
					/>
				</div>
			</div>

			<div className="mt-4">
				<DataTable
					columns={columns}
					rows={runs}
					rowKey={row => row.id}
					loading={isLoading || isFetching}
					pagination={{
						currentPage,
						totalPages,
						onPageChange: setCurrentPage,
					}}
					pageSize={pageSize}
					emptyStateConfig={{
						title: "No runs found",
						subtitle:
							"No runs are available for the selected filter. Try switching filters or refresh.",
						onRefetch: () => void refetch(),
						refetchLabel: "Refresh",
					}}
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
