import { Button, Modal } from "antd"
import { useAppStore } from "../../../store"
import { GitCommit, LinktreeLogo, Path } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"
import React from "react"

interface EntityCancelModalProps {
	type: string
	navigateTo: string
}

const EntityCancelModal: React.FC<EntityCancelModalProps> = ({
	type,
	navigateTo,
}) => {
	const { showSourceCancelModal, setShowSourceCancelModal } = useAppStore()
	const navigate = useNavigate()
	return (
		<Modal
			open={showSourceCancelModal}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-6 py-4">
				<div className="rounded-xl bg-[#F0F0F0] p-2">
					{type === "source" ? (
						<LinktreeLogo className="z-10 size-6 text-[#6E6E6E]" />
					) : type === "destination" ? (
						<Path className="z-10 size-6 text-[#6E6E6E]" />
					) : (
						<GitCommit className="z-10 size-6 text-[#6E6E6E]" />
					)}
				</div>
				<div className="mb-4 text-center text-xl font-medium">
					{type === "job"
						? "Are you sure you want to cancel the job?"
						: type === "job-edit"
							? "Are you sure you want to cancel the job edit?"
							: type === "source"
								? "Are you sure you want to cancel the source?"
								: "Are you sure you want to cancel the destination?"}
				</div>
				<div className="flex space-x-8">
					<Button
						className="border border-[#D9D9D9]"
						onClick={() => setShowSourceCancelModal(false)}
					>
						Don&apos;t cancel
					</Button>
					<Button
						className="px-8 py-4"
						type="primary"
						danger
						onClick={() => {
							setShowSourceCancelModal(false)
							navigate(`/${navigateTo}`)
						}}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default EntityCancelModal
