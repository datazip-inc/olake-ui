interface EndpointConfigSkeletonProps {
	includeWrapper?: boolean
}

const EndpointConfigSkeleton = ({
	includeWrapper = true,
}: EndpointConfigSkeletonProps) => {
	const content = (
		<>
			{includeWrapper && (
				<div className="mb-6">
					<div className="h-5 w-32 animate-pulse rounded bg-gray-200" />
				</div>
			)}

			<div className="space-y-6">
				<div className="space-y-2">
					<div className="h-4 w-24 animate-pulse rounded bg-gray-200" />
					<div className="h-10 w-full animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>

				<div className="space-y-2">
					<div className="h-4 w-32 animate-pulse rounded bg-gray-200" />
					<div className="h-10 w-full animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>

				<div className="space-y-2">
					<div className="h-4 w-28 animate-pulse rounded bg-gray-200" />
					<div className="h-10 w-full animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>

				<div className="space-y-2">
					<div className="h-4 w-20 animate-pulse rounded bg-gray-200" />
					<div className="h-10 w-2/3 animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>

				<div className="space-y-2">
					<div className="h-4 w-36 animate-pulse rounded bg-gray-200" />
					<div className="h-10 w-full animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>

				<div className="space-y-3 rounded-md border border-gray-200 bg-gray-50 p-4">
					<div className="flex items-center justify-between">
						<div className="h-4 w-40 animate-pulse rounded bg-gray-200" />
						<div className="size-4 animate-pulse rounded bg-gray-200" />
					</div>
				</div>

				<div className="space-y-2">
					<div className="h-4 w-24 animate-pulse rounded bg-gray-200" />
					<div className="h-10 w-full animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>

				<div className="space-y-2">
					<div className="h-4 w-28 animate-pulse rounded bg-gray-200" />
					<div className="h-24 w-full animate-pulse rounded-md border border-gray-200 bg-gray-50" />
				</div>
			</div>
		</>
	)

	if (includeWrapper) {
		return (
			<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
				{content}
			</div>
		)
	}

	return content
}

export default EndpointConfigSkeleton
