import { Button, message, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { Warning } from "@phosphor-icons/react"
import { Entity } from "../../../types"
import { getConnectorImage } from "../../../utils/utils"
import { DeleteModalProps } from "../../../types/modalTypes"
import { formatDistanceToNow } from "date-fns"

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
		deleteSource(selectedSource?.id as unknown as string).catch(error => {
			message.error("Failed to delete source")
			console.error(error)
		})
		setShowDeleteModal(false)
	}
	const handleDeleteDestination = () => {
		message.info(`Deleting destination ${selectedDestination?.name}`)
		deleteDestination(selectedDestination?.id as unknown as string).catch(
			error => {
				message.error("Failed to delete destination")
				console.error(error)
			},
		)
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
			dataIndex: "activate",
			key: "activate",
			render: (activate: boolean) => (
				<span
					className={`rounded px-2 py-1 text-xs ${
						!activate
							? "bg-[#FFF1F0] text-[#F5222D]"
							: "bg-primary-200 text-primary-700"
					}`}
				>
					{activate ? "Active" : "Inactive"}
				</span>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "last_run_time",
			key: "last_run_time",
			render: (text: string) => (
				<div className="flex justify-center">
					{text !== undefined
						? formatDistanceToNow(new Date(text), {
								addSuffix: true,
							})
						: "-"}
				</div>
			),
		},
		...(fromSource
			? [
					{
						title: "Destination",
						dataIndex: "destination_name",
						key: "destination_name",
						render: (destination_name: string, record: any) => (
							<div className="flex items-center">
								<img
									src={getConnectorImage(record.destination_type || "")}
									alt={record.destination_type}
									className="mr-2 size-6"
								/>
								{destination_name || "N/A"}
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
									alt={record.destination_type}
									className="mr-2 size-6"
								/>
								{source_name || "N/A"}
							</div>
						),
					},
				]),
	]

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
					className="w-full rounded-md border"
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
