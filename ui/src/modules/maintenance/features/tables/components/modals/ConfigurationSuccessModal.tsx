import { Modal } from "antd"
import { useEffect, useRef, useState } from "react"

import ConfigSuccessIcon from "@/assets/config-success-icon.svg"

import { DEFAULT_TABLE_MODAL_STYLES } from "../../constants"

type ConfigurationSuccessModalProps = {
	open: boolean
	onClose: () => void
	/** Upcoming optimize runs across Lite/Medium/Full, earliest first. */
	optimizeRuns?: string[]
	/** Upcoming cleanup (expire snapshots) runs, earliest first. */
	cleanupRuns?: string[]
}

const NextRuns: React.FC<{ title: string; runs: string[] }> = ({
	title,
	runs,
}) => (
	<div className="flex w-full flex-col items-center gap-2 text-center">
		<p className="text-base font-medium leading-6 text-olake-success-strong">
			{title}
		</p>
		{runs.length > 0 ? (
			<div className="flex w-full items-center gap-2">
				{runs.map(run => (
					<span
						key={`${title}-${run}`}
						className="min-w-0 flex-1 whitespace-nowrap rounded-[6px] bg-olake-surface-muted px-2 py-1 text-center text-sm leading-[22px] text-olake-text-secondary"
					>
						{run}
					</span>
				))}
			</div>
		) : (
			<p className="text-sm leading-[22px] text-olake-text-secondary">
				No runs scheduled
			</p>
		)}
	</div>
)

const COUNTDOWN_START = 3

const ConfigurationSuccessModal: React.FC<ConfigurationSuccessModalProps> = ({
	open,
	onClose,
	optimizeRuns,
	cleanupRuns,
}) => {
	const [countdown, setCountdown] = useState(COUNTDOWN_START)
	const onCloseRef = useRef(onClose)
	onCloseRef.current = onClose

	useEffect(() => {
		if (!open) {
			setCountdown(COUNTDOWN_START)
			return
		}

		const interval = setInterval(() => {
			setCountdown(prev => {
				if (prev <= 1) {
					clearInterval(interval)
					onCloseRef.current()
					return 0
				}
				return prev - 1
			})
		}, 1000)

		return () => clearInterval(interval)
	}, [open])

	return (
		<Modal
			open={open}
			onCancel={onClose}
			footer={null}
			closable={false}
			centered
			width={696}
			destroyOnHidden
			styles={DEFAULT_TABLE_MODAL_STYLES}
		>
			<div className="flex h-[808px] flex-col items-center bg-white pt-[200px]">
				{/* Icon + title group */}
				<div className="flex flex-col items-center gap-6">
					<p className="whitespace-nowrap text-[30px] font-medium leading-[38px] text-olake-success-strong">
						Configuration Successful
					</p>
					<img
						src={ConfigSuccessIcon}
						alt=""
						aria-hidden
						className="h-14 w-28"
					/>
				</div>

				{/* Upcoming runs */}
				{(optimizeRuns || cleanupRuns) && (
					<div className="mt-10 flex w-[632px] flex-col items-center gap-6">
						{optimizeRuns && (
							<NextRuns
								title="Next 3 optimize runs"
								runs={optimizeRuns}
							/>
						)}
						{cleanupRuns && (
							<NextRuns
								title="Next 3 cleanup runs"
								runs={cleanupRuns}
							/>
						)}
					</div>
				)}

				{/* Button */}
				<button
					type="button"
					onClick={onClose}
					className="mb-[63px] mt-auto flex h-10 w-48 items-center justify-center rounded-lg border border-olake-border bg-white text-base leading-6 text-olake-text"
				>
					{countdown > 0 ? `Closing in ${countdown}` : "Close"}
				</button>
			</div>
		</Modal>
	)
}

export default ConfigurationSuccessModal
