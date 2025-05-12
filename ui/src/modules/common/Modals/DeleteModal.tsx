import { Button, message, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { getStatusIcon } from "../../../utils/statusIcons"
import { Warning } from "@phosphor-icons/react"
import { Entity } from "../../../types"
import { getConnectorImage } from "../../../utils/utils"

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
	let entity: Entity

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
			dataIndex: "name",
			key: "name",
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
						dataIndex: "dest_name",
						key: "dest_name",
						render: (dest_name: string, record: any) => (
							<div className="flex items-center">
								<img
									src={getConnectorImage(record.dest_type || "")}
									alt={record.dest_type}
									className="mr-2 size-6"
								/>
								{dest_name || "N/A"}
							</div>
						),
					},
				]
			: [
					{
						title: "Source",
						dataIndex: "source_name",
						key: "source_name",
						render: (source_name: string, record: any) => (
							<div className="flex items-center">
								<img
									src={getConnectorImage(record.source_type || "")}
									alt={record.dest_type}
									className="mr-2 size-6"
								/>
								{source_name || "N/A"}
							</div>
						),
					},
				]),
	]

	// Create an array with the entity if it exists
	const dataSource = entity?.jobs

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
					className="w-full rounded-[6px] border"
					rowClassName="no-hover"
					scroll={{ y: 300 }}
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
