const RenderTypeItems = ({ initialList, item }: any) => {
	if (typeof initialList[item]?.type === "string") {
		return (
			<div>
				<span className="rounded-md bg-primary-200 px-2 py-1 text-sm uppercase text-primary">
					{initialList[item].type}
				</span>
			</div>
		)
	}
	if (Array.isArray(initialList[item]?.type)) {
		return (
			<div className="flex items-center gap-2">
				{initialList[item].type.map((val: any) => (
					<span
						key={val}
						className="rounded-md bg-primary-200 px-2 py-1 text-sm uppercase text-primary"
					>
						{val}
					</span>
				))}
			</div>
		)
	}
	return <></>
}

export default RenderTypeItems
