import { lazy, Suspense, useEffect } from "react"
import {
	BrowserRouter as Router,
	Routes,
	Route,
	Navigate,
} from "react-router-dom"
import { ConfigProvider, App as AntApp, Spin } from "antd"
import Layout from "./modules/common/components/Layout"
import { ErrorBoundary } from "./modules/common/components/ErrorBoundary"
import { useAppStore } from "./store"

// Lazy load components
const Login = lazy(() => import("./modules/auth/pages/Login"))
const Jobs = lazy(() => import("./modules/jobs/pages/Jobs"))
const JobHistory = lazy(() => import("./modules/jobs/pages/JobHistory"))
const JobLogs = lazy(() => import("./modules/jobs/pages/JobLogs"))
const JobSettings = lazy(() => import("./modules/jobs/pages/JobSettings"))
const JobCreation = lazy(() => import("./modules/jobs/pages/JobCreation"))
const JobEdit = lazy(() => import("./modules/jobs/pages/JobEdit"))
const Sources = lazy(() => import("./modules/sources/pages/Sources"))
const SourceEdit = lazy(() => import("./modules/sources/pages/SourceEdit"))
const CreateSource = lazy(() => import("./modules/sources/pages/CreateSource"))
const Destinations = lazy(
	() => import("./modules/destinations/pages/Destinations"),
)
const DestinationEdit = lazy(
	() => import("./modules/destinations/pages/DestinationEdit.tsx"),
)
const CreateDestination = lazy(
	() => import("./modules/destinations/pages/CreateDestination"),
)

// Loading component
const LoadingFallback = () => (
	<div className="flex h-[calc(100vh-64px)] items-center justify-center">
		<Spin size="large" />
	</div>
)

// Auth loading component
const AuthLoadingScreen = () => (
	<div className="flex h-screen items-center justify-center">
		<div className="text-center">
			<Spin size="large" />
			<p className="mt-4 text-gray-500">Logging in...</p>
		</div>
	</div>
)

// Main app content
const AppContent = () => {
	const { isAuthLoading, isAuthenticated, initAuth } = useAppStore()

	useEffect(() => {
		initAuth()
	}, [initAuth])

	if (isAuthLoading) {
		return <AuthLoadingScreen />
	}

	if (!isAuthenticated) {
		return (
			<Router>
				<Suspense fallback={<LoadingFallback />}>
					<Routes>
						<Route
							path="/login"
							element={<Login />}
						/>
						<Route
							path="*"
							element={
								<Navigate
									to="/login"
									replace
								/>
							}
						/>
					</Routes>
				</Suspense>
			</Router>
		)
	}

	return (
		<Router>
			<Layout>
				<Suspense fallback={<LoadingFallback />}>
					<Routes>
						<Route
							path="/jobs"
							element={<Jobs />}
						/>
						<Route
							path="/jobs/new"
							element={<JobCreation />}
						/>
						<Route
							path="/jobs/:jobId/edit"
							element={<JobEdit />}
						/>
						<Route
							path="/jobs/:jobId/history"
							element={<JobHistory />}
						/>
						<Route
							path="/jobs/:jobId/history/:historyId/logs"
							element={<JobLogs />}
						/>
						<Route
							path="/jobs/:jobId/tasks/:taskId/logs"
							element={<JobLogs />}
						/>
						<Route
							path="/jobs/:jobId/settings"
							element={<JobSettings />}
						/>
						<Route
							path="/sources"
							element={<Sources />}
						/>
						<Route
							path="/sources/new"
							element={<CreateSource />}
						/>
						<Route
							path="/sources/:sourceId"
							element={<SourceEdit />}
						/>
						<Route
							path="/destinations"
							element={<Destinations />}
						/>
						<Route
							path="/destinations/new"
							element={<CreateDestination />}
						/>
						<Route
							path="/destinations/:destinationId"
							element={<DestinationEdit />}
						/>
						<Route
							path="*"
							element={
								<Navigate
									to="/jobs"
									replace
								/>
							}
						/>
					</Routes>
				</Suspense>
			</Layout>
		</Router>
	)
}

function App() {
	return (
		<ConfigProvider
			theme={{
				token: {
					colorPrimary: "#203FDD",
					borderRadius: 6,
				},
			}}
		>
			<AntApp>
				<ErrorBoundary>
					<AppContent />
				</ErrorBoundary>
			</AntApp>
		</ConfigProvider>
	)
}

export default App
