import { GenderNeuter } from "@phosphor-icons/react"
import { EndpointTitleProps } from "../types"

const EndpointTitle = ({ title = "Endpoint config" }: EndpointTitleProps) => (
	<div className="mb-4 flex items-center gap-1">
		<div className="mb-2 flex items-center gap-2">
			<GenderNeuter className="size-5" />
			<div className="text-base font-medium">{title}</div>
		</div>
	</div>
)

export default EndpointTitle
