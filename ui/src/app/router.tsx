import { lazy } from "react"
import {
	createBrowserRouter,
	Navigate,
	Outlet,
	useLocation,
} from "react-router-dom"

import { ErrorBoundary } from "@/common/components"
import { useAuthStore } from "@/core/auth/stores/authStore"
import Layout from "@/core/layout"
import { buildProtectedChildren } from "@/core/modules/buildRoutes"
import { moduleRegistry } from "@/core/modules/registry"

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

const lazyComponents = {
	Login: lazy(() => import("@/core/auth/pages/Login")),
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
			...buildProtectedChildren(moduleRegistry),
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
