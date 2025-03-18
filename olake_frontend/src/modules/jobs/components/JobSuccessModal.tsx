import { Button, Modal } from "antd"
import { useAppStore } from "../../../store"
import { ArrowRight, Check, GitCommit } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"

const JobSuccessModal: React.FC = () => {
	const { showJobSuccessModal, setShowJobSuccessModal } = useAppStore()
	const navigate = useNavigate()
	return (
		<Modal
			open={showJobSuccessModal}
			footer={null}
			closable={false}
			centered
			width={450}
			height={290}
		>
			<div className="flex flex-col items-center justify-center gap-3 p-3">
				<div className="flex w-[90%] flex-col items-center gap-3">
					<div className="rounded-xl bg-[#F0F0F0] p-2">
						<GitCommit className="z-10 size-8 text-[#000]" />
					</div>
					<div className="mb-4 text-center text-xl font-medium">
						Your job is running successfully
					</div>
					<div className="mb-4 flex w-full items-center justify-between gap-3 rounded-xl border border-[#D9D9D9] px-2 py-2">
						<div className="flex items-center gap-2">
							<GitCommit className="size-5" />
							<span className="font-bold">&lt;Job-Name&gt;</span>
						</div>
						<div className="flex gap-1 rounded-xl bg-[#F6FFED] px-2 py-1">
							<Check className="size-5 text-[#389E0D]" />
							<span className="ml-auto text-[#389E0D]">Success</span>
						</div>
					</div>
				</div>
				<div className="flex items-center space-x-4">
					<Button
						type="primary"
						onClick={() => {
							setShowJobSuccessModal(false)
							navigate("/jobs")
						}}
						className="flex items-center border-[#203FDD] bg-[##203FDD] px-8 py-4 text-sm hover:bg-[#203FDD]"
					>
						<span>Jobs</span>
						<ArrowRight className="size-4" />
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default JobSuccessModal
