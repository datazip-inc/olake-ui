import { ArrowSquareOutIcon, XIcon } from "@phosphor-icons/react"
import { Button, Drawer, Spin, Tooltip } from "antd"

import type { RunMetricRow } from "../types"

type RunMetricsSidebarProps = {
	open: boolean
	onClose: () => void
	rows: RunMetricRow[]
	loading?: boolean
	runId?: string
}

const RunMetricsSidebar: React.FC<RunMetricsSidebarProps> = ({
	open,
	onClose,
	rows,
	loading = false,
	runId,
}) => {
	return (
		<Drawer
			placement="right"
			open={open}
			onClose={onClose}
			destroyOnHidden
			width={707}
			closeIcon={null}
			styles={{
				body: {
					padding: 0,
				},
			}}
		>
			<div className="min-h-full bg-olake-surface">
				<div className="relative px-7 pb-5 pt-10">
					<Button
						type="text"
						icon={<XIcon size={24} />}
						onClick={onClose}
						className="absolute right-4 top-4 text-olake-text-tertiary"
					/>
					<div className="flex items-center gap-x-2 pr-12">
						<h2 className="font-sans text-xl font-medium leading-7 text-olake-text">
							Run Metrics for{" "}
							<span className="text-olake-primary">Run ID {runId ?? "--"}</span>
						</h2>
						<Tooltip title="Learn more">
							{/* TODO: Update link whenever available */}
							<a
								href="https://olake.io/docs/maintenance/run-metrics"
								className="text-olake-icon-muted"
								target="_blank"
								rel="noopener noreferrer"
								aria-label="Learn more"
							>
								<ArrowSquareOutIcon className="size-5" />
							</a>
						</Tooltip>
					</div>
				</div>

				{loading ? (
					<div className="flex h-40 items-center justify-center">
						<Spin />
					</div>
				) : rows.length === 0 ? (
					<div className="px-7 py-8 font-sans text-sm font-normal leading-5 text-olake-text-tertiary">
						No metrics found.
					</div>
				) : (
					<div className="flex flex-col">
						{rows.map(item => (
							<div
								key={item.label}
								className="flex min-h-16 items-center justify-between gap-4 border-b border-olake-border px-6 py-4"
							>
								<p className="shrink-0 whitespace-nowrap font-sans text-base font-normal leading-6 text-olake-text">
									{item.label}
								</p>
								<p className="break-words pl-8 text-right font-sans text-base font-medium leading-6 text-olake-text">
									{item.value}
								</p>
							</div>
						))}
					</div>
				)}
			</div>
		</Drawer>
	)
}

export default RunMetricsSidebar
