import { JobCreationSteps } from "../types"

export const JOB_CREATION_STEPS: Record<string, JobCreationSteps> = {
	CONFIG: "config",
	STREAMS: "streams",
} as const

export const TAB_TYPES = {
	CONFIG: "config",
	SCHEMA: "schema",
	JOBS: "jobs",
}

export const JOB_STATUS = {
	ACTIVE: "active",
	INACTIVE: "inactive",
	SAVED: "saved",
	FAILED: "failed",
}

export const steps: string[] = [
	JOB_CREATION_STEPS.CONFIG,
	JOB_CREATION_STEPS.STREAMS,
]

export const JobTutorialYTLink =
	"https://youtu.be/_qRulFv-BVM?si=NPTw9V0hWQ3-9wOP"

export const DAYS_MAP = {
	Sunday: 0,
	Monday: 1,
	Tuesday: 2,
	Wednesday: 3,
	Thursday: 4,
	Friday: 5,
	Saturday: 6,
}

export const DAYS = [
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
]

export const JOB_STEP_NUMBERS = {
	CONFIG: 1,
	STREAMS: 2,
} as const

export const FREQUENCY_OPTIONS = [
	{ value: "minutes", label: "Every Minute" },
	{ value: "hours", label: "Every Hour" },
	{ value: "days", label: "Every Day" },
	{ value: "weeks", label: "Every Week" },
	{ value: "custom", label: "Custom" },
]
