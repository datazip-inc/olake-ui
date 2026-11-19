import { Tooltip } from "antd"
import clsx from "clsx"

interface TruncatedTextProps {
	text: string
	className?: string
}

// Single-line text that ellipsizes inside its container; the full value is in the tooltip.
const TruncatedText: React.FC<TruncatedTextProps> = ({ text, className }) => (
	<Tooltip title={text}>
		<span className={clsx("block min-w-0 truncate", className)}>{text}</span>
	</Tooltip>
)

export default TruncatedText
