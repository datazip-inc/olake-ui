import { lazy } from "react"
import { RouteObject } from "react-router-dom"

import { ActionRoute } from "@/common/components/action"

const lazyComponents = {
	Jobs: lazy(() => import("./features/jobs/pages/Jobs")),
	JobHistory: lazy(() => import("./features/jobs/pages/JobHistory")),
	JobLogs: lazy(() => import("./features/jobs/pages/JobLogs")),
	JobSettings: lazy(() => import("./features/jobs/pages/JobSettings")),
	JobCreation: lazy(() => import("./features/jobs/pages/JobCreation")),
	JobEdit: lazy(() => import("./features/jobs/pages/JobEdit")),
	Sources: lazy(() => import("./features/sources/pages/Sources")),
	SourceEdit: lazy(() => import("./features/sources/pages/SourceEdit")),
	CreateSource: lazy(() => import("./features/sources/pages/CreateSource")),
	Destinations: lazy(
		() => import("./features/destinations/pages/Destinations"),
	),
	DestinationEdit: lazy(
		() => import("./features/destinations/pages/DestinationEdit"),
	),
	CreateDestination: lazy(
		() => import("./features/destinations/pages/CreateDestination"),
	),
}

export const ingestionRoutes: RouteObject[] = [
	{
		path: "jobs",
		element: <lazyComponents.Jobs />,
	},
	{
		path: "jobs/new",
		element: (
			<ActionRoute>
				<lazyComponents.JobCreation />
			</ActionRoute>
		),
	},
	{
		path: "jobs/:jobId/edit",
		element: (
			<ActionRoute>
				<lazyComponents.JobEdit />
			</ActionRoute>
		),
	},
	{
		path: "jobs/:jobId/history",
		element: <lazyComponents.JobHistory />,
	},
	{
		path: "jobs/:jobId/history/:historyId/logs",
		element: <lazyComponents.JobLogs />,
	},
	{
		path: "jobs/:jobId/tasks/:taskId/logs",
		element: <lazyComponents.JobLogs />,
	},
	{
		path: "jobs/:jobId/settings",
		element: (
			<ActionRoute>
				<lazyComponents.JobSettings />
			</ActionRoute>
		),
	},
	{
		path: "sources",
		element: <lazyComponents.Sources />,
	},
	{
		path: "sources/new",
		element: (
			<ActionRoute>
				<lazyComponents.CreateSource />
			</ActionRoute>
		),
	},
	{
		path: "sources/:sourceId",
		element: (
			<ActionRoute>
				<lazyComponents.SourceEdit />
			</ActionRoute>
		),
	},
	{
		path: "destinations",
		element: <lazyComponents.Destinations />,
	},
	{
		path: "destinations/new",
		element: (
			<ActionRoute>
				<lazyComponents.CreateDestination />
			</ActionRoute>
		),
	},
	{
		path: "destinations/:destinationId",
		element: (
			<ActionRoute>
				<lazyComponents.DestinationEdit />
			</ActionRoute>
		),
	},
]
