import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { LinktreeLogo, Plus } from "@phosphor-icons/react"
import EmptyState from "../../common/components/EmptyState"
import { useAppStore } from "../../../store"
import analyticsService from "../../../api/services/analyticsService"
import { Entity } from "../../../types"
import {EmptyStateType, sourceTabs } from "../../../utils/constants"
import SourceTable from "../components/SourceTable"


const Sources: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
	const navigate = useNavigate()
	const {
		sources,
		isLoadingSources,
		sourcesError,
		fetchSources,
		setShowDeleteModal,
		setSelectedSource,
		deleteSource,
	} = useAppStore()

	useEffect(() => {
		fetchSources().catch(error => {
			message.error("Failed to fetch sources")
			console.error(error)
		})
	}, [fetchSources])

	const handleCreateSource = () => {
		analyticsService.trackEvent("create_source_clicked")
		navigate("/sources/new")
	}

	const handleEditSource = (id: string) => {
		navigate(`/sources/${id}`)
	}

	const handleDeleteSource = (source: Entity) => {
		setSelectedSource(source)

		// For inactive sources, delete directly without showing modal
		if (!source?.jobs || source.jobs.length === 0) {
			message.info(`Deleting source ${source?.name}`)
			deleteSource(String(source.id)).catch(error => {
				message.error("Failed to delete source")
				console.error(error)
			})
			return
		}

		// For active sources with jobs, show the delete confirmation modal
		setTimeout(() => {
			setShowDeleteModal(true)
		}, 1000)
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
					Error loading sources: {sourcesError}
				</div>
				<Button
					onClick={() => {
						fetchSources()
					}}
					className="mt-4"
				>
					Retry
				</Button>
			</div>
		)
	}

	return (
		<div className="p-6">
			<div className="mb-4 flex items-center justify-between">
				<div className="flex items-center">
					<LinktreeLogo className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">Sources</h1>
				</div>
				<button
					className="flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white hover:bg-primary-600"
					onClick={handleCreateSource}
				>
					<Plus className="size-4 text-white" />
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
						<EmptyState
							type={EmptyStateType.SOURCE}
							onButtonClick={handleCreateSource}
						/>
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
		</div>
	)
}

export default Sources
