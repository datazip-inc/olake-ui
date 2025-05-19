import { Button, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { getConnectorImage } from "../../../utils/utils"
import { CheckCircle, PencilLine, Warning } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"
import { message } from "antd"
import { destinationService } from "../../../api/services/destinationService"
import { formatDistanceToNow } from "date-fns"

const EditDestinationModal = () => {
	const navigate = useNavigate()
	const {
		showEditDestinationModal,
		setShowEditDestinationModal,
		showSuccessModal,
		setShowSuccessModal,
		selectedDestination,
	} = useAppStore()

	const handleEdit = async () => {
		if (!selectedDestination?.id) {
			message.error("Destination ID is missing")
			return
		}

		try {
			setShowEditDestinationModal(false)
			await destinationService.updateDestination(
				selectedDestination.id.toString(),
				selectedDestination,
			)
			setShowSuccessModal(true)
			setTimeout(() => {
				setShowSuccessModal(false)
				navigate("/destinations")
			}, 1000)
		} catch (error) {
			message.error("Failed to update destination")
			console.error(error)
		}
	}

	return (
		<>
			<Modal
				title={
					<div className="flex justify-center">
						<Warning
							weight="fill"
							className="size-12 text-[#203FDD]"
						/>
					</div>
				}
				open={showEditDestinationModal}
				onCancel={() => setShowEditDestinationModal(false)}
				footer={[
					<Button
						key="edit"
						type="primary"
						onClick={handleEdit}
						icon={<PencilLine size={16} />}
						className="bg-blue-600"
					>
						Edit
					</Button>,
					<Button
						key="cancel"
						onClick={() => setShowEditDestinationModal(false)}
					>
						Cancel
					</Button>,
				]}
				centered
				width="38%"
			>
				<div className="mt-4 text-center">
					<h3 className="text-lg font-medium">
						Due to the editing, the jobs are going to get affected
					</h3>
					<p className="mt-2 text-xs text-black text-opacity-45">
						Editing this destination will affect the following jobs that are
						associated with this destination and as a result will fail
						immediately. Do you still want to edit the destination?
					</p>
				</div>
				<div className="mt-6">
					<Table
						columns={[
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
												: "bg-[#E6F4FF] text-[#0958D9]"
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
								render: (text: string) =>
									formatDistanceToNow(new Date(text), { addSuffix: true }),
							},
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
						]}
						dataSource={selectedDestination?.jobs}
						pagination={false}
						rowKey="key"
						scroll={{ y: 300 }}
					/>
				</div>
			</Modal>

			{/* Success Modal */}
			<Modal
				open={showSuccessModal}
				footer={null}
				closable={false}
				centered
				width={400}
			>
				<div className="flex flex-col items-center justify-center gap-7 py-6">
					<CheckCircle
						weight="fill"
						className="size-16 text-[#13AA52]"
					/>
					<div className="flex flex-col items-center text-xl font-medium">
						Changes are saved successfully
					</div>
				</div>
			</Modal>
		</>
	)
}

export default EditDestinationModal
