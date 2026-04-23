import { FilterConfig, StreamData } from "@/modules/ingestion/common/types"

import DataFilterSectionView from "./DataFilterSectionView"
import { selectUseFilterConfig, useStreamSelectionStore } from "../../stores"

interface DataFilterSectionBulkProps {
	isDirty?: boolean
	bulkStream: StreamData
	bulkFilter: string
	onBulkFilterChange?: (filterString: string) => void
	bulkFilterConfig: FilterConfig | undefined
	onBulkFilterConfigChange?: (filterConfig: FilterConfig | undefined) => void
}

const DataFilterSectionBulk = ({
	isDirty,
	bulkStream,
	bulkFilter,
	onBulkFilterChange,
	bulkFilterConfig,
	onBulkFilterConfigChange,
}: DataFilterSectionBulkProps) => {
	const useFilterConfig = useStreamSelectionStore(selectUseFilterConfig)
	const isBulkDisabled =
		(bulkStream.stream.available_cursor_fields || []).length === 0

	return (
		<DataFilterSectionView
			stream={bulkStream}
			isSelected={true}
			isBulkDisabled={isBulkDisabled}
			isDirty={isDirty}
			filter={bulkFilter}
			filterConfig={bulkFilterConfig}
			useFilterConfig={useFilterConfig}
			streamFilterState={false}
			onFilterChange={onBulkFilterChange}
			onFilterConfigChange={onBulkFilterConfigChange}
		/>
	)
}

export default DataFilterSectionBulk
