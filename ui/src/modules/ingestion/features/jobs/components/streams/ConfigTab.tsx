import { SOURCE_INTERNAL_TYPES } from "@/modules/ingestion/common/constants"
import { normalizeConnectorType } from "@/modules/ingestion/common/utils"

import DataFilterSectionSingle from "./DataFilterSectionSingle"
import IngestionModeSectionSingle from "./IngestionModeSectionSingle"
import NormalizationSectionSingle from "./NormalizationSectionSingle"
import SyncModeSectionSingle from "./SyncModeSectionSingle"
import { CARD_STYLE } from "../../constants"
import DedupKeysSectionSingle from "./dedupKeys/dedupKeysSectionSingle"

interface ConfigTabProps {
	sourceType?: string
	destinationType?: string
}

const ConfigTab = ({ sourceType, destinationType }: ConfigTabProps) => {
	return (
		<div className="flex flex-col gap-4">
			<div className={CARD_STYLE}>
				<SyncModeSectionSingle />
				<IngestionModeSectionSingle
					sourceType={sourceType}
					destinationType={destinationType}
				/>
				{!!sourceType &&
					normalizeConnectorType(sourceType).toLowerCase() ===
						SOURCE_INTERNAL_TYPES.KAFKA && <DedupKeysSectionSingle />}
			</div>
			<NormalizationSectionSingle />
			<DataFilterSectionSingle />
		</div>
	)
}

export default ConfigTab
