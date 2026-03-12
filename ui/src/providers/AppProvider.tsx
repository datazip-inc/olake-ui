import { QueryClientProvider } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { App as AntApp, ConfigProvider } from "antd"
import type { ReactNode } from "react"

import { queryClient } from "@/lib/queryClient"
import { THEME_CONFIG } from "@/lib/theme"

interface AppProviderProps {
	children: ReactNode
}

export function AppProvider({ children }: AppProviderProps) {
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
