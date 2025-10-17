import EmptyState from "../../common/components/EmptyState"
import { EMPTY_STATE_CONFIGS } from "../../../utils/emptyStateConfigs"

const DestinationEmptyState = ({
	handleCreateDestination,
}: {
	handleCreateDestination: () => void
}) => {
	return (
		<EmptyState
			config={EMPTY_STATE_CONFIGS.destination}
			onButtonClick={handleCreateDestination}
		/>
	)
}

export default DestinationEmptyState
