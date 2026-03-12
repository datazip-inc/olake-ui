import DataFilterSection from "./DataFilterSection"
import IngestionModeSection from "./IngestionModeSection"
import NormalizationSection from "./NormalizationSection"
import SyncModeSection from "./SyncModeSection"
import { CARD_STYLE } from "../../constants"

interface ConfigTabProps {
	sourceType?: string
	destinationType?: string
}

const ConfigTab = ({ sourceType, destinationType }: ConfigTabProps) => {
	return (
		<div className="flex flex-col gap-4">
			<div className={CARD_STYLE}>
				<SyncModeSection />
				<IngestionModeSection
					sourceType={sourceType}
					destinationType={destinationType}
				/>
			</div>
			<NormalizationSection />
			<DataFilterSection />
		</div>
	)
}

export default ConfigTab
