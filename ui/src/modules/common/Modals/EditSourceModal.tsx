import { Button, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { getConnectorImage } from "../../../utils/utils"
import { CheckCircle, PencilLine, Warning } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"

const EditSourceModal = ({
	mockAssociatedJobs,
}: {
	mockAssociatedJobs: any[]
}) => {
	const navigate = useNavigate()
	const {
		showEditSourceModal,
		setShowEditSourceModal,
		showSuccessModal,
		setShowSuccessModal,
	} = useAppStore()

	const handleEdit = () => {
		setShowEditSourceModal(false)
		// Show success modal
		setShowSuccessModal(true)

		// Redirect after 2 seconds
		setTimeout(() => {
			setShowSuccessModal(false)
			navigate("/sources")
		}, 1000)
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
								dataIndex: "state",
								key: "status",
								render: () => <span className="text-yellow-500">warning</span>,
							},
							{
								title: "Last runtime",
								dataIndex: "lastRuntime",
								key: "lastRuntime",
							},
							{
								title: "Destination",
								dataIndex: "destination",
								key: "destination",
								render: (destination: any) => (
									<div className="flex items-center">
										<img
											src={getConnectorImage(destination?.type || "Amazon S3")}
											alt={destination?.type || "Amazon S3"}
											className="mr-2 size-6"
										/>
										Amazon S3 destination
									</div>
								),
							},
						]}
						dataSource={mockAssociatedJobs.slice(0, 3).map((job, index) => ({
							...job,
							key: index,
							name: `MongoDB_source_${index + 2}`,
							lastRuntime: "3 hours ago",
							destination: {
								type: "Amazon S3",
							},
						}))}
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
