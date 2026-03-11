import {
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
} from "@phosphor-icons/react"
import { NavItem, TestConnectionStatus } from "../types"

export const THEME_CONFIG = {
	token: {
		colorPrimary: "#203FDD",
		borderRadius: 6,
	},
}

export const NAV_ITEMS: NavItem[] = [
	{ path: "/jobs", label: "Jobs", icon: GitCommitIcon },
	{ path: "/sources", label: "Sources", icon: LinktreeLogoIcon },
	{ path: "/destinations", label: "Destinations", icon: PathIcon },
]

// not showing oneof and const errors
export const transformErrors = (errors: any[]) => {
	return errors.filter(err => err.name !== "oneOf" && err.name !== "const")
}

export const TEST_CONNECTION_STATUS: Record<TestConnectionStatus, string> = {
	SUCCEEDED: "SUCCEEDED",
	FAILED: "FAILED",
} as const

export const HTTP_STATUS = {
	UNAUTHORIZED: 401,
	FORBIDDEN: 403,
	SERVER_ERROR: 500,
}

export const ERROR_MESSAGES = {
	AUTH_REQUIRED: "Authentication required. Please log in.",
	NO_PERMISSION: "You do not have permission to access this resource",
	SERVER_ERROR: "Server error occurred. Please try again later.",
	NO_RESPONSE:
		"No response received from server. Please check your connection.",
}

export const OLAKE_LATEST_VERSION_URL = "https://olake.io/docs/release/overview"

export const LOCALSTORAGE_TOKEN_KEY = "token"

export const LOCALSTORAGE_USERNAME_KEY = "username"
