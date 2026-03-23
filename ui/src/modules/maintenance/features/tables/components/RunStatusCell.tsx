import { CircleDashedIcon } from "@phosphor-icons/react"
import { Popover } from "antd"
import clsx from "clsx"

import { Tag } from "@/common/components"

import { compactionSlots, runStatusConfig } from "../constants"
import type { Table } from "../types"
import CompactionPopoverContent from "./CompactionPopoverContent"

const RunStatusCell: React.FC<{ row: Table }> = ({ row }) => {
	const runs = { minor: row.minor, major: row.major, full: row.full }
	const hasAnyRun = row.minor || row.major || row.full

	if (!hasAnyRun) {
		return (
			<Tag
				color="muted"
				className="text-xs"
			>
				Not Optimized
			</Tag>
		)
	}

	const tags = (
		<div className="flex items-center justify-center gap-1">
			{compactionSlots.map(slot => {
				const run = runs[slot.key] ?? null
				if (!run) {
					return (
						<div
							key={slot.key}
							className="inline-flex items-center gap-1 rounded-[20px] bg-olake-surface-muted px-2 py-px"
							aria-label={`${slot.name} — no run`}
						>
							<CircleDashedIcon
								size={16}
								weight="regular"
								className="text-olake-text-disabled"
							/>
							<span className="text-xs font-medium leading-5 text-olake-text-tertiary">
								{slot.tag}
							</span>
						</div>
					)
				}

				const cfg = runStatusConfig[run.status]
				if (!cfg) {
					return (
						<div
							key={slot.key}
							className="inline-flex items-center gap-1 rounded-[20px] bg-olake-surface-muted px-2 py-px"
							aria-label={`${slot.name} — unknown status`}
						>
							<CircleDashedIcon
								size={16}
								weight="regular"
								className="text-olake-text-disabled"
							/>
							<span className="text-xs font-medium leading-5 text-olake-text-tertiary">
								{slot.tag}
							</span>
						</div>
					)
				}

				return (
					<div
						key={slot.key}
						className={`inline-flex items-center gap-1 rounded-[20px] px-2 py-px ${cfg.bgClass}`}
						aria-label={`${slot.name} — ${cfg.label}`}
					>
						<cfg.Icon
							size={16}
							className={clsx(cfg.textClass, cfg.iconClass)}
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
