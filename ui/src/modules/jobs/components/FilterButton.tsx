type FilterButtonProps = {
	filter: string
	selectedFilters: string[]
	setSelectedFilters: (filters: string[]) => void
}

const FilterButton: React.FC<FilterButtonProps> = ({
	filter,
	selectedFilters,
	setSelectedFilters,
}) => {
	const isSyncModeSelected = selectedFilters.includes(filter)

	const handleFilterSelect = (filter: string) => {
		if (filter === "All tables") {
			setSelectedFilters(["All tables"])
		} else if (selectedFilters.includes(filter)) {
			setSelectedFilters(selectedFilters.filter(f => f !== filter))
		} else {
			setSelectedFilters([
				...selectedFilters.filter(
					(selectedFilter: string) => selectedFilter !== "All tables",
				),
				filter,
			])
		}
	}
	return (
		<button
			type="button"
			className={`cursor-pointer rounded-[6px] border border-solid px-2 py-2 text-sm capitalize ${isSyncModeSelected ? "border-[#203FDD] text-[#203FDD]" : "border-[#D9D9D9] text-[#575757]"}`}
			key={filter}
			onClick={() => handleFilterSelect(filter)}
		>
			{filter}
		</button>
	)
}

export default FilterButton
