import { useState, useEffect } from "react"
import { Button, Tabs, Empty, message } from "antd"
import { useNavigate } from "react-router-dom"
import { useAppStore } from "../../../store"
import DestinationTable from "../components/DestinationTable"
import FirstDestination from "../../../assets/firstDestination.png"
import DestinationTutorial from "../../../assets/DestinationTutorial.png"
import { DownloadSimple, Plus } from "@phosphor-icons/react"

const Destinations: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
	const navigate = useNavigate()
	const {
		destinations,
		isLoadingDestinations,
		destinationsError,
		fetchDestinations,
		deleteDestination,
	} = useAppStore()

	useEffect(() => {
		fetchDestinations().catch(error => {
			message.error("Failed to fetch destinations")
			console.error(error)
		})
	}, [fetchDestinations])

	const handleCreateDestination = () => {
		navigate("/destinations/new")
	}

	const handleEditDestination = (id: string) => {
		navigate(`/destinations/${id}`)
	}

	const handleDeleteDestination = (id: string) => {
		message.info(`Deleting destination ${id}`)
		deleteDestination(id).catch(error => {
			message.error("Failed to delete destination")
			console.error(error)
		})
	}

	const filteredDestinations = destinations.filter(
		destination => destination.status === activeTab,
	)
	const showEmpty = destinations.length === 0

	const destinationTabs = [
		{ key: "active", label: "Active destinations" },
		{ key: "inactive", label: "Inactive destinations" },
		{ key: "saved", label: "Saved destinations" },
	]

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
					<DownloadSimple className="mr-2" />
					<h1 className="text-2xl font-bold">Destinations</h1>
				</div>
				<Button
					type="primary"
					className="bg-blue-600"
					icon={<Plus size={16} />}
					onClick={handleCreateDestination}
				>
					Create Destination
				</Button>
			</div>

			<p className="mb-6 text-gray-600">A list of all your destinations</p>

			<Tabs
				activeKey={activeTab}
				onChange={setActiveTab}
				className="mb-4"
				items={destinationTabs.map(tab => ({
					key: tab.key,
					label: tab.label,
					children:
						tab.key === "active" && showEmpty ? (
							<div className="flex flex-col items-center justify-center py-16">
								<img
									src={FirstDestination}
									alt="Empty state"
									className="mb-8 h-64 w-96"
								/>
								<div className="mb-2 text-blue-600">Welcome User !</div>
								<h2 className="mb-2 text-3xl font-bold">
									Ready to create your first destination
								</h2>
								<p className="mb-8 text-gray-600">
									Get started and experience the speed of OLake by running jobs
								</p>
								<Button
									type="primary"
									className="mb-12 bg-blue-600"
									onClick={handleCreateDestination}
								>
									New Destination
								</Button>
								<div className="w-96 rounded-lg bg-white p-4 shadow-sm">
									<div className="flex items-center gap-4">
										<img
											src={DestinationTutorial}
											alt="Job Tutorial"
											className="h-16 w-24 rounded-lg"
										/>
										<div className="flex-1">
											<div className="mb-1 text-xs text-gray-500">
												OLake/ Tutorial
											</div>
											<div className="text-sm">
												Checkout this tutorial, to know more about running jobs
											</div>
										</div>
									</div>
								</div>
							</div>
						) : filteredDestinations.length === 0 ? (
							<Empty
								image={Empty.PRESENTED_IMAGE_SIMPLE}
								description="No data"
							/>
						) : (
							<DestinationTable
								destinations={filteredDestinations}
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
