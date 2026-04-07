import type { ComponentType } from "react"
import type { RouteObject } from "react-router-dom"

import type { NavModule } from "@/core/layout/nav-config"

/** App module descriptor (drives router routes + sidebar/topbar nav). */
export interface AppModule {
	// Nav configuration for this module (sidebar/topbar).
	nav: NavModule

	// React Router v6 routes belonging to this module.
	routes: RouteObject[]

	// Optional route gate; must render <Outlet /> so route children show.
	gate?: ComponentType

	// Hook that returns whether this module is currently enabled (for nav visibility).
	// If omitted, defaults to true.
	useEnabled?: () => boolean
}
