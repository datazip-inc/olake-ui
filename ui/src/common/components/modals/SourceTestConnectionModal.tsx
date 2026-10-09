import {
	ArrowRightIcon,
	CaretDownIcon,
	CaretUpIcon,
	CheckCircleIcon,
	CircleIcon,
	SpinnerIcon,
	WarningCircleIcon,
	WarningIcon,
} from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import clsx from "clsx"
import { ReactNode, useEffect, useMemo, useState } from "react"

import { PrerequisiteCheck } from "@/common/types"

// Delay between revealing consecutive checks. The backend returns every check at once,
// so the reveal only paces the results for the user.
const REVEAL_INTERVAL_MS = 300

type SectionKey = "required" | "optional"
type SectionStatus = "pending" | "running" | "passed" | "failed" | "warning"
type RowStatus = "pending" | "running" | "done"

const StatusIcon = ({
	status,
	size = 20,
}: {
	status: SectionStatus
	size?: number
}) => {
	switch (status) {
		case "passed":
			return (
				<CheckCircleIcon
					size={size}
					className="text-olake-success"
				/>
			)
		case "failed":
			return (
				<WarningCircleIcon
					size={size}
					className="text-olake-error"
				/>
			)
		case "warning":
			return (
				<WarningIcon
					size={size}
					className="text-olake-warning"
				/>
			)
		case "running":
			return (
				<SpinnerIcon
					size={size}
					className="animate-spin text-olake-text-tertiary"
				/>
			)
		default:
			return (
				<SpinnerIcon
					size={size}
					className="text-olake-text-disabled"
				/>
			)
	}
}

const Section = ({
	title,
	subtitle,
	status,
	expanded,
	onToggle,
	isLast,
	children,
}: {
	title: string
	subtitle: string
	status: SectionStatus
	expanded?: boolean
	onToggle?: () => void
	isLast: boolean
	children?: ReactNode
}) => (
	<div className="flex gap-3">
		<div className="flex flex-col items-center">
			<StatusIcon status={status} />
			{!isLast && <div className="my-1 w-px flex-1 bg-olake-border" />}
		</div>
		<div className={clsx("flex-1", !isLast && "pb-6")}>
			<div className="flex items-start justify-between gap-4">
				<div>
					<p
						className={clsx(
							"text-base font-medium leading-5",
							status === "pending"
								? "text-olake-text-tertiary"
								: "text-olake-text",
						)}
					>
						{title}
					</p>
					<p className="mt-1 text-sm text-olake-text-tertiary">{subtitle}</p>
				</div>
				{onToggle && (
					<button
						type="button"
						onClick={onToggle}
						aria-label={expanded ? "Collapse" : "Expand"}
						aria-expanded={expanded}
						className="p-1 text-olake-text-secondary"
					>
						{expanded ? <CaretUpIcon size={16} /> : <CaretDownIcon size={16} />}
					</button>
				)}
			</div>
			{expanded && children && (
				<div className="mt-4 flex flex-col">{children}</div>
			)}
		</div>
	</div>
)

