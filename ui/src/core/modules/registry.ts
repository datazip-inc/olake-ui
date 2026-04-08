/**
 * SINGLE SOURCE OF TRUTH for all application modules.
 *
 * To add a new module:
 *   1. Create src/modules/<name>/index.ts exporting an AppModule descriptor.
 *   2. Import it here and add it to this array.
 *   3. Add a corresponding hook call in useActiveModuleKeys (maintain order).
 *
 * Order determines:
 *   - The order routes are matched (first match wins in React Router v6).
 *   - The order modules appear in the nav sidebar.
 */
import { ingestionModule } from "@/modules/ingestion/index"
import { maintenanceModule } from "@/modules/maintenance/index"

export const moduleRegistry = [ingestionModule, maintenanceModule]

/**
 * Returns the set of nav keys for modules that are currently enabled.
 * Each module's useEnabled hook is called in a fixed static order to satisfy
 * React's hook rules (no conditional / loop calls).
 */
export function useActiveModuleKeys(): Set<string> {
	// One entry per module in moduleRegistry — order must match the array above.
	const flags = [
		ingestionModule.useEnabled?.() ?? true,
		maintenanceModule.useEnabled?.() ?? true,
	]
	return new Set(moduleRegistry.filter((_, i) => flags[i]).map(m => m.nav.key))
}
