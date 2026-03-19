import { useEffect, useState } from "react"
import {
	Alert,
	Button,
	Descriptions,
	Modal,
	Tag,
	Typography,
	Upload,
	message,
} from "antd"
import type { UploadFile, UploadProps } from "antd"
import { UploadSimpleIcon } from "@phosphor-icons/react"

import { jobService } from "../../../api"
import { ApplyCLIBundleResponse, ApplyPlanAction } from "../../../types"

interface ApplyCLIBundleModalProps {
	open: boolean
	onClose: () => void
	onApplied: () => Promise<void> | void
}

const actionColor: Record<ApplyPlanAction, string> = {
	created: "green",
	updated: "blue",
	unchanged: "default",
	preserved: "gold",
}

const ApplyCLIBundleModal: React.FC<ApplyCLIBundleModalProps> = ({
	open,
	onClose,
	onApplied,
}) => {
	const [selectedFile, setSelectedFile] = useState<File | null>(null)
	const [preview, setPreview] = useState<ApplyCLIBundleResponse | null>(null)
	const [previewLoading, setPreviewLoading] = useState(false)
	const [applyLoading, setApplyLoading] = useState(false)

	useEffect(() => {
		if (!open) {
			setSelectedFile(null)
			setPreview(null)
			setPreviewLoading(false)
			setApplyLoading(false)
		}
	}, [open])

	const uploadProps: UploadProps = {
		maxCount: 1,
		accept: ".zip,.tar.gz,.tgz",
		beforeUpload: file => {
			setSelectedFile(file)
			setPreview(null)
			return false
		},
		onRemove: () => {
			setSelectedFile(null)
			setPreview(null)
			return true
		},
		fileList: selectedFile
			? [
					{
						uid: selectedFile.name,
						name: selectedFile.name,
						status: "done",
					} satisfies UploadFile,
			  ]
			: [],
	}

	const handlePreview = async () => {
		if (!selectedFile) {
			message.warning("Select a bundle first")
			return
		}

		try {
			setPreviewLoading(true)
			const response = await jobService.previewCLIBundleApply(selectedFile)
			setPreview(response)
		} finally {
			setPreviewLoading(false)
		}
	}

	const handleApply = async () => {
		if (!selectedFile) {
			message.warning("Select a bundle first")
			return
		}

		try {
			setApplyLoading(true)
			const response = await jobService.applyCLIBundle(selectedFile)
			setPreview(response)
			await onApplied()
		} finally {
			setApplyLoading(false)
		}
	}

	return (
		<Modal
			open={open}
			onCancel={onClose}
			title="Apply CLI Bundle"
			width={840}
			footer={[
				<Button
					key="close"
					onClick={onClose}
				>
					Close
				</Button>,
				<Button
					key="preview"
					onClick={handlePreview}
					loading={previewLoading}
					disabled={!selectedFile}
				>
					Preview Apply
				</Button>,
				<Button
					key="apply"
					type="primary"
					onClick={handleApply}
					loading={applyLoading}
					disabled={!selectedFile}
				>
					Apply Bundle
				</Button>,
			]}
		>
			<div className="flex flex-col gap-4">
				<Alert
					type="info"
					showIcon
					message="The bundle must contain source.json, destination.json, and streams.json. Add olake-ui.json when the server cannot infer connector metadata such as source_type, source_version, and destination_version."
				/>

				<Upload.Dragger {...uploadProps}>
					<div className="flex flex-col items-center gap-2 py-4">
						<UploadSimpleIcon className="size-8 text-primary" />
						<Typography.Text strong>
							Drop a CLI bundle here or click to select one
						</Typography.Text>
						<Typography.Text type="secondary">
							Supported formats: .zip, .tar.gz, .tgz
						</Typography.Text>
					</div>
				</Upload.Dragger>

				{preview && (
					<div className="flex flex-col gap-4">
						<Descriptions
							column={2}
							size="small"
							bordered
						>
							<Descriptions.Item label="Bundle">
								{preview.bundle}
							</Descriptions.Item>
							<Descriptions.Item label="Apply Identity">
								{preview.effective.apply_identity}
							</Descriptions.Item>
							<Descriptions.Item label="Job">
								{preview.effective.job_name}
							</Descriptions.Item>
							<Descriptions.Item label="Frequency">
								{preview.effective.frequency || "manual only"}
							</Descriptions.Item>
							<Descriptions.Item label="Source">
								{preview.effective.source_name}
							</Descriptions.Item>
							<Descriptions.Item label="Destination">
								{preview.effective.destination_name}
							</Descriptions.Item>
						</Descriptions>

						<div className="grid gap-3 md:grid-cols-2">
							<PlanCard
								title="Source"
								action={preview.source.action}
								name={preview.source.name}
								fields={preview.source.fields}
							/>
							<PlanCard
								title="Destination"
								action={preview.destination.action}
								name={preview.destination.name}
								fields={preview.destination.fields}
							/>
							<PlanCard
								title="Job"
								action={preview.job.action}
								name={preview.job.name}
								fields={preview.job.fields}
							/>
							<PlanCard
								title="State"
								action={preview.state.action}
								name="state.json"
								fields={preview.state.fields}
							/>
						</div>
					</div>
				)}
			</div>
		</Modal>
	)
}

const PlanCard = ({
	title,
	action,
	name,
	fields,
}: {
	title: string
	action: ApplyPlanAction
	name: string
	fields?: string[]
}) => {
	return (
		<div className="rounded-md border border-gray-200 p-4">
			<div className="mb-2 flex items-center justify-between gap-2">
				<Typography.Text strong>{title}</Typography.Text>
				<Tag color={actionColor[action]}>{action}</Tag>
			</div>
			<Typography.Text className="block">{name}</Typography.Text>
			<Typography.Text
				type="secondary"
				className="mt-2 block"
			>
				{fields && fields.length > 0
					? `Fields: ${fields.join(", ")}`
					: "No changes"}
			</Typography.Text>
		</div>
	)
}

export default ApplyCLIBundleModal
