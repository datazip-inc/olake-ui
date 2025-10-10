import EmptyState from "../../common/components/EmptyState"
import { EMPTY_STATE_CONFIGS } from "../../../utils/emptyStateConfigs"

const SourceEmptyState = ({
	handleCreateSource,
}: {
	handleCreateSource: () => void
}) => {
	return (
		<EmptyState
			config={EMPTY_STATE_CONFIGS.source}
			onButtonClick={handleCreateSource}
		/>
	)
}

export default SourceEmptyState
