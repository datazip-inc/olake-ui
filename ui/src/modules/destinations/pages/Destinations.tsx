import { useState, useEffect } from "react"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { useNavigate } from "react-router-dom"
import { useAppStore } from "../../../store"
import DestinationTable from "../components/DestinationTable"
import FirstDestination from "../../../assets/FirstDestination.svg"
import DestinationTutorial from "../../../assets/DestinationTutorial.svg"
import { DownloadSimple, PlayCircle, Plus } from "@phosphor-icons/react"
import { Entity } from "../../../types"

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
				destination => destination?.jobs && destination.jobs.length > 0,
			)
		} else if (activeTab === "inactive") {
			return destinations.filter(
				destination => !destination?.jobs || destination.jobs.length === 0,
			)
		} else if (activeTab === "saved") {
			return []
		}
		return []
	}

	const showEmpty = !isLoadingDestinations && destinations.length === 0

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
					<DownloadSimple className="mr-2 size-6" />
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
						<div className="flex flex-col items-center justify-center py-16">
							<img
								src={FirstDestination}
								alt="Empty state"
								className="mb-8 h-64 w-96"
							/>
							<div className="mb-2 text-[#193AE6]">Welcome User !</div>
							<h2 className="mb-2 text-2xl font-bold">
								Ready to create your first destination
							</h2>
							<p className="mb-8 text-[#0A0A0A]">
								Get started and experience the speed of OLake by running jobs
							</p>
							<Button
								type="primary"
								className="border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black"
								onClick={handleCreateDestination}
							>
								<Plus />
								New Destination
							</Button>
							<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-4 shadow-sm">
								<div className="flex items-center gap-4">
									<img
										src={DestinationTutorial}
										alt="Job Tutorial"
										className="rounded-lg"
									/>
									<div className="flex-1">
										<div className="mb-1 flex items-center gap-1 text-xs">
											<PlayCircle color="#9f9f9f" />
											<span className="text-[#9F9F9F]">OLake/ Tutorial</span>
										</div>
										<div className="text-xs">
											Checkout this tutorial, to know more about running jobs
										</div>
									</div>
								</div>
							</div>
						</div>
					) : filteredDestinations().length === 0 ? (
						<Empty
							image={Empty.PRESENTED_IMAGE_SIMPLE}
							description="No data"
							className="flex flex-col items-start"
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
