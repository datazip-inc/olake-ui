import clsx from "clsx"
import { Tooltip } from "antd"
import {
	SlidersHorizontalIcon,
	ColumnsPlusRightIcon,
	GridFourIcon,
	InfoIcon,
} from "@phosphor-icons/react"

import { useStreamSelectionStore } from "../../stores"
import { selectActiveStreamData } from "../../stores/streamSelectionStore"
import { formatDestinationPath } from "../../utils"
import { DESTINATION_TABLE_TOOLTIP_TEXT, TAB_STYLES } from "../../constants"

interface StreamConfigHeaderProps {
	activeTab: string
	onTabChange: (tab: string) => void
}

const TabButton = ({
	id,
	label,
	icon,
	activeTab,
	onTabChange,
}: {
	id: string
	label: string
	icon: React.ReactNode
	activeTab: string
	onTabChange: (tab: string) => void
}) => {
	const tabStyle =
		activeTab === id
			? TAB_STYLES.active
			: `${TAB_STYLES.inactive} ${TAB_STYLES.hover}`

	return (
		<button
			className={clsx(
				tabStyle,
				"flex items-center justify-center gap-1 text-xs",
			)}
			style={{ fontWeight: 500, height: "28px", width: "100%" }}
			onClick={() => onTabChange(id)}
			type="button"
		>
			<span className="flex items-center">{icon}</span>
			{label}
		</button>
	)
}

const StreamConfigHeader = ({
	activeTab,
	onTabChange,
}: StreamConfigHeaderProps) => {
	const stream = useStreamSelectionStore(selectActiveStreamData)

	if (!stream) return null

	const destinationPath = formatDestinationPath(
		stream.stream.destination_database,
		stream.stream.destination_table,
	)

	return (
		<>
			<div className="flex items-center justify-between gap-4 pb-4 font-medium">
				<span>{stream.stream.name}</span>
				{destinationPath && (
					<div className="min-w-0 flex-shrink">
						<div className="max-w-full rounded-lg bg-background-primary px-3 py-1">
							<div className="flex min-w-0 items-center text-sm">
								<div className="flex items-center whitespace-nowrap font-medium">
									Destination Table{" "}
									<Tooltip title={DESTINATION_TABLE_TOOLTIP_TEXT}>
										<InfoIcon className="size-5 cursor-help items-center px-0.5 text-gray-500" />
									</Tooltip>{" "}
									:
								</div>
								<Tooltip
									title={`${destinationPath}`}
									placement="top"
								>
									<span className="min-w-0 flex-1 truncate pl-1 font-normal">
										{destinationPath}
									</span>
								</Tooltip>
							</div>
						</div>
					</div>
				)}
			</div>
			<div className="mb-4 w-full">
				<div className="grid grid-cols-3 gap-1 rounded-md bg-background-primary p-1">
					<TabButton
						id="config"
						label="Config"
						icon={<SlidersHorizontalIcon className="size-3.5" />}
						activeTab={activeTab}
						onTabChange={onTabChange}
					/>
					<TabButton
						id="schema"
						label="Schema"
						icon={<ColumnsPlusRightIcon className="size-3.5" />}
						activeTab={activeTab}
						onTabChange={onTabChange}
					/>
					<TabButton
						id="partitioning"
						label="Partitioning"
						icon={<GridFourIcon className="size-3.5" />}
						activeTab={activeTab}
						onTabChange={onTabChange}
					/>
				</div>
			</div>
		</>
	)
}

export default StreamConfigHeader
