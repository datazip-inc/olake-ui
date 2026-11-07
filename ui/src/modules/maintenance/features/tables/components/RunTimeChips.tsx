type RunTimeChipsProps = {
	runs: string[]
}

/** Upcoming run times laid out as equal-width chips on one line. */
const RunTimeChips: React.FC<RunTimeChipsProps> = ({ runs }) => (
	<div className="flex w-full items-center gap-2">
		{runs.map(run => (
			<span
				key={run}
				className="min-w-0 flex-1 whitespace-nowrap rounded-[6px] bg-olake-surface-muted px-2 py-1 text-center text-sm leading-[22px] text-olake-text-secondary"
			>
				{run}
			</span>
		))}
	</div>
)

export default RunTimeChips
