import { PathIcon, PlusIcon } from "@phosphor-icons/react"
import { Button, Spin } from "antd"
import { useNavigate } from "react-router"

import { trackEvent, AnalyticsEvent } from "@/core/analytics"
import { Entity } from "@/modules/ingestion/common/types"

import DestinationEmptyState from "../components/DestinationEmptyState"
import DestinationTable from "../components/DestinationTable"
import { useDestinations, useDeleteDestination } from "../hooks"

const Destinations: React.FC = () => {
	const navigate = useNavigate()

	const {
		data: destinations = [],
		isLoading: isLoadingDestinations,
		error: destinationsError,
		refetch: refetchDestinations,
	} = useDestinations()
	const deleteDestinationMutation = useDeleteDestination()

	const handleCreateDestination = () => {
		trackEvent(AnalyticsEvent.CreateDestinationClicked)
		navigate("/destinations/new")
	}

	const handleEditDestination = (id: string) => {
		navigate(`/destinations/${id}`)
	}

	const handleDeleteDestination = (destination: Entity) => {
		deleteDestinationMutation.mutate(String(destination.id))
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
					data-testid="create-destination-button"
					onClick={handleCreateDestination}
					className="flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white hover:bg-primary-600"
				>
					<PlusIcon className="size-4 text-white" />
					Create Destination
				</button>
			</div>

			<p className="mb-6 text-gray-600">A list of all your destinations</p>

			{isLoadingDestinations ? (
				<div className="flex items-center justify-center py-16">
					<Spin
						size="large"
						tip="Loading destinations..."
					/>
				</div>
			) : showEmpty ? (
				<DestinationEmptyState
					handleCreateDestination={handleCreateDestination}
				/>
			) : (
				<DestinationTable
					destinations={destinations}
					loading={isLoadingDestinations}
					onEdit={handleEditDestination}
					onDelete={handleDeleteDestination}
				/>
			)}
		</div>
	)
}

export default Destinations
