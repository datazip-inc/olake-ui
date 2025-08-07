import { Checkbox, CheckboxChangeEvent } from "antd"
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

	const handleChange = (e: CheckboxChangeEvent) => {
		e.stopPropagation()
		toggle(e)
		setActiveStreamData(stream)
	}

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
				onClick={() => setActiveStreamData(stream)}
			>
				<div className="flex items-center gap-2">
					<Checkbox
						checked={checked}
						onChange={handleChange}
						onClick={e => e.stopPropagation()}
						className={`text-lg ${checked && "text-[#1FA7C9]"}`}
					/>
					{name}
				</div>
			</div>
		</div>
	)
}

export default StreamHeader
