import { ArrowsOutSimpleIcon, InfoIcon } from "@phosphor-icons/react"

import { usePlatformStore } from "@/core/platform/stores"
import { ReleaseMetadataResponse } from "@/core/platform/types"

const UpdateNotification: React.FC<{ onOpen: () => void }> = ({ onOpen }) => {
	const { releases, isLoadingReleases, hasSeenUpdates, setHasSeenUpdates } =
		usePlatformStore()

	const newUpdatesCount = releases
		? Object.values(releases).reduce((total, category) => {
				const count =
					category?.releases.filter((release: ReleaseMetadataResponse) =>
						release.tags.includes("New Release"),
					).length || 0
				return total + count
			}, 0)
		: 0

	const hasNewUpdates = newUpdatesCount > 0

	if (isLoadingReleases) return null

	const handleOpenModal = () => {
		onOpen()
		setHasSeenUpdates(true)
	}

	return (
		<button
			onClick={handleOpenModal}
			className="h-[102px] w-full rounded-lg border border-[#efefef] bg-white p-3 text-left hover:bg-neutral-50"
		>
			<div className="mb-[8px] flex items-center justify-between">
				<div className="flex items-center gap-2">
					<InfoIcon
						size={14}
						weight="fill"
						color="#193AE6"
					/>
					<span className="text-[12px] font-medium text-[#193AE6]">
						New Update
					</span>
				</div>
				<div className="relative">
					{hasNewUpdates && !hasSeenUpdates && (
						<span className="absolute -right-1 -top-1 h-2 w-2 rounded-full bg-red-500" />
					)}
					<ArrowsOutSimpleIcon
						size={14}
						color="#383838"
					/>
				</div>
			</div>
			<p className="text-[12px] leading-4 text-[#383838]">
				{hasNewUpdates
					? `You have ${newUpdatesCount} new update${newUpdatesCount !== 1 ? "s" : ""}.`
					: "You're all up to date!"}
			</p>
		</button>
	)
}

export default UpdateNotification
