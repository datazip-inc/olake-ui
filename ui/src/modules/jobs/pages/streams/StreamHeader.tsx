import { CaretRight } from "@phosphor-icons/react"
import { Checkbox, CheckboxChangeEvent } from "antd"
import clsx from "clsx"

import { StreamHeaderProps } from "../../../../types"

const StreamHeader: React.FC<StreamHeaderProps> = ({
	stream,
	toggle,
	checked,
	activeStreamData,
	setActiveStreamData,
}) => {
	const {
		stream: { name, namespace },
	} = stream

	//activestream is the stream selected in the stream panel
	const isActiveStream =
		activeStreamData?.stream.name === name &&
		activeStreamData?.stream.namespace === namespace

	//opens the stream configuration when user clicks on stream or on checkbox
	const handleChange = (e: CheckboxChangeEvent) => {
		toggle(e)
		setActiveStreamData(stream)
	}

	return (
		<div
			className={clsx(
				"flex w-full items-center justify-between border-b border-solid border-[#e5e7eb] py-3 pl-6",
				isActiveStream
					? "bg-primary-100"
					: "bg-white hover:bg-background-primary",
			)}
		>
			<div
				role="button"
				tabIndex={0}
				className="flex w-full cursor-pointer select-none items-center justify-between"
				onClick={() => setActiveStreamData(stream)}
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
						<CaretRight className="size-4 text-gray-500" />
					</div>
				)}
			</div>
		</div>
	)
}

export default StreamHeader
