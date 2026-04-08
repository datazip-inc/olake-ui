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
			className="h-[102px] w-full rounded-lg border border-[#efefef] bg-olake-surface p-3 text-left hover:bg-neutral-50"
		>
			<div className="mb-2 flex items-center justify-between">
				<div className="flex items-center gap-2">
					<InfoIcon
						size={14}
						weight="fill"
						className="text-olake-primary"
					/>
					<span className="text-[12px] font-medium text-olake-primary">
						New Update
					</span>
				</div>
				<div className="relative">
					{hasNewUpdates && !hasSeenUpdates && (
						<span className="absolute -right-1 -top-1 h-2 w-2 rounded-full bg-olake-error" />
					)}
					<ArrowsOutSimpleIcon
						size={14}
						className="text-gray-900"
					/>
				</div>
			</div>
			<p className="text-xs leading-4 text-gray-900">
				{hasNewUpdates
					? `You have ${newUpdatesCount} new update${newUpdatesCount !== 1 ? "s" : ""}.`
					: "You're all up to date!"}
			</p>
		</button>
	)
}

export default UpdateNotification
