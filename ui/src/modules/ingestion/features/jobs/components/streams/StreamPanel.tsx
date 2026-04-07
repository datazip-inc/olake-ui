import { CheckboxChangeEvent } from "antd"
import { useCallback, useState, useEffect } from "react"

import StreamHeader from "./StreamHeader"
import { StreamPanelProps } from "../../types"

const StreamPanel: React.FC<StreamPanelProps> = ({
	stream,
	onStreamSelect,
	isSelected,
}) => {
	const [checked, setChecked] = useState(isSelected)

	useEffect(() => {
		setChecked(isSelected)
	}, [isSelected])

	const toggle = useCallback(
		(e: CheckboxChangeEvent) => {
			const { checked } = e.target
			e.stopPropagation()
			setChecked(checked)
			onStreamSelect?.(stream.stream.name, checked)
		},
		[stream, onStreamSelect],
	)

	return (
		<div>
			<StreamHeader
				stream={stream}
				toggle={toggle}
				checked={checked}
			/>
		</div>
	)
}

export default StreamPanel
