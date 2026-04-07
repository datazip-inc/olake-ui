import type { AppModule } from "@/core/modules/types"
import { useOptimizationStatus } from "@/core/platform/hooks/useOptimizationStatus"

import { CompactionGate } from "./common/components"
import { maintenanceNavModule } from "./nav"
import { maintenanceRoutes } from "./route"

/** Maintenance module descriptor gated by compaction status. */
export const maintenanceModule: AppModule = {
	nav: maintenanceNavModule,
	routes: maintenanceRoutes,
	gate: CompactionGate,
	useEnabled: () => useOptimizationStatus().data?.enabled ?? false,
}
