import DataFilterSectionSingle from "./DataFilterSectionSingle"
import IngestionModeSectionSingle from "./IngestionModeSectionSingle"
import NormalizationSectionSingle from "./NormalizationSectionSingle"
import SyncModeSectionSingle from "./SyncModeSectionSingle"
import { CARD_STYLE } from "../../constants"

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
			</div>
			<NormalizationSectionSingle />
			<DataFilterSectionSingle />
		</div>
	)
}

export default ConfigTab
