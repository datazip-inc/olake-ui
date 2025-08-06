import { FilterButtonProps } from "../../../types"
import { COLORS } from "../../../utils/constants"

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

	const buttonStyles = `cursor-pointer rounded-md border border-solid px-2 py-2 text-sm capitalize ${
		isFilterSelected
			? `border-[${COLORS.selected.border}] text-[${COLORS.selected.text}]`
			: `border-[${COLORS.unselected.border}] text-[${COLORS.unselected.text}]`
	}`

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
