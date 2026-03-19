import { SlidersIcon } from "@phosphor-icons/react"

import { ingestionNavModule } from "@/modules/ingestion/nav"
import { maintenanceNavModule } from "@/modules/maintenance/nav"

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

// ─── Module registry ──────────────────────────────────────────────────────────
// To add a new module: create a nav.ts in the module folder and add one import
// + one spread entry below. Nothing else in the layout needs to change.

export const NAV_MODULES: NavModule[] = [
	ingestionNavModule,
	maintenanceNavModule,
]

export const SYSTEM_ITEMS: NavItem[] = [
	{ path: "/settings", label: "Settings", icon: SlidersIcon },
]

/** Unique section names, preserved in declaration order */
export const NAV_SECTIONS = [...new Set(NAV_MODULES.map(m => m.section))]

// ─── Breadcrumb utils (fully driven by NAV_MODULES — no manual edits needed) ──

export const getBreadcrumbModuleLabel = (pathname: string): string => {
	const mod = NAV_MODULES.find(m =>
		m.items.some(item => pathname.startsWith(item.path)),
	)
	return mod?.moduleLabel ?? "System"
}

/** Returns breadcrumb segments after the module label, e.g. ["Tables", "Run Logs <foo>"] */
export const getBreadcrumbTrail = (pathname: string): string[] => {
	for (const mod of NAV_MODULES) {
		const trail = mod.getBreadcrumbTrail?.(pathname)
		if (trail) return trail
		const item = mod.items.find(i => pathname.startsWith(i.path))
		if (item) return [item.label]
	}
	const sysItem = SYSTEM_ITEMS.find(i => i.path && pathname.startsWith(i.path))
	return sysItem ? [sysItem.label] : []
}
