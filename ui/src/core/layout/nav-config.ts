import { useActionAccess } from "@/common/components/action"
import { moduleRegistry } from "@/core/modules/registry"

export type NavItem = {
	path: string
	label: string
	icon: React.ElementType
	access?: string
}

export type NavModule = {
	/** Unique key used for open/close state tracking */
	key: string
	/** Section header this module belongs to (e.g. "Services") */
	section: string
	/**
	 * Label for the collapsible module header and breadcrumb root.
	 * Omit to render items directly under the section (no module group).
	 */
	moduleLabel?: string
	icon?: React.ElementType
	iconClassName?: string
	/** Optional badge text shown next to the module label (e.g. "New") */
	badge?: string
	/** Optional action rendered in the module header row (right of the caret). */
	headerAction?: React.FC
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
	const pathA = pathname.toLowerCase()
	const pathB = path.toLowerCase()
	return pathA === pathB || pathA.startsWith(pathB + "/")
}

export const getNavModules = (enabledFeatures: Set<string>): NavModule[] =>
	moduleRegistry
		.filter(m => !m.gate || enabledFeatures.has(m.nav.key))
		.map(m => m.nav)

export const useVisibleNavModules = (
	enabledFeatures: Set<string>,
): NavModule[] => {
	const { canAccess } = useActionAccess()

	return getNavModules(enabledFeatures)
		.map(mod => ({
			...mod,
			items: mod.items.filter(item => !item.access || canAccess(item.access)),
		}))
		.filter(mod => mod.items.length > 0)
}

/** Returns the nav module that owns `pathname` (matched via any of its items). */
const findModuleForPath = (pathname: string, modules: NavModule[]) =>
	modules.find(m => m.items.some(item => matchesPath(pathname, item.path)))

// ─── Breadcrumb utils (fully driven by navModules — no manual edits needed) ──

export const getBreadcrumbModuleLabel = (
	pathname: string,
	modules: NavModule[],
): string => {
	const mod = findModuleForPath(pathname, modules)
	if (!mod) return ""

	if (mod.moduleLabel) return mod.moduleLabel

	const item = mod.items.find(i => matchesPath(pathname, i.path))
	return item?.label ?? mod.section
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
		if (!item) continue

		if (!mod.moduleLabel) return []
		return [item.label]
	}

	return []
}
