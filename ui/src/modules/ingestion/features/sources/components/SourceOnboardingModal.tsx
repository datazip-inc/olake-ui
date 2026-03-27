import {
	ArrowRightIcon,
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
	XIcon,
} from "@phosphor-icons/react"
import type { ReactNode } from "react"

const StepCard = ({
	step,
	title,
	description,
	icon,
	iconContainerClassName,
}: {
	step: string
	title: string
	description: string
	icon: ReactNode
	iconContainerClassName: string
}) => (
	<div className="h-24 w-full rounded-2xl border border-olake-border bg-white p-[7px]">
		<div className="flex h-full items-center gap-5">
			<div
				className={`flex size-20 shrink-0 items-center justify-center rounded-xl ${iconContainerClassName}`}
			>
				{icon}
			</div>
			<div className="flex min-w-0 flex-col gap-1">
				<p className="text-xs font-medium tracking-wide text-text-placeholder">
					STEP {step}
				</p>
				<p className="text-[18px] font-medium leading-[22px] text-black">
					{title}
				</p>
				<p className="text-sm leading-[normal] text-text-primary">
					{description}
				</p>
			</div>
		</div>
	</div>
)

const SourceOnboardingModal = ({
	open,
	handleCreateSource,
	onClose,
}: {
	open: boolean
	handleCreateSource: () => void
	onClose: () => void
}) => {
	if (!open) {
		return null
	}

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center bg-[rgba(47,47,47,0.33)] p-4">
			<div className="relative h-[670px] w-full max-w-[894px] rounded-2xl bg-white">
				<button
					type="button"
					onClick={onClose}
					aria-label="Close onboarding modal"
					className="absolute right-6 top-6 flex size-8 items-center justify-center rounded-md text-olake-icon-muted transition-colors hover:bg-olake-surface-muted hover:text-olake-body"
				>
					<XIcon size={18} />
				</button>
				<div className="px-10 pt-12">
					<div className="w-[540px]">
						<p className="text-sm font-medium text-brand-blue">Hello User !</p>
						<div className="mt-3 flex flex-col gap-1">
							<h2 className="text-2xl font-bold leading-8 text-text-primary">
								Welcome to OLake
							</h2>
							<p className="text-base leading-[normal] text-text-primary">
								Get started by following these simple steps to set up your first
								data pipeline
							</p>
						</div>
					</div>

					<div className="mt-12 flex w-full max-w-[776px] flex-col gap-7">
						<StepCard
							step="I"
							title="Create a Source"
							description="Connect to your database (MongoDB, PostgreSQL, MySQL, etc.) by providing connection details."
							icon={
								<LinktreeLogoIcon
									size={40}
									weight="regular"
									className="text-[#7E48FF]"
								/>
							}
							iconContainerClassName="bg-[#EEE9FF]"
						/>
						<StepCard
							step="II"
							title="Create a Destination"
							description="Set up where your data will be stored (Amazon S3, Apache Iceberg)."
							icon={
								<PathIcon
									size={36}
									weight="regular"
									className="text-[#3AA4E3]"
								/>
							}
							iconContainerClassName="bg-[#EAF7FF]"
						/>
						<StepCard
							step="III"
							title="Setup & Run a Job"
							description="Define a sync job to transfer data from your source to destination."
							icon={
								<GitCommitIcon
									size={34}
									weight="regular"
									className="text-[#335CFF]"
								/>
							}
							iconContainerClassName="bg-[#EEF0FA]"
						/>
					</div>

					<div className="mt-8 w-[242px]">
						<p className="text-xs font-medium text-[#212121]">
							What&apos;s Next: Create your first source
						</p>
						<button
							type="button"
							data-testid="onboarding-create-source-button"
							onClick={handleCreateSource}
							className="mt-4 flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-brand-blue px-4 text-base text-white shadow-[0px_2px_0px_0px_rgba(5,145,255,0.01)] transition-opacity hover:opacity-95"
						>
							<span>Create Source</span>
							<ArrowRightIcon
								size={20}
								weight="regular"
							/>
						</button>
					</div>
				</div>
			</div>
		</div>
	)
}

export default SourceOnboardingModal
