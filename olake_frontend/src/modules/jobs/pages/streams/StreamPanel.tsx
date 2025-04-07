import { useCallback, useMemo } from "react"
import { StreamPanelProps } from "../../../../types"
import StreamHeader from "./StreamHeader"
import { CheckboxChangeEvent } from "antd"

const StreamPanel: React.FC<StreamPanelProps> = ({
	stream,
	activeStreamData,
	setActiveStreamData,
	onStreamSelect,
	isSelected,
}) => {
	const toggle = useCallback(
		(e: CheckboxChangeEvent) => {
			const { checked } = e.target
			e.stopPropagation() // hack to prevent configuration triggers
			onStreamSelect?.(stream.stream.name, checked)
		},
		[stream, onStreamSelect],
	)

	const { header } = useMemo<
		| {
				header: JSX.Element
		  }
		| any
	>(() => {
		return {
			header: (
				<StreamHeader
					stream={stream}
					toggle={toggle}
					checked={isSelected}
					activeStreamData={activeStreamData}
					setActiveStreamData={setActiveStreamData}
				/>
			),
		}
	}, [stream, isSelected, activeStreamData, setActiveStreamData, toggle])

	return <div>{header}</div>
}

export default StreamPanel
