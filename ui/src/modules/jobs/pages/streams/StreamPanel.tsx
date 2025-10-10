import { useCallback, useMemo, useState, useEffect } from "react"
import { CheckboxChangeEvent } from "antd"

import { StreamPanelProps } from "@app-types/index"
import StreamHeader from "./StreamHeader"

const StreamPanel: React.FC<StreamPanelProps> = ({
	stream,
	activeStreamData,
	setActiveStreamData,
	onStreamSelect,
	isSelected,
}) => {
	const [checked, setChecked] = useState(isSelected)

	// Update checked state when isSelected prop changes
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
					checked={checked}
					activeStreamData={activeStreamData}
					setActiveStreamData={setActiveStreamData}
				/>
			),
		}
	}, [stream, checked, activeStreamData, setActiveStreamData, toggle])

	return <div>{header}</div>
}

export default StreamPanel
