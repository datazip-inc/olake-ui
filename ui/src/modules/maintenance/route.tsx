import { lazy } from "react"
import { RouteObject } from "react-router-dom"

const lazyComponents = {
	Tables: lazy(() => import("./features/tables/pages/Tables")),
	RunHistory: lazy(() => import("./features/tables/pages/RunHistory")),
	RunLogs: lazy(() => import("./features/tables/pages/RunLogs")),
	Catalogs: lazy(() => import("./features/catalogs/pages/Catalogs")),
}

export const maintenanceRoutes: RouteObject[] = [
	{
		path: "maintenance/tables",
		element: <lazyComponents.Tables />,
	},
	{
		path: "maintenance/tables/:catalog/:database/:tableName/runs",
		element: <lazyComponents.RunHistory />,
	},
	{
		path: "maintenance/tables/:catalog/:database/:tableName/runs/:runId/logs",
		element: <lazyComponents.RunLogs />,
	},
	{
		path: "maintenance/catalogs",
		element: <lazyComponents.Catalogs />,
	},
]
