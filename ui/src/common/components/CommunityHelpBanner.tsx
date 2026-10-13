import {
	ChatCircleTextIcon,
	InfoIcon,
	WarningIcon,
	XIcon,
} from "@phosphor-icons/react"
import clsx from "clsx"
import { useState } from "react"

import {
	COMMUNITY_BANNER_DISMISSED_SESSION_KEY,
	OLAKE_COMMUNITY_SLACK_URL,
} from "@/common/constants"
import { AnalyticsEvent, trackEvent } from "@/core/analytics"

// "help" is a neutral nudge shown while syncs are fine (including right after
// creating a job); "sync_failure" is shown when the most recent sync failed.
const BANNER_VARIANTS = {
	help: {
		title: "Need help? Contact the OLake community",
		description: "The OLake community usually replies in under 30 minutes.",
		Icon: InfoIcon,
		containerClassName: "border-primary-100 bg-primary-50",
		iconClassName: "text-olake-primary",
		dismissClassName: "hover:bg-primary-100",
	},
	sync_failure: {
		title: "Experiencing sync failures?",
		description:
			"Stuck? The OLake community usually replies in under 30 minutes.",
		Icon: WarningIcon,
		containerClassName: "border-red-200 bg-red-50",
		iconClassName: "text-olake-error",
		dismissClassName: "hover:bg-red-100",
	},
}

export const CommunityHelpBanner: React.FC<{
	source: "jobs" | "job_history"
	variant: keyof typeof BANNER_VARIANTS
	className?: string
}> = ({ source, variant: variantKey, className }) => {
	// Dismissal is per browser session and per variant, shared across pages, so
	// dismissing the help nudge does not hide a later sync-failure banner.
	const storageKey = `${COMMUNITY_BANNER_DISMISSED_SESSION_KEY}_${variantKey}`
	const [dismissedVariants, setDismissedVariants] = useState<string[]>([])
	const dismissed =
		dismissedVariants.includes(variantKey) ||
		sessionStorage.getItem(storageKey) === "true"

	if (dismissed) return null

	const { title, description, Icon, ...variant } = BANNER_VARIANTS[variantKey]

	const handleClick = () => {
		trackEvent(AnalyticsEvent.CommunityHelpClicked, { source })
	}

	const handleDismiss = () => {
		sessionStorage.setItem(storageKey, "true")
		setDismissedVariants(prev => [...prev, variantKey])
	}

	return (
		<div
			className={clsx(
				"mb-6 flex items-center justify-between gap-4 rounded-lg border px-4 py-3",
				variant.containerClassName,
				className,
			)}
		>
			<div className="flex items-center gap-3">
				<Icon
					size={20}
					weight="fill"
					className={clsx("shrink-0", variant.iconClassName)}
				/>
				<div>
					<p className="text-sm font-semibold text-olake-text">{title}</p>
					<p className="text-xs text-olake-text-secondary">{description}</p>
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
					className={clsx(
						"rounded-md p-1.5 text-olake-text-secondary hover:text-olake-text",
						variant.dismissClassName,
					)}
				>
					<XIcon size={16} />
				</button>
			</div>
		</div>
	)
}
