import { useEffect, useMemo, useState } from "react"

import { PAGE_SIZE } from "../constants"
import type { FilterKey, Table } from "../types"

type UseFilteredTablesReturn = {
	searchTerm: string
	setSearchTerm: (term: string) => void
	activeFilter: FilterKey
	setActiveFilter: (filter: FilterKey) => void
	currentPage: number
	setCurrentPage: (page: number) => void
	paginatedRows: Table[]
	totalPages: number
}

/** Filters, searches, and paginates a table list. Resets to page 1 when filters change. */
export function useFilteredTables(tables: Table[]): UseFilteredTablesReturn {
	const [searchTerm, setSearchTerm] = useState("")
	const [activeFilter, setActiveFilter] = useState<FilterKey>("all")
	const [currentPage, setCurrentPage] = useState(1)

	const filteredRows = useMemo(
		() =>
			tables.filter(row => {
				const matchesSearch = row.name
					.toLowerCase()
					.includes(searchTerm.toLowerCase())
				const matchesFilter =
					activeFilter === "all"
						? true
						: activeFilter === "olake"
							? row.byOLake
							: !row.byOLake
				return matchesSearch && matchesFilter
			}),
		[tables, searchTerm, activeFilter],
	)

	useEffect(() => {
		setCurrentPage(1)
	}, [searchTerm, activeFilter])

	const totalPages = Math.max(1, Math.ceil(filteredRows.length / PAGE_SIZE))
	const paginatedRows = filteredRows.slice(
		(currentPage - 1) * PAGE_SIZE,
		currentPage * PAGE_SIZE,
	)

	return {
		searchTerm,
		setSearchTerm,
		activeFilter,
		setActiveFilter,
		currentPage,
		setCurrentPage,
		paginatedRows,
		totalPages,
	}
}
