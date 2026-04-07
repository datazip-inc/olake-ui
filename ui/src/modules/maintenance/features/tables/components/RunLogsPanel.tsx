import {
	ArrowsClockwiseIcon,
	DownloadSimpleIcon,
	MagnifyingGlassIcon,
} from "@phosphor-icons/react"
import { Button, Input, message } from "antd"
import clsx from "clsx"
import { useEffect, useMemo, useState } from "react"
import { Virtuoso } from "react-virtuoso"

import { getLogLevelClass, getLogTextColor } from "@/common/utils/utils"

import { DRIVER_SOURCE_KEY } from "../constants"
import { useDownloadProcessLogFile, useProcessLogs } from "../hooks"
import type { RunLogEntry } from "../types"
import { getProcessLogFileId } from "../utils"

type RunLogsPanelProps = {
	runId: string
	selectedSourceKey: string
}

const RunLogsPanel: React.FC<RunLogsPanelProps> = ({
	runId,
	selectedSourceKey,
}) => {
	const { data, isFetching, refetch } = useProcessLogs(runId)
	const downloadLogFile = useDownloadProcessLogFile()

	const logs = data?.logsBySource[selectedSourceKey] ?? []

	const title =
		selectedSourceKey === DRIVER_SOURCE_KEY
			? "Driver Logs"
			: `Subtask ${selectedSourceKey.replace("task-", "")}`

	const [searchTerm, setSearchTerm] = useState("")

	const filteredLogs = useMemo(() => {
		const normalizedSearch = searchTerm.trim().toLowerCase()
		if (!normalizedSearch) return logs

		return logs.filter(log =>
			`${log.date} ${log.time} ${log.level} ${log.message}`
				.toLowerCase()
				.includes(normalizedSearch),
		)
	}, [logs, searchTerm])

	useEffect(() => {
		setSearchTerm("")
	}, [selectedSourceKey])

	return (
		<div className="flex h-full min-h-0 flex-1 flex-col pt-5">
			<div className="px-4">
				<h2 className="font-sans text-base font-medium leading-7 text-olake-text">
					{title}
				</h2>
				<div className="mt-3 flex items-center gap-3">
					<div className="flex h-8 w-[364px] overflow-hidden rounded-md border border-olake-border">
						<Input
							value={searchTerm}
							onChange={e => setSearchTerm(e.target.value)}
							placeholder="Search Logs"
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
						onClick={() => refetch()}
					/>
					<Button
						icon={<DownloadSimpleIcon size={14} />}
						loading={downloadLogFile.isPending}
						disabled={logs.length === 0}
						onClick={() =>
							downloadLogFile.mutate(
								{
									processId: runId,
									fileId: getProcessLogFileId(selectedSourceKey),
								},
								{
									onSuccess: () => message.success("Downloading logs..."),
								},
							)
						}
					>
						Download Logs
					</Button>
				</div>
			</div>

			<div className="mt-5 min-h-0 flex-1 border-t border-olake-border">
				{filteredLogs.length === 0 ? (
					<div className="px-6 py-8 font-sans text-sm font-normal leading-5 text-olake-text-tertiary">
						{logs.length === 0
							? "No logs are available for this selection."
							: "No logs match your search criteria."}
					</div>
				) : (
					<Virtuoso
						key={selectedSourceKey}
						style={{ height: "100%" }}
						data={filteredLogs}
						itemContent={(_index, row) => <RunLogRow row={row} />}
					/>
				)}
			</div>
		</div>
	)
}

const RunLogRow: React.FC<{ row: RunLogEntry }> = ({ row }) => {
	const levelKey = row.level.toLowerCase()
	const normalizedStackTrace = row.stackTrace
		? row.stackTrace.replace(/\\n/g, "\n").replace(/\\t/g, "\t")
		: null

	return (
		<div className="grid grid-cols-[87px_92px_79px_minmax(0,1fr)] items-start border-b border-olake-border py-2 pl-[30px] pr-5">
			<span className="font-mono text-[10px] font-medium leading-[17px] text-olake-body">
				{row.date}
			</span>
			<span className="font-mono text-[10px] font-medium leading-[17px] text-olake-body">
				{row.time}
			</span>
			<span>
				<span
					className={clsx(
						"inline-flex h-5 items-center rounded-[20px] px-2 font-sans text-[10px] font-medium leading-5",
						getLogLevelClass(levelKey),
					)}
				>
					{row.level}
				</span>
			</span>
			<span
				className={clsx(
					"whitespace-normal break-words font-mono text-[10px] font-medium leading-[17px]",
					getLogTextColor(levelKey),
				)}
			>
				{row.message}
				{normalizedStackTrace && (
					<pre className="mt-1 overflow-x-auto whitespace-pre-wrap font-mono text-[10px] font-medium leading-[17px] text-olake-body">
						{normalizedStackTrace}
					</pre>
				)}
			</span>
		</div>
	)
}

export default RunLogsPanel
