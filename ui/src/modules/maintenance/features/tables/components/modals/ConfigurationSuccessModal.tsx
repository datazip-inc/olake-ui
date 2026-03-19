import { Modal } from "antd"
import { useEffect, useRef, useState } from "react"

import ConfigSuccessIcon from "@/assets/config-success-icon.svg"

type ConfigurationSuccessModalProps = {
	open: boolean
	onClose: () => void
	firstRunAt?: string
}

const COUNTDOWN_START = 3

const ConfigurationSuccessModal: React.FC<ConfigurationSuccessModalProps> = ({
	open,
	onClose,
	firstRunAt,
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
			styles={{
				content: { padding: 0, borderRadius: 20, overflow: "hidden" },
				body: { padding: 0 },
			}}
		>
			<div className="flex h-[808px] flex-col items-center bg-white pt-[280px]">
				{/* Icon + title group */}
				<div className="flex flex-col items-center gap-6">
					<img
						src={ConfigSuccessIcon}
						alt=""
						aria-hidden
						className="h-14 w-28"
					/>
					<p className="whitespace-nowrap text-[30px] font-medium leading-[38px] text-olake-success-strong">
						Configuration Successful
					</p>
				</div>

				{/* First run section */}
				{firstRunAt && (
					<div className="mt-[46px] flex w-80 flex-col items-center gap-1 text-center">
						<p className="w-full text-xl leading-normal text-olake-success-strong">
							Your first run will start on
						</p>
						<p className="w-full text-sm leading-[22px] text-olake-text-secondary">
							{firstRunAt}
						</p>
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
