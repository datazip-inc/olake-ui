import { Tooltip } from "antd"

const RenderTypeItems = ({ initialList, item }: any) => {
	if (typeof initialList[item]?.type === "string") {
		return (
			<div className="flex justify-start">
				<span className="whitespace-nowrap rounded-md bg-primary-200 px-2 py-1 text-xs uppercase text-primary">
					{initialList[item].type}
				</span>
			</div>
		)
	}
	if (Array.isArray(initialList[item]?.type)) {
		const types = initialList[item].type
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
	return <></>
}

export default RenderTypeItems
