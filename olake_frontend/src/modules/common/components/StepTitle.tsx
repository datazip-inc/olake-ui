const StepTitle = ({
	stepNumber,
	stepTitle,
}: {
	stepNumber: number | string
	stepTitle: string
}) => {
	return (
		<div className="mb-4 flex flex-col gap-2">
			<div className="flex items-center gap-2">
				<div className="size-2 rounded-full border border-[#203FDD] outline outline-2 outline-[#203FDD]"></div>
				<span className="text-xs text-[#8A8A8A]">Step {stepNumber}</span>
			</div>
			<h1 className="text-xl font-medium">{stepTitle}</h1>
		</div>
	)
}

export default StepTitle
