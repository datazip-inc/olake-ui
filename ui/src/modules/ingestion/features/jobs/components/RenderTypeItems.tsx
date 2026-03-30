import { Tooltip } from "antd"

const RenderTypeItems = ({ initialList, item }: any) => {
	const types = initialList[item]?.type
	if (!Array.isArray(types) || types.length === 0) return <></>

	const [firstType, ...remainingTypes] = types
	const hasRemaining = remainingTypes.length > 0

	return (
		<div className="flex items-center justify-start">
			<Tooltip
				title={hasRemaining ? remainingTypes.join(", ").toUpperCase() : ""}
				placement="top"
			>
				<span className="whitespace-nowrap rounded-md bg-primary-200 px-2 py-1 text-xs uppercase text-primary">
					{firstType}
					{hasRemaining && ` +${remainingTypes.length}`}
				</span>
			</Tooltip>
		</div>
	)
}

export default RenderTypeItems
