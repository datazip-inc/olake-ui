import clsx from "clsx"

import { compactionSlots } from "../constants"
import { RUN_STATUS } from "../constants"
import type { CompactionRun } from "../types"
import { getRunStatusConfig } from "../utils"

type Props = {
	minor: CompactionRun
	major: CompactionRun
	full: CompactionRun
}

const CompactionPopoverContent: React.FC<Props> = ({ minor, major, full }) => {
	const runs = { minor, major, full }

	return (
		<div className="w-60">
			{compactionSlots.map((slot, idx) => {
				const run = runs[slot.key]
				const cfg = run ? getRunStatusConfig(run.status) : null

				return (
					<div key={slot.key}>
						{idx > 0 && <div className="border-t border-olake-border" />}
						<div className="flex items-start justify-between px-5 py-4">
							<div className="flex flex-col gap-0.5">
								<p className="text-xs font-medium leading-4 text-olake-text">
									{slot.name}
								</p>
								{run?.status === RUN_STATUS.RUNNING ? (
									<p className="text-[10px] leading-normal text-olake-text-tertiary">
										running
									</p>
								) : run?.lastRun ? (
									<p className="text-[10px] leading-normal text-olake-text-tertiary">
										last run {run.lastRun}
									</p>
								) : (
									<p className="text-[10px] leading-normal text-olake-text-tertiary">
										never run
									</p>
								)}
							</div>
							{cfg && (
								<div className="flex items-center gap-1">
									<cfg.Icon
										size={12}
										className={clsx(cfg.textClass, cfg.iconClass)}
									/>
									<span
										className={clsx(
											"text-xs font-medium leading-5",
											cfg.textClass,
										)}
									>
										{cfg.label}
									</span>
								</div>
							)}
						</div>
					</div>
				)
			})}
		</div>
	)
}

export default CompactionPopoverContent
