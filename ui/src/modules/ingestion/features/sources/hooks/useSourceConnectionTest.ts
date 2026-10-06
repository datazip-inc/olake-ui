import { useRef, useState } from "react"

import { TEST_CONNECTION_STATUS } from "@/common/constants"
import type {
	PrerequisiteCheck,
	TestConnectionError,
	TestConnectionResponse,
} from "@/common/types"
import type { EntityTestRequest } from "@/modules/ingestion/common/types"

import { useTestSourceConnection } from "./mutations/useSourceMutations"

// How long the success modal stays up before the caller continues
const SUCCESS_MODAL_MS = 1000

export type SourceConnectionTestResult =
	| { outcome: "passed" }
	| { outcome: "cancelled" }
	| { outcome: "connection_failed"; error: TestConnectionError }

/**
 * Runs a source test connection. While the request runs the caller shows TestConnectionModal
 * (`testing`); once it succeeds, SourceTestConnectionModal lists the CDC prerequisite checks.
 * The returned promise resolves once the user is done: all checks passed, "Continue anyway" on
 * failed optional checks, Cancel, or a failed connection (shown by the caller's failure modal).
 * Try Again re-tests the same payload without resolving.
 */
export const useSourceConnectionTest = () => {
	const testSourceMutation = useTestSourceConnection()

	const [testing, setTesting] = useState(false)
	const [open, setOpen] = useState(false)
	const [showSuccess, setShowSuccess] = useState(false)
	const [prerequisites, setPrerequisites] = useState<PrerequisiteCheck[]>([])
	const [startedAt, setStartedAt] = useState(() => Date.now())
	const [finishedAt, setFinishedAt] = useState(() => Date.now())

	const requestRef = useRef<{
		source: EntityTestRequest
		existing: boolean
	} | null>(null)
	const resolveRef = useRef<((r: SourceConnectionTestResult) => void) | null>(
		null,
	)
	const attemptRef = useRef(0)

	const finish = (result: SourceConnectionTestResult) => {
		attemptRef.current += 1 // drops any response still in flight
		setTesting(false)
		setOpen(false)
		resolveRef.current?.(result)
		resolveRef.current = null
	}

	const execute = async () => {
		const request = requestRef.current
		if (!request) return
		const attempt = ++attemptRef.current

		const started = Date.now()
		setOpen(false)
		setTesting(true)

		const result = await testSourceMutation
			.mutateAsync(request)
			.catch(() => null)
		if (attempt !== attemptRef.current) return
		setTesting(false)

		const connection: TestConnectionResponse["connection_result"] | undefined =
			result?.data?.connection_result
		const checks = connection?.prerequisites ?? []
		const requiredFailed = checks.some(c => c.required && !c.passed)
		if (
			connection?.status !== TEST_CONNECTION_STATUS.SUCCEEDED &&
			!requiredFailed
		) {
			finish({
				outcome: "connection_failed",
				error: {
					message: connection?.message || "Source connection test failed",
					logs: result?.data?.logs || [],
				},
			})
			return
		}

		if (checks.length === 0) {
			// Nothing to review (driver without CDC checks): keep the plain success flow
			showSuccessThenPass()
			return
		}
		setPrerequisites(checks)
		setStartedAt(started)
		setFinishedAt(Date.now())
		setOpen(true)
	}

	const run = (source: EntityTestRequest, existing = false) => {
		// A previous run still waiting on the modal counts as cancelled
		resolveRef.current?.({ outcome: "cancelled" })
		requestRef.current = { source, existing }
		return new Promise<SourceConnectionTestResult>(resolve => {
			resolveRef.current = resolve
			void execute()
		})
	}

	const showSuccessThenPass = () => {
		const attempt = attemptRef.current
		setOpen(false)
		setShowSuccess(true)
		setTimeout(() => {
			setShowSuccess(false)
			if (attempt === attemptRef.current) finish({ outcome: "passed" })
		}, SUCCESS_MODAL_MS)
	}

	return {
		run,
		testing,
		showSuccess,
		modalProps: {
			open,
			prerequisites,
			startedAt,
			finishedAt,
			onConfirm: () => finish({ outcome: "passed" }),
			onAllPassed: showSuccessThenPass,
			onRetry: () => void execute(),
			onCancel: () => finish({ outcome: "cancelled" }),
		},
	}
}
