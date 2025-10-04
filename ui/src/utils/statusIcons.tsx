import {
	ArrowsCounterClockwiseIcon,
	CheckIcon,
	XCircleIcon,
} from "@phosphor-icons/react"

export const getStatusIcon = (status: string | undefined) => {
	if (status === "success" || status === "completed") {
		return <CheckIcon className="text-green-500" />
	} else if (status === "failed" || status === "cancelled") {
		return <XCircleIcon className="text-red-500" />
	} else if (status === "running") {
		return <ArrowsCounterClockwiseIcon className="text-blue-500" />
	}
	return null
}
