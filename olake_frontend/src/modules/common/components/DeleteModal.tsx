import { Button, message, Modal, Table } from "antd"
import Error from "../../../assets/Error.svg"
import { useAppStore } from "../../../store"
import {
	ArrowsCounterClockwise,
	CheckCircle,
	XCircle,
} from "@phosphor-icons/react"
import { useEffect } from "react"
interface DeleteModalProps {
	from: "DESTINATION" | "SOURCE"
}
const DeleteModal = ({ from }: DeleteModalProps) => {
	const { showDeleteModal } = useAppStore()
	const {
		jobs,
		fetchJobs,
		selectedSource,
		selectedDestination,
		setShowDeleteModal,
		deleteSource,
		deleteDestination,
	} = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchJobs, jobs.length])
	const handleDeleteMaster = () => {
		if (from === "DESTINATION") {
			handleDeleteDestination()
		} else {
			handleDeleteSource()
		}
	}
	const associatedJobs =
		from == "DESTINATION"
			? jobs
					.filter(job => job.destination === selectedDestination?.name)
					.map(job => ({
						...job,
						lastRuntime: "3 hours ago",
					}))
			: jobs
					.filter(job => job.source === selectedSource?.name)
					.map(job => ({
						...job,
						lastRuntime: "3 hours ago",
					}))

	const handleDeleteSource = () => {
		message.info(`Deleting source ${selectedSource?.name}`)
		deleteSource(selectedSource?.id as string).catch(error => {
			message.error("Failed to delete source")
			console.error(error)
		})
		setShowDeleteModal(false)
	}
	const handleDeleteDestination = () => {
		message.info(`Deleting destination ${selectedDestination?.name}`)
		deleteDestination(selectedDestination?.id as string).catch(error => {
			message.error("Failed to delete destination")
			console.error(error)
		})
		setShowDeleteModal(false)
	}

	const getStatusIcon = (status: string | undefined) => {
		if (status === "success") {
			return <CheckCircle className="text-green-500" />
		} else if (status === "failed") {
			return <XCircle className="text-red-500" />
		} else if (status === "running") {
			return <ArrowsCounterClockwise className="text-blue-500" />
		}
		return null
	}
	const loading = false

	const columns = [
		{
			title: "Name",
			dataIndex: "name",
			key: "name",
		},
		{
			title: "Last sync status",
			dataIndex: "lastSyncStatus",
			key: "lastSyncStatus",
			render: (status: string) => (
				<div className="flex items-center">
					{getStatusIcon(status)}
					<span
						className={`ml-1 ${
							status === "success"
								? "text-green-500"
								: status === "failed"
									? "text-red-500"
									: ""
						}`}
					>
						{status === "success"
							? "Success"
							: status === "failed"
								? "Failed"
								: status}
					</span>
				</div>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "lastRuntime",
			key: "lastRuntime",
		},
		...(from === "DESTINATION"
			? [
					{
						title: "Destination",
						dataIndex: "destination",
						key: "destination",
						render: (text: string) => (
							<div className="flex items-center">
								<span className="mr-2 inline-block h-2 w-2 rounded-full bg-red-600"></span>
								{text}
							</div>
						),
					},
				]
			: [
					{
						title: "Source",
						dataIndex: "source",
						key: "source",
						render: (text: string) => (
							<div className="flex items-center">
								<span className="mr-2 inline-block h-2 w-2 rounded-full bg-blue-600"></span>
								{text}
							</div>
						),
					},
				]),
	]

	return (
		<Modal
			open={showDeleteModal}
			footer={null}
			closable={false}
			centered
			width={600}
		>
			<div className="flex flex-col items-center justify-center gap-7 py-8">
				<img src={Error} />
				<div className="flex flex-col items-center">
					{from === "DESTINATION" ? (
						<div className="text-center text-xl font-medium text-[#2B2B2B]">
							Deleting {selectedDestination?.name} source will disable these
							jobs. Are you sure you want to continue?
						</div>
					) : (
						<div className="text-center text-xl font-medium text-[#2B2B2B]">
							Deleting ${selectedSource?.name} source will disable these jobs.
							Are you sure you want to continue?
						</div>
					)}
				</div>
				<Table
					dataSource={associatedJobs}
					columns={columns}
					rowKey="id"
					loading={loading}
					pagination={false}
					className="overflow-hidden rounded-l border"
				/>
				<div className="flex w-full justify-end space-x-2">
					<Button
						className="px-4 py-4"
						type="primary"
						danger
						onClick={handleDeleteMaster}
					>
						Delete
					</Button>
					<Button
						className="px-4 py-4"
						type="default"
						onClick={() => {
							setShowDeleteModal(false)
						}}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default DeleteModal
