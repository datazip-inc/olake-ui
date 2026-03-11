import { Suspense, useEffect } from "react"
import { RouterProvider } from "react-router-dom"
import { ConfigProvider, App as AntApp } from "antd"

import { useAuthStore } from "@/core/auth/stores/authStore"
import { router } from "./router"
import { THEME_CONFIG } from "@/common/constants/constants"
import { AuthLoadingScreen } from "@/modules/ingestion/common/components/LoadingStates"
import {
	MutationCache,
	QueryClient,
	QueryClientProvider,
} from "@tanstack/react-query"

// After any mutation succeeds, invalidate all queries that share its mutationKey.
// Mutations with no mutationKey set will not invalidate anything.
const queryClient = new QueryClient({
	mutationCache: new MutationCache({
		onSuccess: (_data, _variables, _context, mutation) => {
			queryClient.invalidateQueries({
				queryKey: mutation.options.mutationKey,
			})
		},
	}),
})

function App() {
	const { isAuthLoading, initAuth } = useAuthStore()

	useEffect(() => {
		initAuth()
	}, [initAuth])

	if (isAuthLoading) {
		return <AuthLoadingScreen />
	}

	return (
		<QueryClientProvider client={queryClient}>
			<ConfigProvider theme={THEME_CONFIG}>
				<AntApp>
					<Suspense fallback={<AuthLoadingScreen />}>
						<RouterProvider router={router} />
					</Suspense>
				</AntApp>
			</ConfigProvider>
		</QueryClientProvider>
	)
}

export default App
