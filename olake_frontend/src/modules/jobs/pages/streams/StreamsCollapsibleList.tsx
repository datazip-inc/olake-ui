import { Checkbox, Empty } from "antd"
import { StreamsCollapsibleListProps } from "../../../../types"
import { useEffect, useState } from "react"
import StreamPanel from "./StreamPanel"

const StreamsCollapsibleList = ({
	streamsToDisplay,
	allChecked,
	handleToggleAllStreams,
	activeStreamData,
	setActiveStreamData,
	selectedStreams,
	onStreamSelect,
}: StreamsCollapsibleListProps) => {
	const [options, SetOptions] = useState({ page: 1, pageSize: 20 })

	useEffect(() => {
		SetOptions(prev => ({ ...prev, page: 1 }))
	}, [streamsToDisplay])

	return (
		<div className="flex h-full flex-col">
			{/* To select all streams */}
			{streamsToDisplay?.length !== 0 && (
				<div
					className="flex select-none items-center border-[1px] border-solid bg-[#ffffff] p-3"
					onAnimationEndCapture={() => {}}
				>
					<div
						role="button"
						tabIndex={0}
						className="flex items-center gap-3"
						onClick={e => e.stopPropagation()}
					>
						<Checkbox
							checked={allChecked}
							onChange={handleToggleAllStreams}
						>
							Sync All
						</Checkbox>
					</div>
				</div>
			)}

			<div className="flex-1 overflow-y-scroll">
				{streamsToDisplay?.length !== 0 ? (
					streamsToDisplay
						.slice(
							(options.page - 1) * options.pageSize,
							options.page * options.pageSize,
						)
						.map(streamData => {
							return (
								<StreamPanel
									stream={streamData}
									key={streamData?.stream?.name}
									activeStreamData={activeStreamData}
									setActiveStreamData={setActiveStreamData}
									onStreamSelect={onStreamSelect}
									isSelected={selectedStreams.includes(streamData.stream.name)}
								/>
							)
						})
				) : (
					<Empty className="pt-10" />
				)}
			</div>
		</div>
	)
}

export default StreamsCollapsibleList
