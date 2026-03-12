import {
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
} from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import React from "react"
import { useNavigate } from "react-router-dom"

interface EntityCancelModalProps {
	open: boolean
	onClose: () => void
	type: string
	navigateTo: string
}

const EntityCancelModal: React.FC<EntityCancelModalProps> = ({
	open,
	onClose,
	type,
	navigateTo,
}) => {
	const navigate = useNavigate()
	return (
		<Modal
			open={open}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-6 py-4">
				<div className="rounded-xl bg-neutral-light p-2">
					{type === "source" ? (
						<LinktreeLogoIcon className="z-10 size-6 text-text-link" />
					) : type === "destination" ? (
						<PathIcon className="z-10 size-6 text-text-link" />
					) : (
						<GitCommitIcon className="z-10 size-6 text-text-link" />
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
						onClick={onClose}
					>
						Don&apos;t cancel
					</Button>
					<Button
						className="px-8 py-4"
						type="primary"
						danger
						onClick={() => {
							onClose()
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
