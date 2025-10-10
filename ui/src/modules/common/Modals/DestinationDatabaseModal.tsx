import { useState, useEffect } from "react"
import { Modal, Radio, Input, Button, message } from "antd"
import { useAppStore } from "../../../store"
import { validateAlphanumericUnderscore } from "../../../utils/utils"
import {
	DESTINATION_INTERNAL_TYPES,
	FORMAT_OPTIONS,
	LABELS,
	NAMESPACE_PLACEHOLDER,
} from "@utils/constants"
import { DotOutline } from "@phosphor-icons/react"
import { DestinationDatabaseModalProps } from "@app-types/index"

type FormatType = (typeof FORMAT_OPTIONS)[keyof typeof FORMAT_OPTIONS]

import {
	extractDatabasePrefix,
	determineDefaultFormat,
	generateDatabaseNames,
} from "@utils/destination-database"

const DestinationDatabaseModal = ({
	destinationType,
	destinationDatabase,
	allStreams,
	onSave,
	originalDatabase,
	initialStreams,
}: DestinationDatabaseModalProps) => {
	const { showDestinationDatabaseModal, setShowDestinationDatabaseModal } =
		useAppStore()
	const [selectedFormat, setSelectedFormat] = useState<FormatType>(
		FORMAT_OPTIONS.DYNAMIC,
	)
	const [databaseName, setDatabaseName] = useState("")
	const [databaseNameError, setDatabaseNameError] = useState("")

	// Initialize modal state when opened
	useEffect(() => {
		if (showDestinationDatabaseModal) {
			if (destinationDatabase) {
				setDatabaseName(extractDatabasePrefix(destinationDatabase))
			}
			if (originalDatabase) {
				setSelectedFormat(determineDefaultFormat(originalDatabase))
			}
		}
	}, [showDestinationDatabaseModal, destinationDatabase, originalDatabase])

	// Get preview database names
	const previewDatabases = generateDatabaseNames(
		databaseName,
		allStreams,
		initialStreams,
	)

	const handleSaveChanges = () => {
		if (databaseName.trim() === "") {
			message.error(
				`${selectedFormat === FORMAT_OPTIONS.DYNAMIC ? `${labels.title} Prefix` : `${labels.title}`} can not be empty`,
			)
			return
		}
		onSave(selectedFormat, databaseName)
		setShowDestinationDatabaseModal(false)
	}

	const handleClose = () => {
		setShowDestinationDatabaseModal(false)
	}

	const isS3 = destinationType === DESTINATION_INTERNAL_TYPES.S3
	const labels = isS3 ? LABELS.S3 : LABELS.ICEBERG

	return (
		<Modal
			open={showDestinationDatabaseModal}
			footer={null}
			closable={false}
			centered
			width={600}
			title={<div className="text-xl font-semibold">{labels.title}</div>}
		>
			<div className="flex flex-col gap-6 py-4">
				{/* Select Database Format */}
				<div>
					<h3 className="mb-3 text-base font-medium">Select Database Format</h3>
					<Radio.Group
						value={selectedFormat}
						onChange={e => setSelectedFormat(e.target.value)}
						className="w-full"
					>
						<div className="flex gap-8">
							<Radio
								value={FORMAT_OPTIONS.DYNAMIC}
								className="flex flex-1 items-start"
							>
								<div className="ml-2">
									<div className="font-medium">Dynamic (Default)</div>
									<div className="text-sm text-text-sub">
										Your tables will be saved in respective folders
									</div>
								</div>
							</Radio>
							<Radio
								value={FORMAT_OPTIONS.CUSTOM}
								className="flex flex-1 items-start"
							>
								<div className="ml-2">
									<div className="font-medium">Custom</div>
									<div className="text-sm text-text-sub">
										All tables will be stored in a single DB folder
									</div>
								</div>
							</Radio>
						</div>
					</Radio.Group>
				</div>

				{/* Database Name Input */}
				<div>
					<label className="mb-2 block text-base font-medium">
						{selectedFormat === FORMAT_OPTIONS.DYNAMIC
							? `${labels.title} Prefix*`
							: `${labels.title}*`}
					</label>
					<div className="flex gap-2">
						<div className="w-3/5">
							<Input
								placeholder={`Enter your ${labels.title} (a-z, 0-9, _)`}
								value={databaseName}
								onChange={e => {
									const { validValue, errorMessage } =
										validateAlphanumericUnderscore(e.target.value)
									setDatabaseName(validValue)
									setDatabaseNameError(errorMessage)
								}}
								status={databaseNameError ? "error" : undefined}
								className="mb-1"
							/>
							{databaseNameError && (
								<div className="mb-3 text-sm text-red-500">
									{databaseNameError}
								</div>
							)}
						</div>

						{selectedFormat === FORMAT_OPTIONS.DYNAMIC && (
							<span className="w-2/5">+{NAMESPACE_PLACEHOLDER}</span>
						)}
					</div>
				</div>

				{/* Preview Message */}
				{selectedFormat === FORMAT_OPTIONS.CUSTOM && databaseName && (
					<div className="rounded-md bg-blue-50 p-3">
						<div className="text-sm">
							All Tables are saved in one {labels.folderType} folder as{" "}
							<span className="font-medium">{databaseName}</span>
						</div>
					</div>
				)}
				{selectedFormat === FORMAT_OPTIONS.DYNAMIC && (
					<div className="rounded-md bg-blue-50 p-3">
						<div className="flex flex-col">
							<span>Tables are saved in {labels.folderType} folders as</span>
							<span className="font-medium">
								{`$\{${labels.title} Database name prefix (taken from above}${NAMESPACE_PLACEHOLDER}`}
							</span>
							<br />
							<span className="font-bold">
								Your {labels.folderType} {isS3 ? "Folders" : "Databases"}
							</span>
							<div className="mt-2 gap-1">
								{previewDatabases.map((db, index) => (
									<div
										key={index}
										className="flex items-center text-sm"
									>
										<DotOutline
											size={24}
											weight="fill"
										/>
										{db}
									</div>
								))}
							</div>
						</div>
					</div>
				)}

				{/* Action Buttons */}
				<div className="flex justify-start gap-3">
					<Button
						type="primary"
						onClick={handleSaveChanges}
					>
						Save Changes
					</Button>
					<Button onClick={handleClose}>Close</Button>
				</div>
			</div>
		</Modal>
	)
}

export default DestinationDatabaseModal
