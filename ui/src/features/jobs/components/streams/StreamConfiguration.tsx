import { useEffect, useState } from "react"

import { SelectedColumns } from "../../../../common/types"
import { DESTINATION_INTERNAL_TYPES } from "@/common/constants/constants"
import { StreamConfigurationProps } from "../../types"

import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
} from "../../stores/streamSelectionStore"

import StreamConfigHeader from "./StreamConfigHeader"
import ConfigTab from "./ConfigTab"
import PartitionRegexSection from "./PartitionRegexSection"
import StreamsSchema from "./StreamsSchema"

const StreamConfiguration = ({
	destinationType = DESTINATION_INTERNAL_TYPES.S3,
	sourceType,
}: StreamConfigurationProps) => {
	const store = useStreamSelectionStore()
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)

	const [activeTab, setActiveTab] = useState("config")

	// Reset to config tab whenever the viewed stream changes
	useEffect(() => {
		setActiveTab("config")
	}, [stream?.stream.name, stream?.stream.namespace])

	if (!stream || !selectedStream) return null

	// Guard: old driver versions don't have selected_columns
	const handleSelectedColumnsChange = (selected_columns: SelectedColumns) => {
		if (!selectedStream.selected_columns) return
		store.updateSelectedColumns(
			stream.stream.name,
			stream.stream.namespace || "",
			selected_columns,
		)
	}

	return (
		<div>
			<StreamConfigHeader
				activeTab={activeTab}
				onTabChange={setActiveTab}
			/>

			{activeTab === "config" && (
				<ConfigTab
					sourceType={sourceType}
					destinationType={destinationType}
				/>
			)}
			{activeTab === "schema" && store.streamsData && (
				<StreamsSchema
					initialStreamsData={stream}
					initialSelectedStream={selectedStream}
					onSelectedColumnChange={handleSelectedColumnsChange}
				/>
			)}
			{activeTab === "partitioning" && (
				<PartitionRegexSection destinationType={destinationType} />
			)}
		</div>
	)
}

export default StreamConfiguration
