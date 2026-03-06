import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { PathIcon, PlusIcon } from "@phosphor-icons/react"
import { Button, Tabs, Empty, Spin } from "antd"

import analyticsService from "@/common/services/analyticsService"
import { Entity } from "../../../common/types"
import DestinationEmptyState from "../components/DestinationEmptyState"
import DestinationTable from "../components/DestinationTable"
import { AnalyticsEvent } from "@/common/enums"
import { destinationTabs } from "../constants"
import { useDestinations, useDeleteDestination } from "../hooks"

const Destinations: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
	const navigate = useNavigate()

	const {
		data: destinations = [],
		isLoading: isLoadingDestinations,
		error: destinationsError,
		refetch: refetchDestinations,
	} = useDestinations()
	const deleteDestinationMutation = useDeleteDestination()

	const handleCreateDestination = () => {
		analyticsService.trackEvent(AnalyticsEvent.CreateDestinationClicked)
		navigate("/destinations/new")
	}

	const handleEditDestination = (id: string) => {
		navigate(`/destinations/${id}`)
	}

	const handleDeleteDestination = (destination: Entity) => {
		deleteDestinationMutation.mutate(String(destination.id))
	}

	const filteredDestinations = (): Entity[] => {
		// a destination is active if it has jobs and at least one job is active
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
					Error loading destinations: {destinationsError.message}
				</div>
				<Button
					onClick={() => refetchDestinations()}
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
