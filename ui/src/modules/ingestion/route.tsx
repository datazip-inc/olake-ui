import { lazy } from "react"
import { RouteObject } from "react-router-dom"

import { Action } from "@/common/components/action"

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
			<Action>
				<lazyComponents.JobCreation />
			</Action>
		),
	},
	{
		path: "jobs/:jobId/edit",
		element: (
			<Action>
				<lazyComponents.JobEdit />
			</Action>
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
			<Action>
				<lazyComponents.JobSettings />
			</Action>
		),
	},
	{
		path: "sources",
		element: <lazyComponents.Sources />,
	},
	{
		path: "sources/new",
		element: (
			<Action>
				<lazyComponents.CreateSource />
			</Action>
		),
	},
	{
		path: "sources/:sourceId",
		element: (
			<Action>
				<lazyComponents.SourceEdit />
			</Action>
		),
	},
	{
		path: "destinations",
		element: <lazyComponents.Destinations />,
	},
	{
		path: "destinations/new",
		element: (
			<Action>
				<lazyComponents.CreateDestination />
			</Action>
		),
	},
	{
		path: "destinations/:destinationId",
		element: (
			<Action>
				<lazyComponents.DestinationEdit />
			</Action>
		),
	},
]
