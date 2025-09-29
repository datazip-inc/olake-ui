import clsx from "clsx"

import { COLORS } from "../../../utils/constants"
import { FilterButtonProps } from "../../../types"

const FilterButton: React.FC<FilterButtonProps> = ({
	filter,
	selectedFilters,
	setSelectedFilters,
}) => {
	const isFilterSelected = selectedFilters.includes(filter)

	const handleFilterSelect = (filter: string) => {
		if (filter === "All tables") {
			setSelectedFilters(["All tables"])
			return
		}

		if (selectedFilters.includes(filter)) {
			setSelectedFilters(selectedFilters.filter(f => f !== filter))
			return
		}

		setSelectedFilters([
			...selectedFilters.filter(
				(selectedFilter: string) => selectedFilter !== "All tables",
			),
			filter,
		])
	}

	const buttonStyles = clsx(
		"cursor-pointer rounded-md border border-solid px-2 py-2 text-sm capitalize",
		isFilterSelected
			? "border-primary text-primary"
			: [
					`border-[${COLORS.unselected.border}]`,
					`text-[${COLORS.unselected.text}]`,
				],
	)

	return (
		<button
			type="button"
			className={buttonStyles}
			key={filter}
			onClick={() => handleFilterSelect(filter)}
		>
			{filter}
		</button>
	)
}

export default FilterButton
