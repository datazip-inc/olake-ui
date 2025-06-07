import { Checkbox } from "antd"
import { StreamHeaderProps } from "../../../../types"

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
			className={`flex w-full items-center justify-between border-b border-solid border-[#e5e7eb] py-3 pl-6 ${
				isActiveStream ? "bg-[#e9ebfc]" : "bg-[#ffffff] hover:bg-[#f5f5f5]"
			}`}
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
							className={`text-lg ${checked && "text-[#1FA7C9]"}`}
						/>
					</div>
					{name}
				</div>
			</div>
		</div>
	)
}

export default StreamHeader
