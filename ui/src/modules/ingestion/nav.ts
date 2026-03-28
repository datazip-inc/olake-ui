import {
	ArrowsMergeIcon,
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
} from "@phosphor-icons/react"

import type { NavModule } from "@/core/layout/nav-config"

interface BreadcrumbRoute {
	pattern: RegExp
	crumbs: (match: RegExpMatchArray) => string[]
}

// TODO: Make breadcrumbs links clickable for navigation
const breadcrumbRoutes: BreadcrumbRoute[] = [
	{
		pattern: /^\/jobs\/([^/]+)\/history\/([^/]+)\/logs$/,
		crumbs: () => ["Jobs", "History", "Logs"],
	},
	{
		pattern: /^\/jobs\/([^/]+)\/history$/,
		crumbs: () => ["Jobs", "History"],
	},
	{
		pattern: /^\/jobs\/([^/]+)\/settings$/,
		crumbs: () => ["Jobs", "Settings"],
	},
	{
		pattern: /^\/jobs\/([^/]+)\/edit$/,
		crumbs: () => ["Jobs", "Edit Job"],
	},
	{
		pattern: /^\/jobs\/new$/,
		crumbs: () => ["Jobs", "Create Job"],
	},
	{
		pattern: /^\/sources\/new$/,
		crumbs: () => ["Sources", "Create Source"],
	},
	{
		pattern: /^\/sources\/([^/]+)$/,
		crumbs: () => ["Sources", "Edit Source"],
	},
	{
		pattern: /^\/destinations\/new$/,
		crumbs: () => ["Destinations", "Create Destination"],
	},
	{
		pattern: /^\/destinations\/([^/]+)$/,
		crumbs: () => ["Destinations", "Edit Destination"],
	},
]

const matchBreadcrumbs = (pathname: string): string[] | null => {
	for (const route of breadcrumbRoutes) {
		const match = pathname.match(route.pattern)
		if (match) return route.crumbs(match)
	}
	return null
}

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
	getBreadcrumbTrail: matchBreadcrumbs,
}
