import { Suspense, useEffect } from "react"
import { RouterProvider } from "react-router-dom"
import { ConfigProvider, App as AntApp } from "antd"
import { useAppStore } from "./store"
import { router } from "./routes"
import { AuthLoadingScreen } from "./modules/common/components/LoadingStates"

function App() {
	const { isAuthLoading, initAuth } = useAppStore()

	useEffect(() => {
		initAuth()
	}, [initAuth])

	if (isAuthLoading) {
		return <AuthLoadingScreen />
	}

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
				<Suspense fallback={<AuthLoadingScreen />}>
					<RouterProvider router={router} />
				</Suspense>
			</AntApp>
		</ConfigProvider>
	)
}

export default App
