import EmptyState from "../../common/components/EmptyState"
import { EMPTY_STATE_CONFIGS } from "../../../utils/emptyStateConfigs"

const JobEmptyState = ({
	handleCreateJob,
}: {
	handleCreateJob: () => void
}) => {
	return (
		<EmptyState
			config={EMPTY_STATE_CONFIGS.job}
			onButtonClick={handleCreateJob}
		/>
	)
}

export default JobEmptyState
