import { SlidersIcon } from "@phosphor-icons/react"

import { moduleRegistry } from "@/core/modules/registry"

export type NavItem = {
	path: string
	label: string
	icon: React.ElementType
}

export type NavModule = {
	/** Unique key used for open/close state tracking */
	key: string
	/** Section header this module belongs to (e.g. "Services") */
	section: string
	/** Label shown in the breadcrumb and sidebar toggle button */
	moduleLabel: string
	icon: React.ElementType
	iconClassName?: string
	/** Optional badge text shown next to the module label (e.g. "New") */
	badge?: string
	items: NavItem[]
	/**
	 * Optional per-module breadcrumb trail resolver.
	 * Return an array of label segments (e.g. ["Tables", "Run Logs <foo>"]) for
	 * deep/dynamic routes this module owns, or null to fall through to the default.
	 */
	getBreadcrumbTrail?: (pathname: string) => string[] | null
}

// Boundary-safe route prefix matcher: exact match or `path/` prefix
export const matchesPath = (pathname: string, path: string) => {
	const a = pathname.toLowerCase()
	const b = path.toLowerCase()
	return a === b || a.startsWith(b + "/")
}

export const getNavModules = (enabledFeatures: Set<string>): NavModule[] =>
	moduleRegistry
		.filter(m => !m.gate || enabledFeatures.has(m.nav.key))
		.map(m => m.nav)

export const SYSTEM_ITEMS: NavItem[] = [
	{ path: "/settings", label: "Settings", icon: SlidersIcon },
]

// ─── Breadcrumb utils (fully driven by navModules — no manual edits needed) ──

export const getBreadcrumbModuleLabel = (
	pathname: string,
	modules: NavModule[],
): string => {
	const mod = modules.find(m =>
		m.items.some(item => matchesPath(pathname, item.path)),
	)
	return mod?.moduleLabel ?? "System"
}

/** Returns breadcrumb segments after the module label, e.g. ["Tables", "Run Logs <foo>"] */
export const getBreadcrumbTrail = (
	pathname: string,
	modules: NavModule[],
): string[] => {
	for (const mod of modules) {
		const trail = mod.getBreadcrumbTrail?.(pathname)
		if (trail) return trail
		const item = mod.items.find(i => matchesPath(pathname, i.path))
		if (item) return [item.label]
	}
	const sysItem = SYSTEM_ITEMS.find(
		i => i.path && matchesPath(pathname, i.path),
	)
	return sysItem ? [sysItem.label] : []
}
