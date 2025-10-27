import { lazy } from "react"
import { createBrowserRouter, Navigate, Outlet } from "react-router-dom"

import { useAppStore } from "../store"
import Layout from "../modules/common/components/Layout"
import { ErrorBoundary } from "../modules/common/components/ErrorBoundary"

// eslint-disable-next-line react-refresh/only-export-components
const RootHandler = () => {
	const isAuthenticated = useAppStore(state => state.isAuthenticated)

	if (isAuthenticated) {
		return (
			<Layout>
				<Outlet />
			</Layout>
		)
	} else {
		return (
			<Navigate
				to="/login"
				replace
			/>
		)
	}
}

//lazy load components
const lazyComponents = {
	Login: lazy(() => import("../modules/auth/pages/Login")),
	Jobs: lazy(() => import("../modules/jobs/pages/Jobs")),
	JobHistory: lazy(() => import("../modules/jobs/pages/JobHistory")),
	JobLogs: lazy(() => import("../modules/jobs/pages/JobLogs")),
	JobSettings: lazy(() => import("../modules/jobs/pages/JobSettings")),
	JobCreation: lazy(() => import("../modules/jobs/pages/JobCreation")),
	JobEdit: lazy(() => import("../modules/jobs/pages/JobEdit")),
	Sources: lazy(() => import("../modules/sources/pages/Sources")),
	SourceEdit: lazy(() => import("../modules/sources/pages/SourceEdit")),
	CreateSource: lazy(() => import("../modules/sources/pages/CreateSource")),
	Destinations: lazy(
		() => import("../modules/destinations/pages/Destinations"),
	),
	DestinationEdit: lazy(
		() => import("../modules/destinations/pages/DestinationEdit"),
	),
	CreateDestination: lazy(
		() => import("../modules/destinations/pages/CreateDestination"),
	),
}

const publicRoutes = [
	{
		path: "/login",
		element: <lazyComponents.Login />,
		errorElement: (
			<ErrorBoundary>
				<lazyComponents.Login />
			</ErrorBoundary>
		),
	},
	{
		path: "*",
		element: (
			<Navigate
				to="/login"
				replace
			/>
		),
	},
]

//these are the protected routes which are only accessible after login
const protectedRoutes = [
	{
		path: "/",
		element: <RootHandler />,
		errorElement: (
			<ErrorBoundary>
				<RootHandler />
			</ErrorBoundary>
		),
		children: [
			{
				index: true,
				element: (
					<Navigate
						to="/jobs"
						replace
					/>
				),
			},
			{
				path: "jobs",
				element: <lazyComponents.Jobs />,
			},
			{
				path: "jobs/new",
				element: <lazyComponents.JobCreation />,
			},
			{
				path: "jobs/:jobId/edit",
				element: <lazyComponents.JobEdit />,
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
				element: <lazyComponents.JobSettings />,
			},
			{
				path: "sources",
				element: <lazyComponents.Sources />,
			},
			{
				path: "sources/new",
				element: <lazyComponents.CreateSource />,
			},
			{
				path: "sources/:sourceId",
				element: <lazyComponents.SourceEdit />,
			},
			{
				path: "destinations",
				element: <lazyComponents.Destinations />,
			},
			{
				path: "destinations/new",
				element: <lazyComponents.CreateDestination />,
			},
			{
				path: "destinations/:destinationId",
				element: <lazyComponents.DestinationEdit />,
			},
			{
				path: "*",
				element: (
					<Navigate
						to="/jobs"
						replace
					/>
				),
			},
		],
	},
]

export const router = createBrowserRouter([...publicRoutes, ...protectedRoutes])

export { publicRoutes, protectedRoutes }
