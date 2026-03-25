import {
	CheckIcon,
	GitCommitIcon,
	PathIcon,
	LinktreeLogoIcon,
} from "@phosphor-icons/react"
import { Button, Modal, Tooltip } from "antd"
import { useNavigate } from "react-router-dom"

import { EntitySavedModalProps } from "@/modules/ingestion/common/types"

const EntitySavedModal: React.FC<EntitySavedModalProps> = ({
	open,
	onClose,
	type,
	fromJobFlow,
	entityName = "",
}) => {
	const navigate = useNavigate()

	const isSource = type === "source"
	const isDestination = type === "destination"
	const isJob = type === "streams"
	const displayEntityName =
		entityName ||
		(isSource ? "Source-Name" : isJob ? "Job-Name" : "Destination-Name")

	return (
		<Modal
			open={open}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-4 py-4">
				<div className="rounded-xl bg-neutral-light p-2">
					{isSource ? (
						<LinktreeLogoIcon className="z-10 size-5 text-text-link" />
					) : isJob ? (
						<GitCommitIcon className="z-10 size-5 text-text-link" />
					) : (
						<PathIcon className="z-10 size-5 text-text-link" />
					)}
				</div>
				<div className="mb-4 text-center text-xl font-medium">
					{isSource
						? "Source is connected and saved successfully"
						: isDestination
							? "Destination is connected and saved successfully"
							: "Your job is created successfully"}
				</div>
				<div className="mb-4 flex w-full items-center justify-between gap-3 rounded-xl border border-[#D9D9D9] px-4 py-2">
					<div className="flex min-w-0 flex-1 items-center gap-1">
						{isSource ? (
							<LinktreeLogoIcon className="size-5" />
						) : isJob ? (
							<GitCommitIcon className="size-5" />
						) : (
							<PathIcon className="size-5" />
						)}
						<Tooltip title={displayEntityName}>
							<span className="block truncate">{displayEntityName}</span>
						</Tooltip>
					</div>
					<div className="flex gap-1 rounded-xl bg-[#f6ffed] px-2 py-1">
						<CheckIcon className="size-5 text-success" />
						<span className="ml-auto text-success">Success</span>
					</div>
				</div>
				<div className="flex space-x-4">
					{!isJob && !fromJobFlow && (
						<Button
							type="default"
							className="border border-[#D9D9D9]"
							onClick={() => {
								onClose()
								if (isSource) {
									navigate("/sources")
								} else {
									navigate("/destinations")
								}
							}}
						>
							{isSource ? "Sources" : "Destinations"}
						</Button>
					)}
					{!fromJobFlow && (
						<Button
							type="primary"
							onClick={() => {
								onClose()
								if (isSource) {
									navigate("/destinations/new")
								} else if (isDestination) {
									navigate("/jobs/new")
								} else {
									navigate("/jobs")
								}
							}}
						>
							{isSource
								? "Destinations →"
								: isDestination
									? "Create a job →"
									: "Jobs →"}
						</Button>
					)}
					{isJob && fromJobFlow && (
						<Button
							type="primary"
							onClick={() => {
								onClose()
								navigate("/jobs")
							}}
						>
							Jobs →
						</Button>
					)}
				</div>
			</div>
		</Modal>
	)
}

export default EntitySavedModal
