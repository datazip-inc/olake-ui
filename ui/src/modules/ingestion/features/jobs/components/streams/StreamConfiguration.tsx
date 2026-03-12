import { useEffect, useState } from "react"

import { DESTINATION_INTERNAL_TYPES } from "@/modules/ingestion/common/constants/constants"

import ConfigTab from "./ConfigTab"
import PartitionRegexSection from "./PartitionRegexSection"
import StreamConfigHeader from "./StreamConfigHeader"
import StreamsSchema from "./StreamsSchema"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectStreamsData,
	useStreamSelectionStore,
} from "../../stores"
import { StreamConfigurationProps } from "../../types"

const StreamConfiguration = ({
	destinationType = DESTINATION_INTERNAL_TYPES.S3,
	sourceType,
}: StreamConfigurationProps) => {
	const streamsData = useStreamSelectionStore(selectStreamsData)
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)

	const [activeTab, setActiveTab] = useState("config")

	// Reset to config tab whenever the viewed stream changes
	useEffect(() => {
		setActiveTab("config")
	}, [stream?.stream.name, stream?.stream.namespace])

	if (!stream || !selectedStream) return null

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
			{activeTab === "schema" && streamsData && <StreamsSchema />}
			{activeTab === "partitioning" && (
				<PartitionRegexSection destinationType={destinationType} />
			)}
		</div>
	)
}

export default StreamConfiguration
