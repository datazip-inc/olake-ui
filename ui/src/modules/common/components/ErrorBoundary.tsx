import { Component, ErrorInfo, ReactNode } from "react"

interface Props {
	children: ReactNode
	fallback?: ReactNode
}

interface State {
	hasError: boolean
	error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
	public state: State = {
		hasError: false,
		error: null,
	}

	public static getDerivedStateFromError(error: Error): State {
		return { hasError: true, error }
	}

	public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
		console.error("Uncaught error:", error, errorInfo)
	}

	public render() {
		if (this.state.hasError) {
			return (
				this.props.fallback || (
					<div className="flex min-h-screen items-center justify-center bg-gray-50">
						<div className="w-full max-w-md space-y-8 rounded-lg bg-white p-8 shadow-lg">
							<div className="text-center">
								<h2 className="mt-6 text-3xl font-extrabold text-gray-900">
									Oops! Something went wrong
								</h2>
								<p className="mt-2 text-sm text-gray-600">
									We apologize for the inconvenience. Please try refreshing the
									page.
								</p>
								{this.state.error && (
									<div className="mt-4 rounded-md bg-red-50 p-4">
										<p className="text-sm text-red-700">
											{this.state.error.message}
										</p>
									</div>
								)}
								<div className="mt-8">
									<button
										onClick={() => window.location.reload()}
										className="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
									>
										Refresh Page
									</button>
								</div>
							</div>
						</div>
					</div>
				)
			)
		}
		return this.props.children
	}
}
