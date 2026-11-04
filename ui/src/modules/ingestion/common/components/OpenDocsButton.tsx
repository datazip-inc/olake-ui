import { ArrowSquareOutIcon } from "@phosphor-icons/react"
import { Button } from "antd"
import clsx from "clsx"

import { OpenDocsButtonProps } from "@/modules/ingestion/common/types"

// Opens the connector docs in a new tab. It sits in the page header rather than
// over the docs iframe, so it never covers the docs site's own navbar.
const OpenDocsButton: React.FC<OpenDocsButtonProps> = ({
	docUrl,
	className,
}) => (
	<Button
		type="default"
		icon={
			<ArrowSquareOutIcon
				size={20}
				weight="bold"
				className="text-primary"
			/>
		}
		onClick={() => window.open(docUrl, "_blank", "noopener,noreferrer")}
		className={clsx(
			"flex items-center gap-2 border border-gray-200 px-2 py-3 hover:border-blue-600 hover:text-blue-600",
			className,
		)}
	>
		Open Docs
	</Button>
)

export default OpenDocsButton
