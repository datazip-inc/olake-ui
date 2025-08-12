import { Checkbox } from "antd"
import { StreamHeaderProps } from "../../../../types"
import { CaretRight } from "@phosphor-icons/react"
import clsx from "clsx"

const StreamHeader: React.FC<StreamHeaderProps> = ({
	stream,
	toggle,
	checked,
	activeStreamData,
	setActiveStreamData,
}) => {
	const {
		stream: { name },
	} = stream

	const isActiveStream = activeStreamData?.stream.name === name

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
				onClick={() => {
					setActiveStreamData(stream)
				}}
			>
				<div className="flex items-center gap-2">
					<div
						role="button"
						tabIndex={0}
						onClick={e => e.stopPropagation()}
					>
						<Checkbox
							checked={checked}
							onChange={toggle}
							className={clsx("text-lg", checked && "text-[#1FA7C9]")}
						/>
					</div>
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
