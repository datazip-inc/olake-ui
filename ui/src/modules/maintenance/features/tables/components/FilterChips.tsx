import clsx from "clsx"

type FilterChipsProps<K extends string> = {
	options: ReadonlyArray<{ key: K; label: string }>
	value: K
	onChange: (key: K) => void
	/** sm = 32px (run history), md = 36px (tables list) */
	size?: "sm" | "md"
}

/** Row of mutually exclusive filter buttons, shared by the tables list and run history. */
const FilterChips = <K extends string>({
	options,
	value,
	onChange,
	size = "md",
}: FilterChipsProps<K>) => (
	<div className="flex items-center gap-2">
		{options.map(option => (
			<button
				key={option.key}
				type="button"
				onClick={() => onChange(option.key)}
				className={clsx(
					"whitespace-nowrap rounded-md border border-olake-border px-3 text-sm leading-5 text-olake-text-secondary",
					size === "sm" ? "h-8" : "h-9",
					value === option.key ? "bg-olake-surface-muted" : "bg-white",
				)}
			>
				{option.label}
			</button>
		))}
	</div>
)

export default FilterChips
