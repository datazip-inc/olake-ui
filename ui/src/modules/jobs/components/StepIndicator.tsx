import { StepIndicatorProps, StepProgressProps } from "@types/index"
import { steps } from "@utils/constants"

const StepIndicator: React.FC<StepIndicatorProps> = ({
	step,
	index,
	currentStep,
	onStepClick,
	isEditMode,
	disabled,
}) => {
	const currentStepIndex = steps.indexOf(currentStep)
	const isActive = currentStepIndex >= index
	const isNextActive = currentStepIndex >= index + 1
	const isLastStep = index === steps.length - 1

	const handleClick = () => {
		if (isEditMode && !disabled && onStepClick) {
			onStepClick(step)
		}
	}

	return (
		<div className="flex flex-col items-start">
			<div className="flex items-center">
				<button
					className={`z-10 size-3 rounded-full border ${
						isActive
							? "border-primary outline outline-2 outline-primary"
							: "border-gray-300 bg-white"
					} ${isEditMode && !disabled ? "cursor-pointer hover:bg-[#e8ebff]" : "cursor-not-allowed"}`}
					onClick={handleClick}
					disabled={!isEditMode || disabled}
					type="button"
				/>
				{!isLastStep && (
					<div className="relative h-[2px] w-20">
						<div className="absolute inset-0 bg-gray-300" />
						{isNextActive && (
							<div className="absolute inset-0 bg-primary transition-all duration-300" />
						)}
					</div>
				)}
			</div>
			<button
				className={`mt-2 inline translate-x-[-40%] text-xs ${
					isActive ? "text-primary" : "text-[#6b7280]"
				} ${isEditMode && !disabled ? "cursor-pointer hover:text-primary" : "cursor-not-allowed"}`}
				onClick={handleClick}
				disabled={!isEditMode || disabled}
				type="button"
			>
				{step === "config"
					? "Job Config"
					: step.charAt(0).toUpperCase() + step.slice(1)}
			</button>
		</div>
	)
}

const StepProgress: React.FC<StepProgressProps> = ({
	currentStep,
	onStepClick,
	isEditMode,
	disabled,
}) => {
	return (
		<div className="flex items-center">
			{steps.map((step, index) => (
				<StepIndicator
					key={step}
					step={step}
					index={index}
					currentStep={currentStep}
					onStepClick={onStepClick}
					isEditMode={isEditMode}
					disabled={disabled}
				/>
			))}
		</div>
	)
}

export default StepProgress
