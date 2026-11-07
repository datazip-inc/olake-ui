import { CaretUpIcon, QuestionIcon } from "@phosphor-icons/react"
import { Tooltip } from "antd"
import clsx from "clsx"
import { useState } from "react"

type AdvancedConfigSectionProps = {
	children: React.ReactNode
	className?: string
}

/** Collapsible "Advanced Config" block used by both configure tabs; open by default. */
const AdvancedConfigSection: React.FC<AdvancedConfigSectionProps> = ({
	children,
	className,
}) => {
	const [open, setOpen] = useState(true)

	return (
		<div className={className}>
			<button
				type="button"
				className="flex w-full items-center justify-between py-2"
				onClick={() => setOpen(prev => !prev)}
			>
				<span className="text-base font-medium leading-6 text-olake-text">
					Advanced Config
				</span>
				<CaretUpIcon
					size={16}
					className={clsx(
						"text-olake-text-tertiary transition-transform",
						!open && "rotate-180",
					)}
				/>
			</button>
			{open && <div className="flex flex-col gap-4 pb-8 pt-2">{children}</div>}
		</div>
	)
}

type AdvancedConfigRowProps = {
	label: string
	tooltip: string
	children: React.ReactNode
}

/** One labelled setting inside AdvancedConfigSection: label + help icon on the left, control on the right. */
export const AdvancedConfigRow: React.FC<AdvancedConfigRowProps> = ({
	label,
	tooltip,
	children,
}) => (
	<div className="flex items-center justify-between">
		<div className="flex items-center gap-2">
			<span className="text-base leading-6 text-olake-text">{label}</span>
			<Tooltip title={tooltip}>
				<QuestionIcon
					size={14}
					className="cursor-help text-olake-text-tertiary"
				/>
			</Tooltip>
		</div>
		{children}
	</div>
)

export default AdvancedConfigSection
