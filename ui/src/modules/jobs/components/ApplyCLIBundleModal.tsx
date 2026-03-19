import { ChangeEvent, useEffect, useRef, useState } from "react"
import {
	Alert,
	Button,
	Descriptions,
	Modal,
	Tag,
	Typography,
	message,
} from "antd"
import {
	FileArrowUpIcon,
	FolderOpenIcon,
	FilesIcon,
	TrashIcon,
} from "@phosphor-icons/react"

import { jobService } from "../../../api"
import { ApplyCLIBundleResponse, ApplyPlanAction } from "../../../types"
import {
	REQUIRED_BUNDLE_FILES,
	StagedCLIBundle,
	materializeBundleFile,
	mergeStagedBundles,
	stageArchiveFiles,
	stageStructuredFiles,
} from "./cliBundleFiles"

interface ApplyCLIBundleModalProps {
	open: boolean
	onClose: () => void
	onApplied: () => Promise<void> | void
}

interface BundlePreviewResult {
	bundle: StagedCLIBundle
	response?: ApplyCLIBundleResponse
	error?: string
}

const actionColor: Record<ApplyPlanAction, string> = {
	created: "green",
	updated: "blue",
	unchanged: "default",
	preserved: "gold",
}

const getBundleStatus = (bundle: StagedCLIBundle) => {
	if (bundle.duplicateFiles.length > 0) {
		return {
			label: "invalid",
			color: "red",
		}
	}

	if (bundle.missingRequiredFiles.length > 0) {
		return {
			label: "incomplete",
			color: "gold",
		}
	}

	return {
		label: "ready",
		color: "green",
	}
}

