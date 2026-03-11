import { useCallback, useMemo, useState, useEffect } from "react"
import { CheckboxChangeEvent } from "antd"

import { StreamPanelProps } from "../../types"
import StreamHeader from "./StreamHeader"

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

	const { header } = useMemo<{ header: JSX.Element } | any>(
		() => ({
			header: (
				<StreamHeader
					stream={stream}
					toggle={toggle}
					checked={checked}
				/>
			),
		}),
		[stream, checked, toggle],
	)

	return <div>{header}</div>
}

export default StreamPanel