const PrerequisiteRow = ({
	check,
	status,
	isLast,
}: {
	check: PrerequisiteCheck
	status: RowStatus
	isLast: boolean
}) => {
	const failed = status === "done" && !check.passed
	const failColor = check.required ? "text-olake-error" : "text-olake-warning"

	const icon =
		status === "pending" ? (
			<CircleIcon
				size={14}
				className="text-olake-text-disabled"
			/>
		) : (
			<StatusIcon
				size={14}
				status={
					status === "running"
						? "running"
						: check.passed
							? "passed"
							: check.required
								? "failed"
								: "warning"
				}
			/>
		)

	return (
		<div className="flex gap-2">
			<div className="flex flex-col items-center pt-[3px]">
				{icon}
				{!isLast && <div className="my-1 w-px flex-1 bg-olake-border" />}
			</div>
			<div className={clsx("flex-1", !isLast && "pb-4")}>
				<p className={clsx("text-sm", failed ? failColor : "text-olake-text")}>
					{check.name}
				</p>
				{status === "done" && check.passed && (
					<p className="text-xs text-olake-text-tertiary">
						{check.current_value
							? `Correct (${check.current_value})`
							: "Correct"}
					</p>
				)}
				{failed && (
					<>
						{check.description && (
							<p className="mt-0.5 text-xs text-olake-text">
								{check.description}
							</p>
						)}
						<div className="mt-3 flex items-center gap-6">
							<div className="flex flex-col gap-1">
								<span className="text-xs text-olake-text-tertiary">
									Current
								</span>
								<span className={clsx("text-sm", failColor)}>
									{check.current_value || "-"}
								</span>
							</div>
							<ArrowRightIcon
								size={20}
								className="shrink-0 text-olake-text-tertiary"
							/>
							<div className="flex flex-col gap-1">
								<span className="text-xs text-olake-text-tertiary">
									{check.required ? "Required" : "Recommended"}
								</span>
								<span className="text-sm text-olake-text-secondary">
									{check.recommended_value || "-"}
								</span>
							</div>
						</div>
					</>
				)}
			</div>
		</div>
	)
}

const formatSeconds = (ms: number) => `${Math.max(1, Math.round(ms / 1000))}s`

