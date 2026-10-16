import { Switch } from "antd"
import clsx from "clsx"
import { ReactNode } from "react"

export type CdcCardState = "on" | "off" | "not_supported"

const DEFAULT_DESCRIPTION =
	"Capture inserts, updates, and deletes using CDC. Turn off to use full refresh or incremental mode."

const NOT_SUPPORTED_DESCRIPTION =
	"Change data capture is not supported for this connector."

/**
 * The "Enable CDC" row of a source form. `children` are the CDC settings shown below the
 * row while CDC is on (replication slot, publication, ...).
 */
const CdcCard = ({
	state,
	description,
	onToggle,
	disabled = false,
	children,
}: {
	state: CdcCardState
	description?: string
	onToggle?: (enabled: boolean) => void
	disabled?: boolean
	children?: ReactNode
}) => {
	const notSupported = state === "not_supported"
	const helper = notSupported
		? NOT_SUPPORTED_DESCRIPTION
		: description || DEFAULT_DESCRIPTION

	return (
		<div className="flex flex-col gap-4">
			<div
				className={clsx(
					"flex items-center justify-between gap-4 rounded-lg border px-5 py-4",
					notSupported
						? "border-dashed border-olake-border bg-olake-surface-subtle"
						: "border-olake-border-secondary bg-olake-surface",
				)}
			>
				<div className="flex flex-col gap-1">
					<div className="flex items-center gap-2">
						<span
							className={clsx(
								"text-sm font-medium",
								notSupported ? "text-olake-text-disabled" : "text-olake-text",
							)}
						>
							Enable CDC
						</span>
						{notSupported && (
							<span className="rounded bg-olake-surface-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-olake-text-tertiary">
								Not supported
							</span>
						)}
					</div>
					<span
						className={clsx(
							"text-xs",
							notSupported
								? "text-olake-text-disabled"
								: "text-olake-text-secondary",
						)}
					>
						{helper}
					</span>
				</div>
				<Switch
					aria-label="Enable CDC"
					checked={state === "on"}
					disabled={notSupported || disabled}
					onChange={checked => onToggle?.(checked)}
				/>
			</div>
			{state === "on" && children}
		</div>
	)
}

export default CdcCard