const ApplyCLIBundleModal: React.FC<ApplyCLIBundleModalProps> = ({
	open,
	onClose,
	onApplied,
}) => {
	const archiveInputRef = useRef<HTMLInputElement | null>(null)
	const jsonInputRef = useRef<HTMLInputElement | null>(null)
	const folderInputRef = useRef<HTMLInputElement | null>(null)
	const [stagedBundles, setStagedBundles] = useState<StagedCLIBundle[]>([])
	const [previewResults, setPreviewResults] = useState<BundlePreviewResult[]>([])
	const [previewLoading, setPreviewLoading] = useState(false)
	const [applyLoading, setApplyLoading] = useState(false)

	useEffect(() => {
		if (!open) {
			setStagedBundles([])
			setPreviewResults([])
			setPreviewLoading(false)
			setApplyLoading(false)
		}
	}, [open])

	const previewableBundles = stagedBundles.filter(
		bundle =>
			bundle.duplicateFiles.length === 0 &&
			bundle.missingRequiredFiles.length === 0,
	)
	const hasInvalidBundles = previewableBundles.length !== stagedBundles.length

	const addBundles = (incomingBundles: StagedCLIBundle[]) => {
		if (incomingBundles.length === 0) {
			message.warning("No supported bundle files were selected")
			return
		}

		setStagedBundles(currentBundles =>
			mergeStagedBundles(currentBundles, incomingBundles),
		)
		setPreviewResults([])
	}

	const handleArchiveSelection = (event: ChangeEvent<HTMLInputElement>) => {
		const selectedFiles = Array.from(event.target.files || [])
		addBundles(stageArchiveFiles(selectedFiles))
		event.target.value = ""
	}

	const handleJSONSelection = (event: ChangeEvent<HTMLInputElement>) => {
		const selectedFiles = Array.from(event.target.files || [])
		addBundles(stageStructuredFiles(selectedFiles, "json-files"))
		event.target.value = ""
	}

	const handleFolderSelection = (event: ChangeEvent<HTMLInputElement>) => {
		const selectedFiles = Array.from(event.target.files || [])
		addBundles(stageStructuredFiles(selectedFiles, "folder"))
		event.target.value = ""
	}

	const removeBundle = (bundleId: string) => {
		setStagedBundles(currentBundles =>
			currentBundles.filter(bundle => bundle.id !== bundleId),
		)
		setPreviewResults(currentResults =>
			currentResults.filter(result => result.bundle.id !== bundleId),
		)
	}

	const clearBundles = () => {
		setStagedBundles([])
		setPreviewResults([])
	}

	const runBundleAction = async (
		action: "preview" | "apply",
		setLoading: (value: boolean) => void,
	) => {
		if (stagedBundles.length === 0) {
			message.warning("Select at least one bundle, folder, or JSON set first")
			return
		}

		if (previewableBundles.length === 0) {
			message.warning("Remove incomplete bundles before continuing")
			return
		}

		if (hasInvalidBundles) {
			message.warning("Only ready bundles can be applied")
			return
		}

		try {
			setLoading(true)
			const nextResults: BundlePreviewResult[] = []

			for (const bundle of previewableBundles) {
				try {
					const bundleFile = await materializeBundleFile(bundle)
					const response =
						action === "preview"
							? await jobService.previewCLIBundleApply(bundleFile)
							: await jobService.applyCLIBundle(bundleFile)
					nextResults.push({
						bundle,
						response,
					})
				} catch (error) {
					nextResults.push({
						bundle,
						error:
							error instanceof Error ? error.message : "Failed to process bundle",
					})
				}
			}

			setPreviewResults(nextResults)

			if (action === "apply") {
				const appliedCount = nextResults.filter(result => result.response).length
				if (appliedCount > 0) {
					await onApplied()
				}

				if (appliedCount === previewableBundles.length) {
					message.success(
						`Applied ${appliedCount} bundle${appliedCount === 1 ? "" : "s"}`,
					)
				} else {
					message.warning(
						`Applied ${appliedCount} of ${previewableBundles.length} bundles`,
					)
				}
			}
		} finally {
			setLoading(false)
		}
	}

	const footerActionLabel =
		previewableBundles.length === 1 ? "Bundle" : "Bundles"

	return (
		<Modal
			open={open}
			onCancel={onClose}
			title="Apply CLI Bundles"
			width={960}
			footer={[
				<Button
					key="close"
					onClick={onClose}
				>
					Close
				</Button>,
				<Button
					key="clear"
					onClick={clearBundles}
					disabled={stagedBundles.length === 0}
				>
					Clear Selection
				</Button>,
				<Button
					key="preview"
					onClick={() => runBundleAction("preview", setPreviewLoading)}
					loading={previewLoading}
					disabled={previewableBundles.length === 0 || hasInvalidBundles}
				>
					Preview {footerActionLabel}
				</Button>,
				<Button
					key="apply"
					type="primary"
					onClick={() => runBundleAction("apply", setApplyLoading)}
					loading={applyLoading}
					disabled={previewableBundles.length === 0 || hasInvalidBundles}
				>
					Apply {footerActionLabel}
				</Button>,
			]}
		>
			<div className="flex flex-col gap-4">
				<Alert
					type="info"
					showIcon
					message="Each bundle must contain source.json, destination.json, and streams.json. Add olake-ui.json when connector metadata must be preserved. You can add folders repeatedly to queue batch imports."
				/>

				<div className="rounded-md border border-dashed border-gray-300 p-4">
					<div className="flex flex-wrap gap-3">
						<Button
							icon={<FolderOpenIcon className="size-4" />}
							onClick={() => folderInputRef.current?.click()}
							data-testid="cli-bundle-add-folder"
						>
							Add Folder
						</Button>
						<Button
							icon={<FilesIcon className="size-4" />}
							onClick={() => jsonInputRef.current?.click()}
							data-testid="cli-bundle-add-json"
						>
							Add JSON Files
						</Button>
						<Button
							icon={<FileArrowUpIcon className="size-4" />}
							onClick={() => archiveInputRef.current?.click()}
							data-testid="cli-bundle-add-archive"
						>
							Add Bundle Archives
						</Button>
					</div>

					<Typography.Text
						type="secondary"
						className="mt-3 block"
					>
						Folder import can be repeated to build a batch. If you select a
						parent directory with multiple job folders, the modal stages each
						subfolder separately.
					</Typography.Text>

					<input
						ref={archiveInputRef}
						type="file"
						accept=".zip,.tar.gz,.tgz"
						multiple
						className="hidden"
						onChange={handleArchiveSelection}
						data-testid="cli-bundle-archive-input"
					/>
					<input
						ref={jsonInputRef}
						type="file"
						accept=".json"
						multiple
						className="hidden"
						onChange={handleJSONSelection}
						data-testid="cli-bundle-json-input"
					/>
					<input
						ref={input => {
							folderInputRef.current = input
							if (input) {
								input.setAttribute("webkitdirectory", "")
								input.setAttribute("directory", "")
							}
						}}
						type="file"
						multiple
						className="hidden"
						onChange={handleFolderSelection}
						data-testid="cli-bundle-folder-input"
					/>
				</div>

				<div className="flex items-center justify-between gap-3">
					<Typography.Text strong>
						Queued Imports ({stagedBundles.length})
					</Typography.Text>
					{hasInvalidBundles && stagedBundles.length > 0 && (
						<Typography.Text type="warning">
							Remove incomplete bundles before previewing or applying them.
						</Typography.Text>
					)}
				</div>

				{stagedBundles.length === 0 ? (
					<div className="rounded-md border border-gray-200 p-4">
						<Typography.Text type="secondary">
							No bundles queued yet.
						</Typography.Text>
					</div>
				) : (
					<div className="flex flex-col gap-3">
						{stagedBundles.map(bundle => {
							const status = getBundleStatus(bundle)
							return (
								<div
									key={bundle.id}
									className="rounded-md border border-gray-200 p-4"
									data-testid={`cli-bundle-stage-${bundle.archiveRoot}`}
								>
									<div className="mb-3 flex items-start justify-between gap-3">
										<div>
											<div className="flex items-center gap-2">
												<Typography.Text strong>
													{bundle.name}
												</Typography.Text>
												<Tag color={status.color}>{status.label}</Tag>
											</div>
											<Typography.Text type="secondary">
												{bundle.sourceLabel}: {bundle.displayName}
											</Typography.Text>
										</div>
										<Button
											type="text"
											icon={<TrashIcon className="size-4" />}
											onClick={() => removeBundle(bundle.id)}
											aria-label={`Remove ${bundle.name}`}
										/>
									</div>

									<div className="grid gap-3 md:grid-cols-2">
										<Descriptions
											column={1}
											size="small"
											bordered
										>
											<Descriptions.Item label="Detected files">
												{bundle.fileNames.length > 0
													? bundle.fileNames.join(", ")
													: "None"}
											</Descriptions.Item>
											<Descriptions.Item label="Required files">
												{REQUIRED_BUNDLE_FILES.join(", ")}
											</Descriptions.Item>
											<Descriptions.Item label="Ignored files">
												{bundle.ignoredFileCount}
											</Descriptions.Item>
										</Descriptions>

										<div className="flex flex-col gap-2">
											{bundle.missingRequiredFiles.length > 0 && (
												<Alert
													type="warning"
													showIcon
													message={`Missing required files: ${bundle.missingRequiredFiles.join(", ")}`}
												/>
											)}
											{bundle.duplicateFiles.length > 0 && (
												<Alert
													type="error"
													showIcon
													message={`Duplicate files detected: ${bundle.duplicateFiles.join(", ")}`}
												/>
											)}
											{bundle.missingRequiredFiles.length === 0 &&
												bundle.duplicateFiles.length === 0 && (
													<Alert
														type="success"
														showIcon
														message="This bundle is ready to preview or apply."
													/>
												)}
										</div>
									</div>
								</div>
							)
						})}
					</div>
				)}

				{previewResults.length > 0 && (
					<div className="flex flex-col gap-4">
						<Typography.Text strong>Preview Results</Typography.Text>

						{previewResults.map(result => (
							<div
								key={result.bundle.id}
								className="rounded-md border border-gray-200 p-4"
								data-testid={`cli-bundle-preview-${result.bundle.archiveRoot}`}
							>
								<div className="mb-3 flex items-center justify-between gap-3">
									<Typography.Text strong>{result.bundle.name}</Typography.Text>
									<Tag color={result.error ? "red" : "blue"}>
										{result.error ? "error" : "preview"}
									</Tag>
								</div>

								{result.error ? (
									<Alert
										type="error"
										showIcon
										message={result.error}
									/>
								) : (
									result.response && (
										<div className="flex flex-col gap-4">
											<Descriptions
												column={2}
												size="small"
												bordered
											>
												<Descriptions.Item label="Bundle">
													{result.response.bundle}
												</Descriptions.Item>
												<Descriptions.Item label="Apply Identity">
													{result.response.effective.apply_identity}
												</Descriptions.Item>
												<Descriptions.Item label="Job">
													{result.response.effective.job_name}
												</Descriptions.Item>
												<Descriptions.Item label="Frequency">
													{result.response.effective.frequency || "manual only"}
												</Descriptions.Item>
												<Descriptions.Item label="Source">
													{result.response.effective.source_name}
												</Descriptions.Item>
												<Descriptions.Item label="Destination">
													{result.response.effective.destination_name}
												</Descriptions.Item>
											</Descriptions>

											<div className="grid gap-3 md:grid-cols-2">
												<PlanCard
													title="Source"
													action={result.response.source.action}
													name={result.response.source.name}
													fields={result.response.source.fields}
												/>
												<PlanCard
													title="Destination"
													action={result.response.destination.action}
													name={result.response.destination.name}
													fields={result.response.destination.fields}
												/>
												<PlanCard
													title="Job"
													action={result.response.job.action}
													name={result.response.job.name}
													fields={result.response.job.fields}
												/>
												<PlanCard
													title="State"
													action={result.response.state.action}
													name="state.json"
													fields={result.response.state.fields}
												/>
											</div>
										</div>
									)
								)}
							</div>
						))}
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
