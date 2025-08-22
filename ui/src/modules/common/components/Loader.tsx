// components/Loader.tsx
import React from "react"
import LoaderSquares from "../../../assets/loader-squares.gif"
import { LoaderComponentSize } from "../../../utils/constants"

type LoaderProps = {
	size?: "large" | "small"
	tip?: string
}

const Loader: React.FC<LoaderProps> = ({ size = "small", tip }) => {
	const width =
		size === "large" ? LoaderComponentSize.large : LoaderComponentSize.small

	return (
		<div className="flex h-full w-full items-center justify-center">
			<img
				width={width}
				src={LoaderSquares}
				alt="Loading..."
			/>
			{tip && <span className="invisible">{tip}</span>}
		</div>
	)
}

export default Loader
