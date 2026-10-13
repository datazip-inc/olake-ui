import {
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
} from "@phosphor-icons/react"
import type { RJSFValidationError } from "@rjsf/utils"

import { NavItem, TestConnectionStatus } from "../types"

export const NAV_ITEMS: NavItem[] = [
	{ path: "/jobs", label: "Jobs", icon: GitCommitIcon },
	{ path: "/sources", label: "Sources", icon: LinktreeLogoIcon },
	{ path: "/destinations", label: "Destinations", icon: PathIcon },
]

// every oneOf branch an error sits in, outermost first:
// ".../oneOf/3/dependencies/x/oneOf/1/required" -> [".../oneOf/3", ".../oneOf/3/dependencies/x/oneOf/1"]
const oneOfBranches = (schemaPath = "") =>
	Array.from(schemaPath.matchAll(/\/oneOf\/\d+/g), m =>
		schemaPath.slice(0, m.index + m[0].length),
	)

// AJV validates form data against every oneOf branch (e.g. every catalog
// type), so fields required only by other branches would be flagged. A branch
// that reports a const error does not match the form data; drop all errors
// inside it, including nested ones. oneOf and const errors are not shown either.
export const transformErrors = (errors: RJSFValidationError[]) => {
	const inactiveBranches = new Set<string>()
	for (const err of errors) {
		if (err.name === "const") {
			const innermost = oneOfBranches(err.schemaPath).pop()
			if (innermost) inactiveBranches.add(innermost)
		}
	}

	return errors.filter(
		err =>
			err.name !== "oneOf" &&
			err.name !== "const" &&
			!oneOfBranches(err.schemaPath).some(branch =>
				inactiveBranches.has(branch),
			),
	)
}

// Toast text for spec-driven form errors, keyed by AJV keyword ("type" errors
// by "type.<expected type>"). {label} is the field label shown on screen and
// {limit} the keyword's limit.
export const FORM_ERROR_MESSAGES: Record<string, string> = {
	required: "{label} can't be empty",
	// specs only use minLength: 1, i.e. non-empty
	minLength: "{label} can't be empty",
	pattern: "{label} has an invalid format",
	minimum: "{label} must be at least {limit}",
	maximum: "{label} must be at most {limit}",
	exclusiveMinimum: "{label} must be greater than {limit}",
	enum: "Select a valid option for {label}",
	"type.integer": "{label} must be a whole number",
	"type.number": "{label} must be a number",
}

// Spec patterns explained in words, since "invalid format" alone does not tell
// the user what to change
export const PATTERN_ERROR_MESSAGES: Record<string, string> = {
	"^[a-z_]+$":
		"{label} can only contain lowercase letters (a-z) and underscores (_)",
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

export const OLAKE_COMMUNITY_SLACK_URL = "https://olake.io/slack/"

export const LOCALSTORAGE_TOKEN_KEY = "token"

export const LOCALSTORAGE_USERNAME_KEY = "username"

export const SOURCE_ONBOARDING_DISMISSED_SESSION_KEY =
	"source_onboarding_modal_dismissed"

export const COMMUNITY_BANNER_DISMISSED_SESSION_KEY =
	"community_help_banner_dismissed"
