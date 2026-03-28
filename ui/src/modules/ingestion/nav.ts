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
		crumbs: match => ["Jobs", match[1], "History", "Logs"],
	},
	{
		pattern: /^\/jobs\/([^/]+)\/history$/,
		crumbs: match => ["Jobs", match[1], "History"],
	},
	{
		pattern: /^\/jobs\/([^/]+)\/settings$/,
		crumbs: match => ["Jobs", match[1], "Settings"],
	},
	{
		pattern: /^\/jobs\/([^/]+)\/edit$/,
		crumbs: match => ["Jobs", match[1], "Edit Job"],
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
		crumbs: match => ["Sources", match[1], "Edit Source"],
	},
	{
		pattern: /^\/destinations\/new$/,
		crumbs: () => ["Destinations", "Create Destination"],
	},
	{
		pattern: /^\/destinations\/([^/]+)$/,
		crumbs: match => ["Destinations", match[1], "Edit Destination"],
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
