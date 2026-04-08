import type { AppModule } from "@/core/modules/types"

import { ingestionNavModule } from "./nav"
import { ingestionRoutes } from "./route"

/**
 * Ingestion module descriptor.
 * No gate — routes are accessible to any authenticated user.
 */
export const ingestionModule: AppModule = {
	nav: ingestionNavModule,
	routes: ingestionRoutes,
}
