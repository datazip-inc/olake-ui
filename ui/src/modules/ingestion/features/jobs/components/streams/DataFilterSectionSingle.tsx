import DataFilterSectionView from "./DataFilterSectionView"
import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectIsStreamEnabled,
	selectStreamFilterState,
	selectUseFilterConfig,
	useStreamSelectionStore,
} from "../../stores"

const DataFilterSectionSingle = () => {
	const updateFilter = useStreamSelectionStore(state => state.updateFilter)
	const updateFilterConfig = useStreamSelectionStore(
		state => state.updateFilterConfig,
	)
	const setStreamFilterState = useStreamSelectionStore(
		state => state.setStreamFilterState,
	)
	const useFilterConfig = useStreamSelectionStore(selectUseFilterConfig)
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)
	const storeIsSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, storeStream),
	)

	// Unique stream key to differentiate a stream with same name and different namespace
	const streamKey = storeStream
		? `${storeStream.stream.namespace || ""}_${storeStream.stream.name}`
		: ""
	const streamFilterState = useStreamSelectionStore(
		selectStreamFilterState(streamKey),
	)

	if (!storeStream || !storeSelectedStream) return null

	return (
		<DataFilterSectionView
			stream={storeStream}
			isSelected={storeIsSelected}
			isBulkDisabled={false}
			filter={storeSelectedStream.filter || ""}
			filterConfig={storeSelectedStream.filter_config}
			useFilterConfig={useFilterConfig}
			streamFilterState={streamFilterState}
			onFilterChange={filterString =>
				updateFilter(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					filterString,
				)
			}
			onFilterConfigChange={fc =>
				updateFilterConfig(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					fc,
				)
			}
			onSetStreamFilterState={enabled =>
				setStreamFilterState(streamKey, enabled)
			}
		/>
	)
}

export default DataFilterSectionSingle
