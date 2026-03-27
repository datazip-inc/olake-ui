import { useEffect, useMemo, useRef, useState } from "react"

export const DEFAULT_PAGE_SIZE = 10

type UsePaginatedSearchOptions<T, F extends string = string> = {
	// Full row dataset to process
	rows: T[]
	// Items per page; defaults to DEFAULT_PAGE_SIZE
	pageSize?: number
	// Returns true when a row matches the search term
	searchFn: (row: T, searchTerm: string) => boolean
	// Returns true when a row matches the active filter (optional)
	filterFn?: (row: T, activeFilter: F) => boolean
	// Initial filter value; defaults to empty string
	initialFilter?: F
}

export type UsePaginatedSearchReturn<T, F extends string = string> = {
	searchTerm: string
	setSearchTerm: (term: string) => void
	activeFilter: F
	setActiveFilter: (filter: F) => void
	currentPage: number
	setCurrentPage: (page: number) => void
	// All rows after search + filter, before pagination
	filteredRows: T[]
	// Current page slice of filteredRows
	paginatedRows: T[]
	totalPages: number
}

/** Use for list pages that need client-side search, optional filtering, and pagination state. */
export function usePaginatedSearch<T, F extends string = string>({
	rows,
	pageSize = DEFAULT_PAGE_SIZE,
	searchFn,
	filterFn,
	initialFilter = "" as F,
}: UsePaginatedSearchOptions<T, F>): UsePaginatedSearchReturn<T, F> {
	const [searchTerm, setSearchTerm] = useState("")
	const [activeFilter, setActiveFilter] = useState<F>(initialFilter)
	const [currentPage, setCurrentPage] = useState(1)
	const normalizedSearchTerm = searchTerm.trim().toLowerCase()

	const filteredRows = useMemo(
		() =>
			rows.filter(row => {
				const matchesSearch =
					!normalizedSearchTerm || searchFn(row, normalizedSearchTerm)
				const matchesFilter = !filterFn || filterFn(row, activeFilter)
				return matchesSearch && matchesFilter
			}),
		// searchFn / filterFn should be stable references (module-level or useCallback);
		// intentionally omitted from deps to avoid spurious recomputes on inline fns.
		[rows, normalizedSearchTerm, activeFilter],
	)

	const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize))

	// Track previous query to distinguish "query changed" (→ page 1) from
	// "underlying data changed" (→ clamp to last valid page).
	const prevQueryRef = useRef<{ searchTerm: string; activeFilter: F }>({
		searchTerm: "",
		activeFilter: initialFilter,
	})

	useEffect(() => {
		const prev = prevQueryRef.current
		const queryChanged =
			prev.searchTerm !== searchTerm || prev.activeFilter !== activeFilter
		prevQueryRef.current = { searchTerm, activeFilter }

		setCurrentPage(page => {
			if (queryChanged) return 1
			return Math.min(page, totalPages)
		})
	}, [searchTerm, activeFilter, totalPages])

	const paginatedRows = useMemo(
		() =>
			filteredRows.slice((currentPage - 1) * pageSize, currentPage * pageSize),
		[filteredRows, currentPage, pageSize],
	)

	return {
		searchTerm,
		setSearchTerm,
		activeFilter,
		setActiveFilter,
		currentPage,
		setCurrentPage,
		filteredRows,
		paginatedRows,
		totalPages,
	}
}
