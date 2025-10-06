import {
	ArrowsCounterClockwiseIcon,
	CheckIcon,
	XCircleIcon,
	XIcon,
} from "@phosphor-icons/react"

export const getStatusIcon = (status: string | undefined) => {
	if (status === "success" || status === "completed") {
		return <CheckIcon className="text-green-500" />
	} else if (status === "failed") {
		return <XCircleIcon className="text-red-500" />
	} else if (status === "running") {
		return <ArrowsCounterClockwiseIcon className="text-blue-500" />
	} else if (status === "canceled") {
		return <XIcon className="text-amber-500" />
	}
	return null
}
