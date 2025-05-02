import { Button, message, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { getStatusIcon } from "../../../utils/statusIcons"
import { Warning } from "@phosphor-icons/react"

interface DeleteModalProps {
	fromSource: boolean
}

const DeleteModal = ({ fromSource }: DeleteModalProps) => {
	const {
		showDeleteModal,
		setShowDeleteModal,
		selectedSource,
		selectedDestination,
		deleteSource,
		deleteDestination,
	} = useAppStore()
	let entity: Source | Destination | null = null

	if (fromSource) {
		entity = selectedSource
	} else {
		entity = selectedDestination
	}
	const handleDelete = () => {
		if (fromSource) {
			handleDeleteSource()
		} else {
			handleDeleteDestination()
		}
	}

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

	const loading = false

	const columns = [
		{
			title: "Name",
			dataIndex: "jobName",
			key: "jobName",
		},
		{
			title: "Status",
			dataIndex: "lastSyncStatus",
			key: "lastSyncStatus",
			render: (status: string) => (
				<div className="flex items-center">
					{getStatusIcon(status)}
					<span className="ml-1 rounded-[6px] bg-[#f6ffed] px-1.5 py-1 text-xs text-[#52c41a]">
						success
					</span>
				</div>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "lastRuntime",
			key: "lastRuntime",
			render: () => (
				<div className="flex items-center">
					<span>3 hours ago</span>
				</div>
			),
		},
		...(fromSource
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
							<div className="flex items-center">{text}</div>
						),
					},
				]),
	]

	// Create an array with the entity if it exists
	const dataSource = entity?.associatedJobs ? entity?.associatedJobs : []

	return (
		<Modal
			open={showDeleteModal}
			footer={null}
			closable={false}
			centered
			width={600}
		>
			<div className="flex flex-col items-center justify-center gap-7 py-8">
				<Warning
					weight="fill"
					className="h-[55px] w-[63px] text-[#F5222D]"
				/>
				<div className="flex flex-col items-center">
					<div className="text-center text-xl font-medium text-[#2B2B2B]">
						Deleting {entity?.name} {fromSource ? "source" : "destination"} will
						disable these <br></br>jobs. Are you sure you want to continue?
					</div>
				</div>

				<Table
					dataSource={dataSource}
					columns={columns}
					rowKey="id"
					loading={loading}
					pagination={false}
					className="w-full overflow-hidden rounded-[6px] border"
					rowClassName="no-hover"
				/>
				<div className="flex w-full justify-end space-x-2">
					<Button
						className="px-4 py-4"
						type="primary"
						danger
						onClick={handleDelete}
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
