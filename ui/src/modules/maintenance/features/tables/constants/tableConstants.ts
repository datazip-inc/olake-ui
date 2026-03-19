import {
	CheckCircleIcon,
	SpinnerIcon,
	WarningCircleIcon,
} from "@phosphor-icons/react"
import type { ElementType } from "react"

import type { RunStatus } from "../types"

export const PAGE_SIZE = 6

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
	tag: "Q" | "S" | "D"
	name: string
}> = [
	{ key: "minor", tag: "Q", name: "Quick" },
	{ key: "major", tag: "S", name: "Standard" },
	{ key: "full", tag: "D", name: "Deep" },
]
