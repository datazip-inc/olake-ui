import {
	ListBulletsIcon,
	MagnifyingGlassIcon,
	WarningCircleIcon,
} from "@phosphor-icons/react"
import { Input } from "antd"
import clsx from "clsx"
import { useMemo, useState } from "react"

import type { RunLogSource } from "../types"

type SubtaskListProps = {
	sources: RunLogSource[]
	selectedSourceKey: string
	onSelectSource: (sourceKey: string) => void
}

// Spark subtask log sources with a search box; only compaction runs have these.
const SubtaskList: React.FC<SubtaskListProps> = ({
	sources,
	selectedSourceKey,
	onSelectSource,
}) => {
	const [searchTerm, setSearchTerm] = useState("")

	const filtered = useMemo(() => {
		const normalized = searchTerm.trim().toLowerCase()
		if (!normalized) return sources
		return sources.filter(source =>
			source.label.toLowerCase().includes(normalized),
		)
	}, [searchTerm, sources])

	return (
		<>
			<div className="flex h-8 flex-1 overflow-hidden rounded-md border border-olake-border">
				<Input
					value={searchTerm}
					onChange={e => setSearchTerm(e.target.value)}
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

			<div className="mt-5 space-y-3">
				<div className="flex h-7 w-full items-center gap-[9px] px-2">
					<ListBulletsIcon
						size={16}
						className="text-olake-text-secondary"
					/>
					<span className="font-sans text-sm font-normal leading-[22px] text-olake-text-secondary">
						Subtasks
					</span>
				</div>
				<div className="flex w-full min-w-0 gap-4 pl-2">
					<div className="w-px shrink-0 self-stretch bg-olake-border" />
					<div className="min-w-0 flex-1 space-y-3">
						{filtered.length === 0 ? (
							<p className="px-2 font-sans text-sm font-normal leading-[22px] text-olake-text-tertiary">
								{sources.length === 0
									? "No subtask logs are available for this run."
									: "No subtasks match your search criteria."}
							</p>
						) : (
							filtered.map(source => (
								<button
									key={source.key}
									type="button"
									onClick={() => onSelectSource(source.key)}
									className={clsx(
										"flex h-6 w-full items-center gap-2 rounded-md px-2",
										selectedSourceKey === source.key
											? "bg-olake-surface-subtle"
											: "bg-transparent",
									)}
								>
									{source.hasError && (
										<WarningCircleIcon
											size={14}
											weight="fill"
											className="text-olake-error"
										/>
									)}
									<span
										className={clsx(
											"font-sans text-sm font-normal leading-[22px]",
											source.hasError
												? "text-olake-error"
												: "text-olake-text-secondary",
										)}
									>
										{source.label}
									</span>
								</button>
							))
						)}
					</div>
				</div>
			</div>
		</>
	)
}

export default SubtaskList
