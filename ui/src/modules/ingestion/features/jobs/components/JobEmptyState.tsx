import {
	LinktreeLogoIcon,
	ArrowRightIcon,
	GitCommitIcon,
	PathIcon,
} from "@phosphor-icons/react"
import type { ReactNode } from "react"
import { useNavigate } from "react-router-dom"

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
	<div className="h-24 w-full rounded-2xl border border-[#d9d9d9] bg-white p-[7px]">
		<div className="flex h-full items-center gap-5">
			<div
				className={`flex size-20 shrink-0 items-center justify-center rounded-xl ${iconContainerClassName}`}
			>
				{icon}
			</div>
			<div className="flex min-w-0 flex-col gap-1">
				<div className="text-xs font-medium tracking-wide text-[#9f9f9f]">
					STEP {step}
				</div>
				<h3 className="text-[18px] font-medium leading-[1.15] text-black">
					{title}
				</h3>
				<p className="text-sm text-[#0a0a0a]">{description}</p>
			</div>
		</div>
	</div>
)

const JobEmptyState = () => {
	const navigate = useNavigate()
	return (
		<div className="mx-auto flex w-full max-w-[742px] flex-col items-center py-16">
			<div className="flex flex-col items-center gap-1 text-center">
				<p className="text-sm font-medium text-brand-blue">Hello User !</p>
				<h2 className="text-2xl font-bold leading-8 text-[#0a0a0a]">
					Welcome to OLake
				</h2>
				<p className="text-base leading-6 text-[#0a0a0a]">
					Get started by following these simple steps to set up your first data
					pipeline
				</p>
			</div>

			<div className="mt-8 flex w-full flex-col gap-4">
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

			<div className="mt-7 flex flex-col items-center gap-4">
				<p className="text-xs font-medium text-[#212121]">
					What&apos;s Next: Create your first source
				</p>
				<button
					onClick={() => {
						navigate("/sources/new")
					}}
					className="flex h-10 min-w-[172px] items-center justify-center gap-2 rounded-lg bg-brand-blue px-6 text-base text-white transition-opacity hover:opacity-95"
				>
					<span>Create Source</span>
					<ArrowRightIcon
						size={16}
						weight="bold"
					/>
				</button>
			</div>
		</div>
	)
}

export default JobEmptyState
