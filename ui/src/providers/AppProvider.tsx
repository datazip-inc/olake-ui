import {
	MutationCache,
	QueryClient,
	QueryClientProvider,
} from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { App as AntApp, ConfigProvider } from "antd"
import { useState, type ReactNode } from "react"

import { THEME_CONFIG } from "@/lib/theme"

interface AppProviderProps {
	children: ReactNode
}

export function AppProvider({ children }: AppProviderProps) {
	const [queryClient] = useState(
		() =>
			new QueryClient({
				mutationCache: new MutationCache({
					onSuccess: (_data, _variables, _context, mutation) => {
						if (!mutation.options.mutationKey) return
						queryClient.invalidateQueries({
							queryKey: mutation.options.mutationKey,
						})
					},
				}),
			}),
	)

	return (
		<QueryClientProvider client={queryClient}>
			<ConfigProvider theme={THEME_CONFIG}>
				<AntApp>
					{children}
					{import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
				</AntApp>
			</ConfigProvider>
		</QueryClientProvider>
	)
}
