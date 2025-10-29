import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { PathIcon, PlusIcon } from "@phosphor-icons/react"
import { Button, Tabs, Empty, message, Spin } from "antd"

import analyticsService from "../../../api/services/analyticsService"
import { useAppStore } from "../../../store"
import { Entity } from "../../../types"
import { destinationTabs } from "../../../utils/constants"
import { SourceStatus } from "../../../enums"
import DestinationEmptyState from "../components/DestinationEmptyState"
import DestinationTable from "../components/DestinationTable"

const Destinations: React.FC = () => {
	const [activeTab, setActiveTab] = useState<SourceStatus>(SourceStatus.ACTIVE)
	const navigate = useNavigate()
	const {
		destinations,
		isLoadingDestinations,
		destinationsError,
		fetchDestinations,
		setShowDeleteModal,
		setSelectedDestination,
		deleteDestination,
	} = useAppStore()

	useEffect(() => {
		fetchDestinations().catch(error => {
			message.error("Failed to fetch destinations")
			console.error(error)
		})
	}, [fetchDestinations])

	const handleCreateDestination = () => {
		analyticsService.trackEvent("create_destination_clicked")
		navigate("/destinations/new")
	}

	const handleEditDestination = (id: string) => {
		navigate(`/destinations/${id}`)
	}

	const handleDeleteDestination = (destination: Entity) => {
		setSelectedDestination(destination)

		// For inactive destinations, delete directly without showing modal
		if (!destination?.jobs || destination.jobs.length === 0) {
			message.info(`Deleting destination ${destination?.name}`)
			deleteDestination(String(destination.id)).catch(error => {
				message.error("Failed to delete destination")
				console.error(error)
			})
			return
		}

		// For active destinations with jobs, show the delete confirmation modal
		setTimeout(() => {
			setShowDeleteModal(true)
		}, 1000)
	}

		const filteredDestinations = (): Entity[] => {
			// a destination is active if it has jobs and at least one job is active
			if (activeTab === SourceStatus.ACTIVE) {
				return destinations.filter(
					destination =>
						destination?.jobs &&
						destination.jobs.length > 0 &&
						destination.jobs.some(job => job.activate === true),
				)
			} else if (activeTab === SourceStatus.INACTIVE) {
				return destinations.filter(
					destination =>
						!destination?.jobs ||
						destination.jobs.length === 0 ||
						destination.jobs.every(job => job.activate === false),
				)
			}
			return []
		}

	const showEmpty = !isLoadingDestinations && destinations.length === 0

	if (destinationsError) {
		return (
			<div className="p-6">
				<div className="text-red-500">
					Error loading destinations: {destinationsError}
				</div>
				<Button
					onClick={() => fetchDestinations()}
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
					<PathIcon className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">Destinations</h1>
				</div>
				<button
					onClick={handleCreateDestination}
					className="flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white hover:bg-primary-600"
				>
					<PlusIcon className="size-4 text-white" />
					Create Destination
				</button>
			</div>

			<p className="mb-6 text-gray-600">A list of all your destinations</p>

			<Tabs
				activeKey={activeTab}
				onChange={(key: string) => setActiveTab(key as SourceStatus)}
				className="mb-4"
				items={destinationTabs.map(tab => ({
					key: tab.key,
					label: tab.label,
					children: isLoadingDestinations ? (
						<div className="flex items-center justify-center py-16">
							<Spin
								size="large"
								tip="Loading destinations..."
							/>
						</div>
					) : tab.key === SourceStatus.ACTIVE && showEmpty ? (
						<DestinationEmptyState
							handleCreateDestination={handleCreateDestination}
						/>
					) : filteredDestinations().length === 0 ? (
						<Empty
							image={Empty.PRESENTED_IMAGE_SIMPLE}
							description="No destinations configured"
							className="flex flex-col items-center"
						/>
					) : (
						<DestinationTable
							destinations={filteredDestinations()}
							loading={isLoadingDestinations}
							onEdit={handleEditDestination}
							onDelete={handleDeleteDestination}
						/>
					),
				}))}
			/>
		</div>
	)
}

export default Destinations
