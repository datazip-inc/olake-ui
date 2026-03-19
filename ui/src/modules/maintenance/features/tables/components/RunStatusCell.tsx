import { Popover } from "antd"

import { compactionSlots, runStatusConfig } from "../constants"
import type { Table } from "../types"
import CompactionPopoverContent from "./CompactionPopoverContent"

const RunStatusCell: React.FC<{ row: Table }> = ({ row }) => {
	const hasAnyRun = row.minor || row.major || row.full

	if (!hasAnyRun) {
		return (
			<div className="inline-flex h-5 items-center rounded-[20px] bg-olake-surface-muted px-2">
				<span className="text-xs font-medium leading-5 text-olake-text-secondary">
					Not Optimised
				</span>
			</div>
		)
	}

	const tags = (
		<div className="flex items-center justify-center gap-1">
			{compactionSlots.map(slot => {
				const run = { minor: row.minor, major: row.major, full: row.full }[
					slot.key
				]
				if (!run) return null
				const cfg = runStatusConfig[run.status]
				return (
					<div
						key={slot.key}
						className={`inline-flex items-center gap-1 rounded-[20px] px-2 py-px ${cfg.bgClass}`}
					>
						<cfg.Icon
							size={16}
							className={`${cfg.textClass} ${cfg.iconClass ?? ""}`}
						/>
						<span className={`text-xs font-medium leading-5 ${cfg.textClass}`}>
							{slot.tag}
						</span>
					</div>
				)
			})}
		</div>
	)

	return (
		<Popover
			content={
				<CompactionPopoverContent
					minor={row.minor}
					major={row.major}
					full={row.full}
				/>
			}
			trigger="hover"
			placement="bottomLeft"
			arrow={false}
			styles={{
				body: {
					padding: 0,
					borderRadius: 8,
					boxShadow:
						"0px 6px 16px 0px rgba(0,0,0,0.08), 0px 3px 6px -4px rgba(0,0,0,0.12), 0px 9px 28px 8px rgba(0,0,0,0.05)",
				},
			}}
		>
			{tags}
		</Popover>
	)
}

export default RunStatusCell
