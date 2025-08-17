import { useState, useEffect } from "react"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { useNavigate } from "react-router-dom"
import { useAppStore } from "../../../store"
import DestinationTable from "../components/DestinationTable"
import { Path, Plus } from "@phosphor-icons/react"
import { Entity } from "../../../types"
import { destinationTabs, EmptyStateType } from "../../../utils/constants"
import analyticsService from "../../../api/services/analyticsService"
import EmptyState from "../../common/components/EmptyState"

const Destinations: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
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
		if (activeTab === "active") {
			return destinations.filter(
				destination =>
					destination?.jobs &&
					destination.jobs.length > 0 &&
					destination.jobs.some(job => job.activate === true),
			)
		} else if (activeTab === "inactive") {
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
					<Path className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">Destinations</h1>
				</div>
				<button
					onClick={handleCreateDestination}
					className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-2 font-light text-white hover:bg-[#132685]"
				>
					<Plus className="size-4 text-white" />
					Create Destination
				</button>
			</div>

			<p className="mb-6 text-gray-600">A list of all your destinations</p>

			<Tabs
				activeKey={activeTab}
				onChange={setActiveTab}
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
					) : tab.key === "active" && showEmpty ? (
						<EmptyState
							type={EmptyStateType.DESTINATION}
							onButtonClick={handleCreateDestination}
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
