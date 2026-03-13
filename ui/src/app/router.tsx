import { lazy } from "react"
import {
	createBrowserRouter,
	Navigate,
	Outlet,
	useLocation,
} from "react-router-dom"

import { useAuthStore } from "@/core/auth/stores/authStore"
import Layout from "@/core/layout"
import { ErrorBoundary } from "@/modules/ingestion/common/components/ErrorBoundary"

// eslint-disable-next-line react-refresh/only-export-components
const RootHandler = () => {
	const isAuthenticated = useAuthStore(state => state.isAuthenticated)
	const location = useLocation()

	if (isAuthenticated) {
		return (
			<Layout>
				<ErrorBoundary key={location.pathname}>
					<Outlet />
				</ErrorBoundary>
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
	Login: lazy(() => import("@/core/auth/pages/Login")),
	Jobs: lazy(() => import("../modules/ingestion/features/jobs/pages/Jobs")),
	JobHistory: lazy(
		() => import("../modules/ingestion/features/jobs/pages/JobHistory"),
	),
	JobLogs: lazy(
		() => import("../modules/ingestion/features/jobs/pages/JobLogs"),
	),
	JobSettings: lazy(
		() => import("../modules/ingestion/features/jobs/pages/JobSettings"),
	),
	JobCreation: lazy(
		() => import("../modules/ingestion/features/jobs/pages/JobCreation"),
	),
	JobEdit: lazy(
		() => import("../modules/ingestion/features/jobs/pages/JobEdit"),
	),
	Sources: lazy(
		() => import("../modules/ingestion/features/sources/pages/Sources"),
	),
	SourceEdit: lazy(
		() => import("../modules/ingestion/features/sources/pages/SourceEdit"),
	),
	CreateSource: lazy(
		() => import("../modules/ingestion/features/sources/pages/CreateSource"),
	),
	Destinations: lazy(
		() =>
			import("../modules/ingestion/features/destinations/pages/Destinations"),
	),
	DestinationEdit: lazy(
		() =>
			import(
				"../modules/ingestion/features/destinations/pages/DestinationEdit"
			),
	),
	CreateDestination: lazy(
		() =>
			import(
				"../modules/ingestion/features/destinations/pages/CreateDestination"
			),
	),
	SystemSettings: lazy(() => import("@/core/settings/pages/SystemSettings")),
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
				path: "settings",
				element: <lazyComponents.SystemSettings />,
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
