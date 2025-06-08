import { ArrowsCounterClockwise, Check, XCircle } from "@phosphor-icons/react"

export const getStatusIcon = (status: string | undefined) => {
	if (status === "success" || status === "completed") {
		return <Check className="text-green-500" />
	} else if (status === "failed" || status === "cancelled") {
		return <XCircle className="text-red-500" />
	} else if (status === "running") {
		return <ArrowsCounterClockwise className="text-blue-500" />
	}
	return null
}
