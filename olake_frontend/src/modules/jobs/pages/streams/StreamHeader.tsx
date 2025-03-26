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
			className={`border-heading flex items-center justify-between border-[1px] border-b-0 border-solid p-3 first:border-t-0 last:border-b-[1px] ${
				isActiveStream ? "bg-[#e9ebfc]" : "bg-[#ffffff]"
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
