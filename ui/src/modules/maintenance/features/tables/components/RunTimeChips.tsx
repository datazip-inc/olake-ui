type RunTimeChipsProps = {
	runs: string[]
	/** Keeps React keys unique when several chip rows share a parent. */
	keyPrefix: string
}

/** Upcoming run times laid out as equal-width chips on one line. */
const RunTimeChips: React.FC<RunTimeChipsProps> = ({ runs, keyPrefix }) => (
	<div className="flex w-full items-center gap-2">
		{runs.map(run => (
			<span
				key={`${keyPrefix}-${run}`}
				className="min-w-0 flex-1 whitespace-nowrap rounded-[6px] bg-olake-surface-muted px-2 py-1 text-center text-sm leading-[22px] text-olake-text-secondary"
			>
				{run}
			</span>
		))}
	</div>
)

export default RunTimeChips
