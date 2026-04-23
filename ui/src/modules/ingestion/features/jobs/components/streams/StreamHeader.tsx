import { CaretRightIcon } from "@phosphor-icons/react"
import { Checkbox, CheckboxChangeEvent } from "antd"
import clsx from "clsx"
import { useShallow } from "zustand/react/shallow"

import { StreamHeaderProps } from "@/modules/ingestion/features/jobs/types"

import { useStreamSelectionStore, selectActiveStreamKey } from "../../stores"

const StreamHeader: React.FC<StreamHeaderProps> = ({
	stream,
	toggle,
	checked,
}) => {
	const activeStreamKey = useStreamSelectionStore(
		useShallow(selectActiveStreamKey),
	)
	const setActiveStreamKey = useStreamSelectionStore(
		state => state.setActiveStreamKey,
	)

	const {
		stream: { name, namespace },
	} = stream

	const isActiveStream =
		activeStreamKey?.streamName === name &&
		activeStreamKey?.namespace === (namespace ?? "")

	const setActive = () =>
		setActiveStreamKey({ streamName: name, namespace: namespace ?? "" })

	const handleChange = (e: CheckboxChangeEvent) => {
		toggle(e)
		setActive()
	}

	return (
		<div
			className={clsx(
				"flex w-full items-center justify-between border-b border-solid border-[#e5e7eb] py-3 pl-6",
				isActiveStream
					? "bg-[#D2D8F7]"
					: "bg-white hover:bg-background-primary",
			)}
		>
			<div
				role="button"
				tabIndex={0}
				className="flex w-full cursor-pointer select-none items-center justify-between"
				onClick={setActive}
			>
				<div className="flex items-center gap-2">
					<Checkbox
						checked={checked}
						onChange={handleChange}
						onClick={e => e.stopPropagation()}
						className={clsx("text-lg", checked && "text-[#1FA7C9]")}
					/>
					{name}
				</div>
				{!isActiveStream && (
					<div className="mr-4">
						<CaretRightIcon className="size-4 text-gray-500" />
					</div>
				)}
			</div>
		</div>
	)
}

export default StreamHeader
