import { Suspense, useEffect } from "react"
import { RouterProvider } from "react-router-dom"

import { useAuthStore } from "@/core/auth/stores/authStore"
import { AuthLoadingScreen } from "@/modules/ingestion/common/components/LoadingStates"
import { AppProvider } from "@/providers/AppProvider"

import { router } from "./router"

function App() {
	const { isAuthLoading, initAuth } = useAuthStore()

	useEffect(() => {
		initAuth()
	}, [initAuth])

	if (isAuthLoading) {
		return <AuthLoadingScreen />
	}

	return (
		<AppProvider>
			<Suspense fallback={<AuthLoadingScreen />}>
				<RouterProvider router={router} />
			</Suspense>
		</AppProvider>
	)
}

export default App
