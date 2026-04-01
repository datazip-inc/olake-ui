import { useState } from "react"
import { useSearchParams } from "react-router-dom"

import { RUN_STATUS } from "../constants"

const RUNS_PAGE_SIZE = 6
const STATUS_QUERY_PARAM_KEY = "status"

export type RunHistoryFilter = "all" | "failed"

const FILTER_TO_STATUS_QUERY_VALUE: Record<
	RunHistoryFilter,
	string | undefined
> = {
	all: undefined,
	failed: RUN_STATUS.FAILED,
}

const getFilterFromSearchParams = (
	params: URLSearchParams,
	queryParamKey: string,
	valueToFilterMap: Record<string, RunHistoryFilter>,
	fallbackFilter: RunHistoryFilter,
): RunHistoryFilter => {
	const queryValue = params.get(queryParamKey)
	if (!queryValue) return fallbackFilter
	return valueToFilterMap[queryValue] ?? fallbackFilter
}

const applyFilterToSearchParams = (
	params: URLSearchParams,
	queryParamKey: string,
	queryValue?: string,
) => {
	if (queryValue) {
		params.set(queryParamKey, queryValue)
		return
	}
	params.delete(queryParamKey)
}
/** Manages Run History filter + pagination state and syncs the `status` query param for persistence across reloads. */
export const useRunHistoryFilters = () => {
	const [searchParams, setSearchParams] = useSearchParams()
	const initialFilter = getFilterFromSearchParams(
		searchParams,
		STATUS_QUERY_PARAM_KEY,
		{ [RUN_STATUS.FAILED]: "failed" },
		"all",
	)
	const [activeFilter, setActiveFilter] =
		useState<RunHistoryFilter>(initialFilter)
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = RUNS_PAGE_SIZE

	const status = FILTER_TO_STATUS_QUERY_VALUE[activeFilter]

	const handleFilterChange = (nextFilter: RunHistoryFilter) => {
		setCurrentPage(1)
		setActiveFilter(nextFilter)
		setSearchParams(prev => {
			const next = new URLSearchParams(prev)
			applyFilterToSearchParams(
				next,
				STATUS_QUERY_PARAM_KEY,
				FILTER_TO_STATUS_QUERY_VALUE[nextFilter],
			)
			return next
		})
	}

	return {
		activeFilter,
		currentPage,
		setCurrentPage,
		pageSize,
		status,
		handleFilterChange,
	} as const
}
