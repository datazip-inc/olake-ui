import { isKafkaSource } from "@/modules/ingestion/common/utils"

import DataFilterSectionSingle from "./DataFilterSectionSingle"
import DedupKeysSelectionSingle from "./DedupKeysSelectionSingle"
import IngestionModeSectionSingle from "./IngestionModeSectionSingle"
import NormalizationSectionSingle from "./NormalizationSectionSingle"
import SyncModeSectionSingle from "./SyncModeSectionSingle"
import UpsertTypeSectionSingle from "./UpsertTypeSectionSingle"
import { CARD_STYLE } from "../../constants"

interface ConfigTabProps {
	sourceType?: string
	sourceVersion?: string
	destinationType?: string
}

const ConfigTab = ({
	sourceType,
	sourceVersion,
	destinationType,
}: ConfigTabProps) => {
	return (
		<div className="flex flex-col gap-4">
			<div className={CARD_STYLE}>
				<SyncModeSectionSingle />
				<IngestionModeSectionSingle
					sourceType={sourceType}
					sourceVersion={sourceVersion}
					destinationType={destinationType}
				/>
				{isKafkaSource(sourceType) && (
					<DedupKeysSelectionSingle
						sourceType={sourceType}
						sourceVersion={sourceVersion}
					/>
				)}
				<UpsertTypeSectionSingle
					sourceType={sourceType}
					destinationType={destinationType}
				/>
			</div>
			<NormalizationSectionSingle />
			<DataFilterSectionSingle />
		</div>
	)
}

export default ConfigTab
