import { LinktreeLogoIcon, PlusIcon } from "@phosphor-icons/react"
import { Button, Tabs, Empty, Spin } from "antd"
import { useState } from "react"
import { useNavigate } from "react-router-dom"

import { SOURCE_ONBOARDING_DISMISSED_SESSION_KEY } from "@/common/constants"
import { trackEvent, AnalyticsEvent } from "@/core/analytics"
import { Entity } from "@/modules/ingestion/common/types"

import SourceEmptyState from "../components/SourceEmptyState"
import SourceOnboardingModal from "../components/SourceOnboardingModal"
import SourceTable from "../components/SourceTable"
import { sourceTabs } from "../constants"
import { useSources, useDeleteSource } from "../hooks"

const Sources: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
	// Initialize from sessionStorage so the onboarding modal stays dismissed
	// for the current browser session after the CTA is clicked once.
	const [isOnboardingDismissed, setIsOnboardingDismissed] = useState(() => {
		if (typeof window === "undefined") {
			return false
		}
		return (
			sessionStorage.getItem(SOURCE_ONBOARDING_DISMISSED_SESSION_KEY) === "true"
		)
	})
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

	const handleOnboardingCreateSource = () => {
		if (typeof window !== "undefined") {
			sessionStorage.setItem(SOURCE_ONBOARDING_DISMISSED_SESSION_KEY, "true")
		}
		setIsOnboardingDismissed(true)
		handleCreateSource()
	}

	const handleEditSource = (id: string) => {
		navigate(`/sources/${id}`)
	}

	const handleDeleteSource = (source: Entity) => {
		deleteSourceMutation.mutate(String(source.id))
	}

	const filteredSources = (): Entity[] => {
		if (activeTab === "active") {
			return sources.filter(
				source =>
					source?.jobs &&
					source.jobs.length > 0 &&
					source.jobs.some(job => job.activate === true),
			)
		} else if (activeTab === "inactive") {
			return sources.filter(
				source =>
					!source?.jobs ||
					source.jobs.length === 0 ||
					source.jobs.every(job => job.activate === false),
			)
		}
		return []
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

			<Tabs
				activeKey={activeTab}
				onChange={setActiveTab}
				className="mb-4"
				items={sourceTabs.map(tab => ({
					key: tab.key,
					label: tab.label,
					children: isLoadingSources ? (
						<div className="flex items-center justify-center py-16">
							<Spin
								size="large"
								tip="Loading sources..."
							/>
						</div>
					) : tab.key === "active" && showEmpty ? (
						<SourceEmptyState handleCreateSource={handleCreateSource} />
					) : filteredSources().length === 0 ? (
						<Empty
							image={Empty.PRESENTED_IMAGE_SIMPLE}
							description="No sources configured"
							className="flex flex-col items-center"
						/>
					) : (
						<SourceTable
							sources={filteredSources()}
							loading={isLoadingSources}
							onEdit={handleEditSource}
							onDelete={handleDeleteSource}
						/>
					),
				}))}
			/>
			<SourceOnboardingModal
				open={activeTab === "active" && showEmpty && !isOnboardingDismissed}
				handleCreateSource={handleOnboardingCreateSource}
			/>
		</div>
	)
}

export default Sources
