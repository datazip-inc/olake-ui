import { useEffect, useState, useMemo } from "react"
import { Input, Empty } from "antd"
import FilterButton from "../components/FilterButton"
import { mockStreamData } from "../../../api/mockData"
import StreamsCollapsibleList from "./streams/StreamsCollapsibleList"
import { CheckboxChangeEvent } from "antd/es/checkbox/Checkbox"
import { StreamData } from "../../../types"
import StreamConfiguration from "./streams/StreamConfiguration"
import StepTitle from "../../common/components/StepTitle"

interface SchemaConfigurationProps {
	selectedStreams: string[]
	setSelectedStreams: React.Dispatch<React.SetStateAction<string[]>>
	stepNumber?: number | string
	stepTitle?: string
}

const SchemaConfiguration: React.FC<SchemaConfigurationProps> = ({
	selectedStreams,
	setSelectedStreams,
	stepNumber = 3,
	stepTitle = "Schema evaluation",
}) => {
	const [searchText, setSearchText] = useState("")
	const [selectedFilters, setSelectedFilters] = useState<string[]>([
		"All tables",
	])
	const [activeStreamData, setActiveStreamData] = useState<StreamData | null>(
		null,
	)

	const filters = [
		"All tables",
		"CDC",
		"Full refresh",
		"Selected",
		"Not selected",
	]

	useEffect(() => {
		if (selectedFilters.length === 0) {
			setSelectedFilters(["All tables"])
		}
	}, [selectedFilters])

	const handleToggleAllStreams = (e: CheckboxChangeEvent) => {
		const { checked } = e.target
		const filteredStreamNames = new Set(
			filteredStreams.map(stream => stream.stream.name),
		)

		setSelectedStreams(prev =>
			checked
				? [...new Set([...prev, ...filteredStreamNames])]
				: prev.filter(name => !filteredStreamNames.has(name)),
		)
	}

	const handleStreamSelect = (streamName: string, checked: boolean) => {
		setSelectedStreams(prev =>
			checked
				? [...prev, streamName]
				: prev.filter(name => name !== streamName),
		)
	}

	// Filter streams based on selected filters and search text
	const filteredStreams = useMemo(() => {
		let filtered = [...mockStreamData]

		// Apply search filter
		if (searchText) {
			filtered = filtered.filter(stream =>
				stream.stream.name.toLowerCase().includes(searchText.toLowerCase()),
			)
		}

		// Return all filtered streams if "All tables" is selected
		if (selectedFilters.includes("All tables")) {
			return filtered
		}

		// Helper function to check if a stream matches selection criteria
		const matchesSelectionCriteria = (stream: StreamData) => {
			const isSelected = selectedStreams.includes(stream.stream.name)
			const hasSelectedFilter = selectedFilters.includes("Selected")
			const hasNotSelectedFilter = selectedFilters.includes("Not selected")

			if (hasSelectedFilter && hasNotSelectedFilter) return true
			if (hasSelectedFilter) return isSelected
			if (hasNotSelectedFilter) return !isSelected
			return true
		}

		// Filter streams based on sync mode and selection criteria
		return filtered.filter(stream => {
			const isCDC =
				selectedFilters.includes("CDC") && stream.sync_mode === "cdc"
			const isFullRefresh =
				selectedFilters.includes("Full refresh") &&
				stream.sync_mode === "full_refresh"

			// If no sync mode filters are active, only check selection criteria
			if (
				!selectedFilters.some(
					filter => filter === "CDC" || filter === "Full refresh",
				)
			) {
				return matchesSelectionCriteria(stream)
			}

			// Check both sync mode and selection criteria
			return (isCDC || isFullRefresh) && matchesSelectionCriteria(stream)
		})
	}, [mockStreamData, searchText, selectedFilters, selectedStreams])

	const { Search } = Input

	return (
		<div className="mb-4 p-6">
			{stepNumber && stepTitle && (
				<StepTitle
					stepNumber={stepNumber}
					stepTitle={stepTitle}
				/>
			)}

			<div className="mb-4 flex items-center justify-between">
				<Search
					placeholder="Search streams"
					allowClear
					className="w-2/4"
					value={searchText}
					onChange={e => setSearchText(e.target.value)}
				/>
				<div className="flex space-x-2">
					{filters.map(filter => (
						<FilterButton
							key={filter}
							filter={filter}
							selectedFilters={selectedFilters}
							setSelectedFilters={setSelectedFilters}
						/>
					))}
				</div>
			</div>

			<div className="flex">
				<div className={`${activeStreamData ? "w-1/2" : "w-full"} `}>
					{filteredStreams?.length ? (
						<StreamsCollapsibleList
							streamsToDisplay={filteredStreams}
							allChecked={filteredStreams.every(stream =>
								selectedStreams.includes(stream.stream.name),
							)}
							handleToggleAllStreams={handleToggleAllStreams}
							activeStreamData={activeStreamData}
							setActiveStreamData={setActiveStreamData}
							selectedStreams={selectedStreams}
							onStreamSelect={handleStreamSelect}
						/>
					) : (
						<>
							<Empty className="flex h-full flex-col items-center justify-center" />
						</>
					)}
				</div>

				{activeStreamData && (
					<div className="mx-4 flex h-full w-1/2 flex-col rounded-xl border bg-[#ffffff] p-4 transition-all duration-150 ease-linear">
						<StreamConfiguration stream={activeStreamData} />
					</div>
				)}
			</div>
		</div>
	)
}

export default SchemaConfiguration
