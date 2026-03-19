import {
	ArrowsMergeIcon,
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
} from "@phosphor-icons/react"

import type { NavModule } from "@/core/layout/nav-config"

export const ingestionNavModule: NavModule = {
	key: "ingestion",
	section: "Services",
	moduleLabel: "Ingestion",
	icon: ArrowsMergeIcon,
	iconClassName: "-rotate-90",
	items: [
		{ path: "/jobs", label: "Jobs", icon: GitCommitIcon },
		{ path: "/sources", label: "Sources", icon: PathIcon },
		{ path: "/destinations", label: "Destinations", icon: LinktreeLogoIcon },
	],
}