const SourceTestConnectionModal = ({
	open,
	prerequisites,
	startedAt,
	finishedAt,
	onConfirm,
	onAllPassed,
	onRetry,
	onCancel,
}: {
	open: boolean
	prerequisites: PrerequisiteCheck[]
	// Connection test timing; finishedAt also identifies the attempt shown
	startedAt: number
	finishedAt: number
	onConfirm: () => void
	onAllPassed: () => void
	onRetry: () => void
	onCancel: () => void
}) => {
	const [revealed, setRevealed] = useState(0)
	const [expandedOverride, setExpandedOverride] = useState<
		Partial<Record<SectionKey, boolean>>
	>({})

	// Keep the backend order: failed checks come before passed ones inside each group
	const required = useMemo(
		() => prerequisites.filter(c => c.required),
		[prerequisites],
	)
	const optional = useMemo(
		() => prerequisites.filter(c => !c.required),
		[prerequisites],
	)
	const total = required.length + optional.length

	// Every new attempt starts from a clean state. Reset during render, not in an effect,
	// so the first render of a new attempt never sees the previous attempt's reveal count.
	const [attempt, setAttempt] = useState(finishedAt)
	if (attempt !== finishedAt) {
		setAttempt(finishedAt)
		setRevealed(0)
		setExpandedOverride({})
	}

	useEffect(() => {
		if (!open || revealed >= total) return
		const timer = setTimeout(() => setRevealed(r => r + 1), REVEAL_INTERVAL_MS)
		return () => clearTimeout(timer)
	}, [open, revealed, total])

	const revealDone = revealed >= total
	const requiredFailed = required.filter(c => !c.passed).length
	const optionalFailed = optional.filter(c => !c.passed).length
	const allPassed = revealDone && requiredFailed === 0 && optionalFailed === 0

	useEffect(() => {
		if (open && allPassed) onAllPassed()
		// onAllPassed is intentionally left out: it must fire once per completed attempt
	}, [open, allPassed])

	const rowStatus = (sequenceIndex: number): RowStatus => {
		if (sequenceIndex > revealed) return "pending"
		return sequenceIndex < revealed ? "done" : "running"
	}

	const sectionState = (
		key: SectionKey,
		checks: PrerequisiteCheck[],
		offset: number,
	) => {
		const failed = checks.filter(c => !c.passed).length
		const doneCount = Math.min(Math.max(revealed - offset, 0), checks.length)
		let status: SectionStatus
		let subtitle: string
		if (revealed < offset) {
			status = "pending"
			subtitle = "Not started yet"
		} else if (doneCount < checks.length) {
			status = "running"
			subtitle = `Checking ${doneCount + 1}/${checks.length} settings...`
		} else if (failed > 0) {
			status = key === "required" ? "failed" : "warning"
			subtitle = `${failed}/${checks.length} ${key} checks failed`
		} else {
			status = "passed"
			subtitle = `${checks.length}/${checks.length} checks passed`
		}
		const autoExpanded = status === "running" || (failed > 0 && doneCount > 0)
		return {
			status,
			subtitle,
			expanded: expandedOverride[key] ?? autoExpanded,
			onToggle:
				checks.length > 0
					? () =>
							setExpandedOverride(o => ({
								...o,
								[key]: !(o[key] ?? autoExpanded),
							}))
					: undefined,
		}
	}

	// The modal opens only after the connection succeeded
	const connectionSubtitle = `Connection Successful • ${formatSeconds(finishedAt - startedAt)}`
	const showRequired = required.length > 0
	const showOptional = optional.length > 0

	const requiredSection = sectionState("required", required, 0)
	const optionalSection = sectionState("optional", optional, required.length)

	const renderRows = (checks: PrerequisiteCheck[], offset: number) =>
		checks.map((check, i) => (
			<PrerequisiteRow
				key={check.name}
				check={check}
				status={rowStatus(offset + i)}
				isLast={i === checks.length - 1}
			/>
		))

	const blocked = revealDone && requiredFailed > 0
	const canConfirm = revealDone && requiredFailed === 0 && optionalFailed > 0

	return (
		<Modal
			open={open}
			title="Testing Source Configuration"
			onCancel={onCancel}
			maskClosable={false}
			centered
			width={780}
			destroyOnHidden
			footer={null}
			styles={{
				content: { padding: 0, overflow: "hidden", borderRadius: 20 },
				header: {
					padding: "32px 36px 20px",
					margin: 0,
					borderBottom: "1px solid #D9D9D9",
				},
				body: { padding: 0 },
			}}
		>
			<div className="flex h-[560px] flex-col">
				<div className="flex-1 overflow-auto px-9 py-6">
					<Section
						title="Test Connection"
						subtitle={connectionSubtitle}
						status="passed"
						isLast={!showRequired && !showOptional}
					/>
					{showRequired && (
						<Section
							title="Required Source CDC Pre-Requisites"
							{...requiredSection}
							isLast={!showOptional}
						>
							{renderRows(required, 0)}
						</Section>
					)}
					{showOptional && (
						<Section
							title="Optional Source CDC Settings"
							{...optionalSection}
							isLast
						>
							{renderRows(optional, required.length)}
						</Section>
					)}
				</div>

				{blocked && (
					<div className="flex items-center gap-2 border-t border-olake-border bg-olake-error-bg px-9 py-2 text-sm text-olake-error">
						<WarningCircleIcon size={16} />
						Some required source server settings are misconfigured. Configure
						them to the recommended values for reliable CDC syncs.
					</div>
				)}
				{canConfirm && (
					<div className="flex items-center gap-2 border-t border-olake-border bg-olake-success-bg px-9 py-2 text-sm text-olake-success-strong">
						<CheckCircleIcon
							size={16}
							className="text-olake-success"
						/>
						Some optional settings are not configured to the recommended values.
						Updating them is recommended for reliable CDC syncs.
					</div>
				)}

				<div className="flex items-center gap-2 border-t border-olake-border px-9 py-5">
					{blocked ? (
						<Button
							type="primary"
							onClick={onRetry}
						>
							Try Again
						</Button>
					) : canConfirm ? (
						<Button
							type="primary"
							onClick={onConfirm}
						>
							Continue anyway
						</Button>
					) : (
						<Button disabled>In Progress...</Button>
					)}
					<Button onClick={onCancel}>Cancel</Button>
				</div>
			</div>
		</Modal>
	)
}

export default SourceTestConnectionModal
