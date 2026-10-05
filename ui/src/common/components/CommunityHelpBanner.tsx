import { ChatCircleTextIcon, WarningIcon, XIcon } from "@phosphor-icons/react"
import clsx from "clsx"
import { useState } from "react"

import {
	COMMUNITY_BANNER_DISMISSED_SESSION_KEY,
	OLAKE_COMMUNITY_SLACK_URL,
} from "@/common/constants"
import { AnalyticsEvent, trackEvent } from "@/core/analytics"

export const CommunityHelpBanner: React.FC<{
	source: "jobs" | "job_history"
	className?: string
}> = ({ source, className }) => {
	// Dismissal is per browser session, shared across pages.
	const [dismissed, setDismissed] = useState(
		() =>
			sessionStorage.getItem(COMMUNITY_BANNER_DISMISSED_SESSION_KEY) === "true",
	)

	if (dismissed) return null

	const handleClick = () => {
		trackEvent(AnalyticsEvent.CommunityHelpClicked, { source })
	}

	const handleDismiss = () => {
		sessionStorage.setItem(COMMUNITY_BANNER_DISMISSED_SESSION_KEY, "true")
		setDismissed(true)
	}

	return (
		<div
			className={clsx(
				"mb-6 flex items-center justify-between gap-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3",
				className,
			)}
		>
			<div className="flex items-center gap-3">
				<WarningIcon
					size={20}
					weight="fill"
					className="shrink-0 text-olake-error"
				/>
				<div>
					<p className="text-sm font-semibold text-olake-text">
						Experiencing sync failures?
					</p>
					<p className="text-xs text-olake-text-secondary">
						Stuck? The OLake community usually replies in under 30 minutes.
					</p>
				</div>
			</div>
			<div className="flex shrink-0 items-center gap-2">
				<a
					href={OLAKE_COMMUNITY_SLACK_URL}
					target="_blank"
					rel="noopener noreferrer"
					onClick={handleClick}
					className="flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-olake-text hover:bg-gray-50"
				>
					<ChatCircleTextIcon size={16} />
					Ask the community
				</a>
				<button
					type="button"
					aria-label="Dismiss"
					onClick={handleDismiss}
					className="rounded-md p-1.5 text-olake-text-secondary hover:bg-red-100 hover:text-olake-text"
				>
					<XIcon size={16} />
				</button>
			</div>
		</div>
	)
}
