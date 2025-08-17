import { useNavigate } from "react-router-dom"
import { Check, GitCommit, Path, LinktreeLogo } from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import { useAppStore } from "../../../store"
import { EntitySavedModalProps } from "../../../types"

const EntitySavedModal: React.FC<EntitySavedModalProps> = ({
	type,
	fromJobFlow,
	entityName = "",
}) => {
	const { showEntitySavedModal, setShowEntitySavedModal } = useAppStore()
	const navigate = useNavigate()
	return (
		<Modal
			open={showEntitySavedModal}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-4 py-4">
				<div className="rounded-xl bg-neutral-light p-2">
					{type === "source" ? (
						<LinktreeLogo className="z-10 size-5 text-text-link" />
					) : type === "config" ? (
						<GitCommit className="z-10 size-5 text-text-link" />
					) : (
						<Path className="z-10 size-5 text-text-link" />
					)}
				</div>
				<div className="mb-4 text-center text-xl font-medium">
					{type === "source"
						? "Source is connected and saved successfully"
						: type === "destination"
							? "Destination is connected and saved successfully"
							: "Your job is created successfully"}
				</div>
				<div className="mb-4 flex w-full items-center justify-between gap-3 rounded-xl border border-[#D9D9D9] px-4 py-2">
					<div className="flex items-center gap-1">
						{type === "source" ? (
							<LinktreeLogo className="size-5" />
						) : type === "config" ? (
							<GitCommit className="size-5" />
						) : (
							<Path className="size-5" />
						)}
						<span>
							{entityName ||
								(type === "source"
									? "Source-Name"
									: type === "config"
										? "Job-Name"
										: "Destination-Name")}
						</span>
					</div>
					<div className="flex gap-1 rounded-xl bg-[#f6ffed] px-2 py-1">
						<Check className="size-5 text-success" />
						<span className="ml-auto text-success">Success</span>
					</div>
				</div>
				<div className="flex space-x-4">
					{type !== "config" && !fromJobFlow && (
						<Button
							type={fromJobFlow ? "primary" : "default"}
							className="border border-[#D9D9D9]"
							onClick={() => {
								setShowEntitySavedModal(false)
								if (type === "source") {
									navigate("/sources")
								} else {
									navigate("/destinations")
								}
							}}
						>
							{type === "source" ? "Sources" : "Destinations"}
						</Button>
					)}
					{!fromJobFlow && (
						<Button
							type="primary"
							onClick={() => {
								setShowEntitySavedModal(false)
								navigate(type === "config" ? "/jobs" : "/jobs/new")
							}}
						>
							{type === "config" ? "Jobs →" : "Create a job →"}
						</Button>
					)}
					{type === "config" && fromJobFlow && (
						<Button
							type="primary"
							onClick={() => {
								setShowEntitySavedModal(false)
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
