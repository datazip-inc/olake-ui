/**
 * ArrayFieldTemplate is a component that renders an array of items this overrides the default array field template of rjsf
 */

import { ArrayFieldTemplateProps } from "@rjsf/utils"
import { Button } from "antd"
import { PlusIcon, TrashIcon } from "@phosphor-icons/react"

const ArrayFieldTemplate = (props: ArrayFieldTemplateProps) => {
	const { items, canAdd, onAddClick } = props

	return (
		<div className="mb-2">
			<div className="space-y-1">
				{items &&
					items.map(item => (
						<div
							key={item.key}
							className="flex items-center gap-2"
						>
							<div className="array-item-field flex-1">{item.children}</div>
							{item.hasRemove && (
								<Button
									aria-label="Remove item"
									type="text"
									danger
									onClick={item.onDropIndexClick(item.index)}
									icon={<TrashIcon className="text-red-500" />}
								/>
							)}
						</div>
					))}
			</div>

			{canAdd && (
				<div className="mr-[6%] mt-2 flex justify-end">
					<Button
						aria-label="Add item"
						type="primary"
						size="middle"
						className="px-3"
						onClick={onAddClick}
						icon={<PlusIcon className="size-4" />}
					>
						Add
					</Button>
				</div>
			)}
		</div>
	)
}

export default ArrayFieldTemplate
