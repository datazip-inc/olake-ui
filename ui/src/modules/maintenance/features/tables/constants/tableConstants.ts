import {
	CheckCircleIcon,
	SpinnerIcon,
	WarningCircleIcon,
} from "@phosphor-icons/react"
import type { ElementType } from "react"

import type { RunStatus } from "../types"

export const PAGE_SIZE = 10

export const runStatusConfig: Record<
	RunStatus,
	{
		Icon: ElementType
		bgClass: string
		textClass: string
		label: string
		iconClass?: string
	}
> = {
	SUCCESS: {
		Icon: CheckCircleIcon,
		bgClass: "bg-olake-success-bg",
		textClass: "text-olake-success",
		label: "Success",
	},
	RUNNING: {
		Icon: SpinnerIcon,
		bgClass: "bg-olake-warning-bg",
		textClass: "text-olake-warning",
		label: "Running",
	},
	FAILED: {
		Icon: WarningCircleIcon,
		bgClass: "bg-olake-error-bg",
		textClass: "text-olake-error",
		label: "Failed",
	},
}

export const runLogsStatusConfig: Record<
	RunStatus,
	{
		Icon: ElementType
		bgClass: string
		textClass: string
		label: string
	}
> = {
	SUCCESS: {
		Icon: CheckCircleIcon,
		bgClass: "bg-olake-success-bg",
		textClass: "text-olake-success",
		label: "Success",
	},
	RUNNING: {
		Icon: SpinnerIcon,
		bgClass: "bg-olake-warning-bg",
		textClass: "text-olake-warning",
		label: "Running",
	},
	FAILED: {
		Icon: WarningCircleIcon,
		bgClass: "bg-olake-error-bg",
		textClass: "text-olake-error-alt",
		label: "Failed",
	},
}

export const compactionSlots: Array<{
	key: "minor" | "major" | "full"
	tag: "L" | "M" | "F"
	name: string
}> = [
	{ key: "minor", tag: "L", name: "Lite" },
	{ key: "major", tag: "M", name: "Medium" },
	{ key: "full", tag: "F", name: "Full" },
]
