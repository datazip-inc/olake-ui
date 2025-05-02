import { Button, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { getConnectorImage } from "../../../utils/utils"
import { CheckCircle, PencilLine, Warning } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"
import { message } from "antd"
import { formatDistanceToNow } from "date-fns"

const EditSourceModal = () => {
	const navigate = useNavigate()
	const {
		showEditSourceModal,
		setShowEditSourceModal,
		showSuccessModal,
		setShowSuccessModal,
		selectedSource,
		updateSource,
	} = useAppStore()

	const handleEdit = async () => {
		if (!selectedSource?.id) {
			message.error("Source ID is missing")
			return
		}

		try {
			setShowEditSourceModal(false)
			await updateSource(selectedSource.id.toString(), selectedSource)
			setShowSuccessModal(true)
			setTimeout(() => {
				setShowSuccessModal(false)
				navigate("/sources")
			}, 1000)
		} catch (error) {
			message.error("Failed to update source")
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
				open={showEditSourceModal}
				onCancel={() => setShowEditSourceModal(false)}
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
						onClick={() => setShowEditSourceModal(false)}
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
						Editing this source will affect the following jobs that are
						associated with this source and as a result will fail immediately.
						Do you still want to edit the source?
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
						]}
						dataSource={selectedSource?.jobs}
						pagination={false}
						rowKey="key"
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

export default EditSourceModal
