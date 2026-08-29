import { lazy } from "react"
import type { RouteObject } from "react-router-dom"

const SystemSettings = lazy(() => import("./pages/SystemSettings"))

export const settingsRoutes: RouteObject[] = [
	{
		path: "settings",
		element: <SystemSettings />,
	},
]
