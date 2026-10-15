import { LinktreeLogoIcon, PlusIcon } from "@phosphor-icons/react"
import { Button, Spin } from "antd"
import { useNavigate } from "react-router-dom"

import { trackEvent, AnalyticsEvent } from "@/core/analytics"
import { Entity } from "@/modules/ingestion/common/types"

import SourceEmptyState from "../components/SourceEmptyState"
import SourceTable from "../components/SourceTable"
import { useSources, useDeleteSource } from "../hooks"

const Sources: React.FC = () => {
	const navigate = useNavigate()

	const {
		data: sources = [],
		isLoading: isLoadingSources,
		error: sourcesError,
		refetch: refetchSources,
	} = useSources()
	const deleteSourceMutation = useDeleteSource()

	const handleCreateSource = () => {
		trackEvent(AnalyticsEvent.CreateSourceClicked)
		navigate("/sources/new")
	}

	const handleEditSource = (id: string) => {
		navigate(`/sources/${id}`)
	}

	const handleDeleteSource = (source: Entity) => {
		deleteSourceMutation.mutate(String(source.id))
	}

	const showEmpty = !isLoadingSources && sources.length === 0

	if (sourcesError) {
		return (
			<div className="p-6">
				<div className="text-red-500">
					Error loading sources: {sourcesError.message}
				</div>
				<Button
					onClick={() => refetchSources()}
					className="mt-4"
				>
					Retry
				</Button>
			</div>
		)
	}

	return (
		<div
			className="p-6"
			data-testid="sources-page"
			data-loaded={!isLoadingSources ? "true" : "false"}
		>
			<div className="mb-4 flex items-center justify-between">
				<div className="flex items-center">
					<LinktreeLogoIcon className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">Sources</h1>
				</div>
				<button
					data-testid="create-source-button"
					className="flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white hover:bg-primary-600"
					onClick={handleCreateSource}
				>
					<PlusIcon className="size-4 text-white" />
					Create Source
				</button>
			</div>

			<p className="mb-6 text-gray-600">A list of all your sources</p>

			{isLoadingSources ? (
				<div className="flex items-center justify-center py-16">
					<Spin
						size="large"
						tip="Loading sources..."
					/>
				</div>
			) : showEmpty ? (
				<SourceEmptyState handleCreateSource={handleCreateSource} />
			) : (
				<SourceTable
					sources={sources}
					loading={isLoadingSources}
					onEdit={handleEditSource}
					onDelete={handleDeleteSource}
				/>
			)}
		</div>
	)
}

export default Sources
