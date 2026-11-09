import clsx from "clsx"

type FilterChipsProps<K extends string> = {
	options: Array<{ key: K; label: string }>
	value: K
	onChange: (key: K) => void
}

const FilterChips = <K extends string>({
	options,
	value,
	onChange,
}: FilterChipsProps<K>) => (
	<div className="flex items-center gap-2">
		{options.map(option => (
			<button
				key={option.key}
				type="button"
				onClick={() => onChange(option.key)}
				className={clsx(
					"h-8 whitespace-nowrap rounded-md border border-olake-border px-3 text-sm leading-5 text-olake-text-secondary",
					value === option.key ? "bg-olake-surface-muted" : "bg-white",
				)}
			>
				{option.label}
			</button>
		))}
	</div>
)

export default FilterChips
