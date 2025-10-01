import { Suspense, useEffect } from "react"
import { RouterProvider } from "react-router-dom"
import { ConfigProvider, App as AntApp } from "antd"

import { useAppStore } from "./store"
import { router } from "./routes"
import { THEME_CONFIG } from "./utils/constants"
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
		<ConfigProvider theme={THEME_CONFIG}>
			<AntApp>
				<Suspense fallback={<AuthLoadingScreen />}>
					<RouterProvider router={router} />
				</Suspense>
			</AntApp>
		</ConfigProvider>
	)
}

export default App
