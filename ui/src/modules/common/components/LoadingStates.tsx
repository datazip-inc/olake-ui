import Loader from "./Loader"

export const LoadingFallback = () => (
	<div className="flex h-[calc(100vh-64px)] items-center justify-center">
		<Loader size="large" />
	</div>
)

export const AuthLoadingScreen = () => (
	<div className="flex h-screen items-center justify-center">
		<div className="text-center">
			<Loader size="large" />
		</div>
	</div>
)
