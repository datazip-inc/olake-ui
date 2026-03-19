import {
	ArrowsInCardinalIcon,
	FolderIcon,
	TableIcon,
} from "@phosphor-icons/react"

import type { NavModule } from "@/core/layout/nav-config"

interface BreadcrumbRoute {
	pattern: RegExp
	crumbs: (match: RegExpMatchArray) => string[]
}

const breadcrumbRoutes: BreadcrumbRoute[] = [
	{
		pattern:
			/^\/maintenance\/tables\/([^/]+)\/([^/]+)\/([^/]+)\/runs\/([^/]+)\/logs$/,
		crumbs: match => {
			const tableName = decodeURIComponent(match[3] ?? "")
			const runId = decodeURIComponent(match[4] ?? "")
			return ["Tables", `Run Logs <${tableName}>`, `Logs: Run ID ${runId}`]
		},
	},
	{
		pattern: /^\/maintenance\/tables\/([^/]+)\/([^/]+)\/([^/]+)\/runs$/,
		crumbs: match => {
			const tableName = decodeURIComponent(match[3] ?? "")
			return ["Tables", `Run Logs <${tableName}>`]
		},
	},
]

const matchBreadcrumbs = (pathname: string): string[] | null => {
	for (const route of breadcrumbRoutes) {
		const match = pathname.match(route.pattern)
		if (match) return route.crumbs(match)
	}
	return null
}

export const maintenanceNavModule: NavModule = {
	key: "maintenance",
	section: "Services",
	moduleLabel: "Maintenance",
	icon: ArrowsInCardinalIcon,
	badge: "New",
	items: [
		{ path: "/maintenance/tables", label: "Tables", icon: TableIcon },
		{ path: "/maintenance/catalogs", label: "Catalogs", icon: FolderIcon },
	],
	getBreadcrumbTrail: matchBreadcrumbs,
}
