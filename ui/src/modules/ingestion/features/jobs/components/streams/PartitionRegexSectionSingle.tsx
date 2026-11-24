import { DESTINATION_INTERNAL_TYPES } from "@/modules/ingestion/common/constants"
import { SyncMode } from "@/modules/ingestion/common/types"

import PartitionRegexSectionView from "./PartitionRegexSectionView"
import { useJobConfigurationStore, useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores"
import { partitionEvolutionBlocker } from "../../utils"

interface PartitionRegexSectionSingleProps {
	destinationType?: string
}

const PartitionRegexSectionSingle = ({
	destinationType,
}: PartitionRegexSectionSingleProps) => {
	const updatePartitionRegex = useStreamSelectionStore(
		state => state.updatePartitionRegex,
	)
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)
	const storeIsSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, storeStream),
	)
	const { isEditMode, selectedDestination, advancedSettings } =
		useJobConfigurationStore()

	if (!storeStream || !storeSelectedStream) return null

	// Only an existing Iceberg table can keep its data across a partitioning change
	const changeImpact = (() => {
		if (!isEditMode || destinationType !== DESTINATION_INTERNAL_TYPES.ICEBERG) {
			return undefined
		}
		const syncMode =
			storeSelectedStream.sync_mode ?? storeStream.stream.sync_mode
		const appendOnly =
			!!storeSelectedStream.append_mode || syncMode === SyncMode.FULL_REFRESH
		if (!appendOnly) {
			return "Changing partitioning clears this stream's destination data: upsert mode can't change the partition layout in place. Turn on append mode to keep existing data."
		}
		const blocker = partitionEvolutionBlocker(
			selectedDestination?.config,
			advancedSettings?.target_query_engines,
		)
		if (blocker) {
			return `Changing partitioning clears this stream's destination data: ${blocker} doesn't support partition evolution.`
		}
		return "Changing partitioning keeps existing data; new files use the new layout."
	})()

	return (
		<PartitionRegexSectionView
			destinationType={destinationType}
			isSelected={storeIsSelected}
			activePartitionRegex={storeSelectedStream.partition_regex || ""}
			changeImpact={changeImpact}
			onChange={regex =>
				updatePartitionRegex(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					regex,
				)
			}
		/>
	)
}

export default PartitionRegexSectionSingle
