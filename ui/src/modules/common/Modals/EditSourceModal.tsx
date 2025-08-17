import { Button, Modal, Table, message } from "antd"
import { CheckCircle, Warning } from "@phosphor-icons/react"
import { formatDistanceToNow } from "date-fns"
import { useNavigate } from "react-router-dom"

import { sourceService } from "../../../api"
import { useAppStore } from "../../../store"
import { getConnectorImage } from "../../../utils/utils"

const EditSourceModal = () => {
	const navigate = useNavigate()
	const {
		showEditSourceModal,
		setShowEditSourceModal,
		showSuccessModal,
		setShowSuccessModal,
		selectedSource,
		updateSource,
		setShowTestingModal,
		setShowFailureModal,
		setSourceTestConnectionError,
	} = useAppStore()

	const getSourceData = () => {
		const configStr =
			typeof selectedSource?.config === "string"
				? selectedSource?.config
				: JSON.stringify(selectedSource?.config)

		const sourceData = {
			name: selectedSource?.name,
			type: selectedSource?.type,
			version: selectedSource?.version,
			config: configStr,
		}
		return sourceData
	}

	const handleEdit = async () => {
		if (!selectedSource?.id) {
			message.error("Source ID is missing")
			return
		}

		try {
			setShowEditSourceModal(false)
			setShowTestingModal(true)
			const testResult =
				await sourceService.testSourceConnection(getSourceData())

			if (testResult.data?.status === "SUCCEEDED") {
				setTimeout(() => {
					setShowTestingModal(false)
					setShowSuccessModal(true)
				}, 1000)

				setTimeout(async () => {
					setShowSuccessModal(false)
					await updateSource(selectedSource.id.toString(), selectedSource)
					navigate("/sources")
				}, 2000)
			} else {
				setShowTestingModal(false)
				setSourceTestConnectionError(testResult.data?.message || "")
				setShowFailureModal(true)
			}
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
							className="size-12 text-primary"
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
						className="bg-blue-600"
					>
						Confirm
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
												? "bg-danger-light text-danger"
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
									<span>
										{text !== undefined
											? formatDistanceToNow(new Date(text), {
													addSuffix: true,
												})
											: "-"}
									</span>
								),
							},
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
						]}
						dataSource={selectedSource?.jobs}
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
						className="size-16 text-success"
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
