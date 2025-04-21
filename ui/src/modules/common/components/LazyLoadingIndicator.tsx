import { Spin } from "antd"

interface LazyLoadingIndicatorProps {
	message?: string
}

const LazyLoadingIndicator: React.FC<LazyLoadingIndicatorProps> = ({
	message = "Loading...",
}) => {
	return (
		<div className="flex h-[calc(100vh-64px)] flex-col items-center justify-center">
			<Spin size="large" />
			<p className="mt-4 text-gray-500">{message}</p>
		</div>
	)
}

export default LazyLoadingIndicator
