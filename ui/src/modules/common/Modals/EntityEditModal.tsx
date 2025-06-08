import { Button, Modal, Table } from "antd"
import { useAppStore } from "../../../store"
import { getConnectorImage } from "../../../utils/utils"
import { CheckCircle, Warning } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"
import { message } from "antd"
import { formatDistanceToNow } from "date-fns"
import { sourceService } from "../../../api"
import { destinationService } from "../../../api/services/destinationService"
import { EntityEditModalProps } from "../../../types"

const EntityEditModal = ({ entityType }: EntityEditModalProps) => {
	const navigate = useNavigate()
	const {
		showEditSourceModal,
		setShowEditSourceModal,
		showEditDestinationModal,
		setShowEditDestinationModal,
		showSuccessModal,
		setShowSuccessModal,
		selectedSource,
		selectedDestination,
		updateSource,
		updateDestination,
		setShowTestingModal,
		setShowFailureModal,
		setSourceTestConnectionError,
		setDestinationTestConnectionError,
	} = useAppStore()

	const isSource = entityType === "source"
	const showModal = isSource ? showEditSourceModal : showEditDestinationModal
	const setShowModal = isSource
		? setShowEditSourceModal
		: setShowEditDestinationModal
	const selectedEntity = isSource ? selectedSource : selectedDestination
	const updateEntity = isSource ? updateSource : updateDestination
	const setTestConnectionError = isSource
		? setSourceTestConnectionError
		: setDestinationTestConnectionError
	const navigatePath = isSource ? "/sources" : "/destinations"

	const getEntityData = () => {
		const configStr =
			typeof selectedEntity?.config === "string"
				? selectedEntity?.config
				: JSON.stringify(selectedEntity?.config)

		return {
			name: selectedEntity?.name,
			type: selectedEntity?.type,
			version: selectedEntity?.version,
			config: configStr,
		}
	}

	const handleEdit = async () => {
		if (!selectedEntity?.id) {
			message.error(
				`${entityType.charAt(0).toUpperCase() + entityType.slice(1)} ID is missing`,
			)
			return
		}

		try {
			setShowModal(false)
			setShowTestingModal(true)

			const testResult = isSource
				? await sourceService.testSourceConnection(getEntityData())
				: await destinationService.testDestinationConnection(getEntityData())

			if (testResult.data?.status === "SUCCEEDED") {
				setTimeout(() => {
					setShowTestingModal(false)
					setShowSuccessModal(true)
				}, 1000)

				setTimeout(async () => {
					setShowSuccessModal(false)
					await updateEntity(selectedEntity.id.toString(), selectedEntity)
					navigate(navigatePath)
				}, 2000)
			} else {
				setShowTestingModal(false)
				setTestConnectionError(testResult.data?.message || "")
				setShowFailureModal(true)
			}
		} catch (error) {
			message.error(`Failed to update ${entityType}`)
			console.error(error)
		}
	}

	const getTableColumns = () => {
		const commonColumns = [
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
		]

		const entitySpecificColumn = {
			title: isSource ? "Destination" : "Source",
			dataIndex: isSource ? "destination_name" : "source_name",
			key: isSource ? "destination_name" : "source_name",
			render: (name: string, record: any) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(
							record[isSource ? "destination_type" : "source_type"] || "",
						)}
						alt={record[isSource ? "destination_type" : "source_type"]}
						className="mr-2 size-6"
					/>
					{name || "N/A"}
				</div>
			),
		}

		return [...commonColumns, entitySpecificColumn]
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
				open={showModal}
				onCancel={() => setShowModal(false)}
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
						onClick={() => setShowModal(false)}
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
						Editing this {entityType} will affect the following jobs that are
						associated with this {entityType} and as a result will fail
						immediately. Do you still want to edit the {entityType}?
					</p>
				</div>
				<div className="mt-6">
					<Table
						columns={getTableColumns()}
						dataSource={selectedEntity?.jobs}
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

export default EntityEditModal
