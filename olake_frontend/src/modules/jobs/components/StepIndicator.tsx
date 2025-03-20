interface StepIndicatorProps {
	step: string
	index: number
	currentStep: string
}

const steps: string[] = ["source", "destination", "schema", "config"]

const StepIndicator: React.FC<StepIndicatorProps> = ({
	step,
	index,
	currentStep,
}) => {
	const isActive = steps.indexOf(currentStep) >= index
	const isNextActive = steps.indexOf(currentStep) >= index + 1
	const isLastStep = index === steps.length - 1

	return (
		<div className="flex flex-col items-start">
			<div className="flex items-center">
				<div
					className={`z-10 size-3 rounded-full border ${
						isActive
							? "border-blue-600 outline outline-2 outline-blue-600"
							: "border-gray-300 bg-white"
					}`}
				></div>
				{!isLastStep && (
					<div className="relative h-[2px] w-20">
						<div className="absolute inset-0 bg-gray-300"></div>
						{isNextActive && (
							<div className="absolute inset-0 bg-blue-600 transition-all duration-300" />
						)}
					</div>
				)}
			</div>
			<span
				className={`mt-2 translate-x-[-40%] text-xs ${
					isActive ? "text-blue-600" : "text-gray-500"
				}`}
			>
				{step === "config"
					? "Job Config"
					: step.charAt(0).toUpperCase() + step.slice(1)}
			</span>
		</div>
	)
}

interface StepProgressProps {
	currentStep: string
}

const StepProgress: React.FC<StepProgressProps> = ({ currentStep }) => {
	return (
		<div className="flex items-center">
			{steps.map((step, index) => (
				<StepIndicator
					key={step}
					step={step}
					index={index}
					currentStep={currentStep}
				/>
			))}
		</div>
	)
}

export default StepProgress
