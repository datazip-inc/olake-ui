import type { AppModule } from "@/core/modules/types"

import { settingsNavModule } from "./nav"
import { settingsRoutes } from "./route"

export const settingsModule: AppModule = {
	nav: settingsNavModule,
	routes: settingsRoutes,
}
