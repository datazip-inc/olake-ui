import {
	CaretLeftIcon,
	DownloadSimpleIcon,
	HardDrivesIcon,
	ListBulletsIcon,
	MagnifyingGlassIcon,
	WarningCircleIcon,
} from "@phosphor-icons/react"
import { Button, Input } from "antd"
import { useMemo, useState } from "react"

import { DRIVER_SOURCE_KEY } from "../constants"
import { useProcessLogs } from "../hooks"

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

	const taskSources = data?.taskSources ?? []

	const [subtaskSearchTerm, setSubtaskSearchTerm] = useState("")

	const filteredSubtasks = useMemo(() => {
		const normalized = subtaskSearchTerm.trim().toLowerCase()
		if (!normalized) return taskSources
		return taskSources.filter(source =>
			source.label.toLowerCase().includes(normalized),
		)
	}, [subtaskSearchTerm, taskSources])

	return (
		<div className="h-full w-[360px] overflow-y-auto border-r border-olake-border px-4 pb-6 pt-6">
			<button
				type="button"
				onClick={onBack}
				className="mb-2 inline-flex items-center gap-1 font-sans text-sm font-normal leading-[22px] text-olake-text-secondary"
			>
				<CaretLeftIcon size={12} />
				<span>{`Run Logs <${tableName}>`}</span>
			</button>

			<h1 className="font-sans text-xl font-medium leading-7 text-olake-text">
				Run logs for <span className="text-olake-primary">Run ID {runId}</span>
			</h1>
			<p className="mt-1 font-sans text-sm font-normal leading-[22px] text-olake-text">
				View spark driver &amp; subtask logs
			</p>

			<div className="mt-4 flex items-center gap-2">
				<div className="flex h-8 flex-1 overflow-hidden rounded-md border border-olake-border">
					<Input
						value={subtaskSearchTerm}
						onChange={e => setSubtaskSearchTerm(e.target.value)}
						placeholder="Search Subtasks"
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
					size="small"
					icon={<DownloadSimpleIcon size={14} />}
				>
					Logs
				</Button>
			</div>

			<div className="mt-5 space-y-4">
				<div className="space-y-3">
					<button
						type="button"
						onClick={() => onSelectSource(DRIVER_SOURCE_KEY)}
						className={`flex h-7 w-full items-center gap-[9px] rounded-md px-2 ${
							selectedSourceKey === DRIVER_SOURCE_KEY
								? "bg-olake-surface-muted"
								: "bg-transparent"
						}`}
					>
						<HardDrivesIcon
							size={16}
							className="text-olake-text-secondary"
						/>
						<span className="font-sans text-sm font-normal leading-[22px] text-olake-text-secondary">
							Driver Logs
						</span>
					</button>
					<div className="flex h-7 w-full items-center gap-[9px] px-2">
						<ListBulletsIcon
							size={16}
							className="text-olake-text-secondary"
						/>
						<span className="font-sans text-sm font-normal leading-[22px] text-olake-text-secondary">
							Subtasks
						</span>
					</div>
				</div>

				<div className="flex gap-4 pl-2">
					<div className="w-px bg-olake-border" />
					<div className="space-y-3">
						{filteredSubtasks.map(source => {
							const isSelected = selectedSourceKey === source.key

							return (
								<button
									key={source.key}
									type="button"
									onClick={() => onSelectSource(source.key)}
									className={`flex h-6 items-center gap-2 rounded-md px-2 ${
										isSelected ? "bg-olake-surface-subtle" : "bg-transparent"
									}`}
								>
									{source.hasError && (
										<WarningCircleIcon
											size={14}
											weight="fill"
											className="text-olake-error"
										/>
									)}
									<span
										className={`font-sans text-sm font-normal leading-[22px] ${
											source.hasError
												? "text-olake-error"
												: "text-olake-text-secondary"
										}`}
									>
										{source.label}
									</span>
								</button>
							)
						})}
					</div>
				</div>
			</div>
		</div>
	)
}

export default RunLogSidebar
