import { useState, useEffect } from "react"
import { Button, Tabs, Empty, message } from "antd"
import { useNavigate } from "react-router-dom"
import { useAppStore } from "../../../store"
import SourceTable from "../components/SourceTable"
import FirstSource from "../../../assets/firstSource.png"
import SourcesTutorial from "../../../assets/sourcesTutorial.png"
import { LinktreeLogo, Plus } from "@phosphor-icons/react"

const Sources: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
	const navigate = useNavigate()
	const {
		sources,
		isLoadingSources,
		sourcesError,
		fetchSources,
		deleteSource,
	} = useAppStore()

	useEffect(() => {
		fetchSources().catch(error => {
			message.error("Failed to fetch sources")
			console.error(error)
		})
	}, [fetchSources])

	const handleCreateSource = () => {
		navigate("/sources/new")
	}

	const handleEditSource = (id: string) => {
		navigate(`/sources/${id}`)
	}

	const handleDeleteSource = (id: string) => {
		message.info(`Deleting source ${id}`)
		deleteSource(id).catch(error => {
			message.error("Failed to delete source")
			console.error(error)
		})
	}

	const filteredSources = sources.filter(source => source.status === activeTab)
	const showEmpty = sources.length === 0

	const sourceTabs = [
		{ key: "active", label: "Active sources" },
		{ key: "inactive", label: "Inactive sources" },
		{ key: "saved", label: "Saved sources" },
	]

	if (sourcesError) {
		return (
			<div className="p-6">
				<div className="text-red-500">
					Error loading sources: {sourcesError}
				</div>
				<Button
					onClick={() => fetchSources()}
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
				<Button
					type="primary"
					className="bg-blue-600"
					icon={<Plus size={16} />}
					onClick={handleCreateSource}
				>
					Create Source
				</Button>
			</div>

			<p className="mb-6 text-gray-600">A list of all your sources</p>

			<Tabs
				activeKey={activeTab}
				onChange={setActiveTab}
				className="mb-4"
				items={sourceTabs.map(tab => ({
					key: tab.key,
					label: tab.label,
					children:
						tab.key === "active" && showEmpty ? (
							<div className="flex flex-col items-center justify-center py-16">
								<img
									src={FirstSource}
									alt="Empty state"
									className="mb-8 h-64 w-96"
								/>
								<div className="mb-2 text-blue-600">Welcome User !</div>
								<h2 className="mb-2 text-3xl font-bold">
									Ready to create your first source
								</h2>
								<p className="mb-8 text-gray-600">
									Get started and experience the speed of OLake by running jobs
								</p>
								<Button
									type="primary"
									className="border-1 mb-12 border-gray-300 bg-white p-4 text-black"
									onClick={handleCreateSource}
								>
									New Source
								</Button>
								<div className="w-96 rounded-lg bg-white p-4 shadow-sm">
									<div className="flex items-center gap-4">
										<img
											src={SourcesTutorial}
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
						) : filteredSources.length === 0 ? (
							<Empty
								image={Empty.PRESENTED_IMAGE_SIMPLE}
								description="No data"
							/>
						) : (
							<SourceTable
								sources={filteredSources}
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
