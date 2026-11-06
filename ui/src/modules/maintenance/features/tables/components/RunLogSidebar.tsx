import {
	ArrowLeftIcon,
	DownloadSimpleIcon,
	HardDrivesIcon,
} from "@phosphor-icons/react"
import { Button, message } from "antd"
import clsx from "clsx"

import { DRIVER_SOURCE_KEY } from "../constants"
import { useDownloadProcessLogsArchive, useProcessLogs } from "../hooks"
import SubtaskList from "./SubtaskList"

type RunLogSidebarProps = {
	tableName: string
	runId: string
	selectedSourceKey: string
	onSelectSource: (sourceKey: string) => void
	onBack: () => void
}

const RunLogSidebar: React.FC<RunLogSidebarProps> = ({
	tableName,
	runId,
	selectedSourceKey,
	onSelectSource,
	onBack,
}) => {
	const { data } = useProcessLogs(runId)
	const downloadArchive = useDownloadProcessLogsArchive()
	const hasAnyLogs = Object.values(data?.logsBySource ?? {}).some(
		logEntries => logEntries.length > 0,
	)
	// Expiry runs are one AMS job with a single log; compaction runs add Spark subtasks.
	const isExpiryRun = data?.isExpiryRun ?? false

	return (
		<div className="h-full w-[360px] overflow-y-auto border-r border-olake-border px-4 pb-6 pt-6">
			<button
				type="button"
				onClick={onBack}
				className="mb-2 inline-flex items-center gap-1 font-sans text-sm font-normal leading-[22px] text-olake-text-secondary"
			>
				<ArrowLeftIcon size={14} />
				<span>{`Run Logs / ${tableName} `}</span>
			</button>

			<h1 className="font-sans text-xl font-medium leading-7 text-olake-text">
				Run logs for <span className="text-olake-primary">Run ID {runId}</span>
			</h1>
			<p className="mt-1 font-sans text-sm font-normal leading-[22px] text-olake-text">
				{isExpiryRun
					? "View snapshot expiry logs"
					: "View spark driver & subtask logs"}
			</p>

			<div className="mt-4 flex items-center gap-2">
				<Button
					size="small"
					className="!h-8"
					icon={<DownloadSimpleIcon size={14} />}
					loading={downloadArchive.isPending}
					disabled={!hasAnyLogs}
					onClick={() =>
						downloadArchive.mutate(runId, {
							onSuccess: () => message.success("Downloading logs..."),
						})
					}
				>
					Logs
				</Button>
			</div>

			<div className="mt-5 space-y-3">
				<button
					type="button"
					onClick={() => onSelectSource(DRIVER_SOURCE_KEY)}
					className={clsx(
						"flex h-7 w-full items-center gap-[9px] rounded-md px-2",
						selectedSourceKey === DRIVER_SOURCE_KEY
							? "bg-olake-surface-muted"
							: "bg-transparent",
					)}
				>
					<HardDrivesIcon
						size={16}
						className="text-olake-text-secondary"
					/>
					<span className="font-sans text-sm font-normal leading-[22px] text-olake-text-secondary">
						{isExpiryRun ? "Expiry Logs" : "Driver Logs"}
					</span>
				</button>

				{!isExpiryRun && (
					<SubtaskList
						sources={data?.taskSources ?? []}
						selectedSourceKey={selectedSourceKey}
						onSelectSource={onSelectSource}
					/>
				)}
			</div>
		</div>
	)
}

export default RunLogSidebar
