import { useEffect, useState } from "react"
import { useSearchParams } from "react-router"

import { RUN_STATUS, RUN_TYPE_FILTER } from "../constants"
import type { RunHistoryFilter, RunTypeFilter } from "../types"

const RUNS_PAGE_SIZE = 6
const STATUS_QUERY_PARAM_KEY = "status"
const TYPE_QUERY_PARAM_KEY = "type"

const FILTER_TO_STATUS_QUERY_VALUE: Record<
	RunHistoryFilter,
	string | undefined
> = {
	all: undefined,
	failed: RUN_STATUS.FAILED,
}

const TYPE_FILTER_VALUES = new Set<string>(Object.values(RUN_TYPE_FILTER))

const getStatusFilter = (params: URLSearchParams): RunHistoryFilter =>
	params.get(STATUS_QUERY_PARAM_KEY) === RUN_STATUS.FAILED ? "failed" : "all"

const getTypeFilter = (params: URLSearchParams): RunTypeFilter => {
	const value = params.get(TYPE_QUERY_PARAM_KEY)
	return value && TYPE_FILTER_VALUES.has(value)
		? (value as RunTypeFilter)
		: RUN_TYPE_FILTER.ALL
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

/**
 * Manages Run History filters + pagination and syncs the `status` and `type`
 * query params so the view survives a reload.
 */
export const useRunHistoryFilters = () => {
	const [searchParams, setSearchParams] = useSearchParams()
	const activeFilter = getStatusFilter(searchParams)
	const activeTypeFilter = getTypeFilter(searchParams)
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = RUNS_PAGE_SIZE

	const status = FILTER_TO_STATUS_QUERY_VALUE[activeFilter]
	// "all" is the endpoint default, so it is not sent.
	const runType =
		activeTypeFilter === RUN_TYPE_FILTER.ALL ? undefined : activeTypeFilter

	useEffect(() => {
		setCurrentPage(1)
	}, [activeFilter, activeTypeFilter])

	const handleFilterChange = (nextFilter: RunHistoryFilter) => {
		setCurrentPage(1)
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

	const handleTypeFilterChange = (nextFilter: RunTypeFilter) => {
		setCurrentPage(1)
		setSearchParams(prev => {
			const next = new URLSearchParams(prev)
			applyFilterToSearchParams(
				next,
				TYPE_QUERY_PARAM_KEY,
				nextFilter === RUN_TYPE_FILTER.ALL ? undefined : nextFilter,
			)
			return next
		})
	}

	return {
		activeFilter,
		activeTypeFilter,
		currentPage,
		setCurrentPage,
		pageSize,
		status,
		runType,
		handleFilterChange,
		handleTypeFilterChange,
	} as const
}
