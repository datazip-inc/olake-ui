import { QuestionIcon } from "@phosphor-icons/react"
import { Tooltip } from "antd"

type AdvancedConfigRowProps = {
	label: string
	tooltip: string
	children: React.ReactNode
}

const AdvancedConfigRow: React.FC<AdvancedConfigRowProps> = ({
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

export default AdvancedConfigRow
