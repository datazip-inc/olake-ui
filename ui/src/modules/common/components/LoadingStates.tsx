import { Spin } from "antd"

export const LoadingFallback = () => (
	<div className="flex h-[calc(100vh-64px)] items-center justify-center">
		<Spin size="large" />
	</div>
)

export const AuthLoadingScreen = () => (
	<div className="flex h-screen items-center justify-center">
		<div className="text-center">
			<Spin size="large" />
			<p className="mt-4 text-gray-500">Logging in...</p>
		</div>
	</div>
)
