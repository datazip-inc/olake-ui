import { useEffect, useState } from "react"
import { useSearchParams } from "react-router"

import {
	PAGE_SIZE,
	RUN_HISTORY_QUERY_PARAM,
	RUN_STATUS,
	RUN_STATUS_FILTER_STATUSES,
	RUN_TYPE_FILTER,
	RUN_TYPE_FILTER_PROCESS_TYPES,
} from "../constants"
import type { RunHistoryFilter, RunTypeFilter } from "../types"

const TYPE_FILTER_VALUES = new Set<string>(Object.values(RUN_TYPE_FILTER))

const getStatusFilter = (params: URLSearchParams): RunHistoryFilter =>
	params.get(RUN_HISTORY_QUERY_PARAM.STATUS) === RUN_STATUS.FAILED
		? "failed"
		: "all"

const getTypeFilter = (params: URLSearchParams): RunTypeFilter => {
	const value = params.get(RUN_HISTORY_QUERY_PARAM.TYPE)
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

// Manages Run History filters + pagination and syncs the `status` and `type`
// query params so the view survives a reload.
export const useRunHistoryFilters = () => {
	const [searchParams, setSearchParams] = useSearchParams()
	const activeFilter = getStatusFilter(searchParams)
	const activeTypeFilter = getTypeFilter(searchParams)
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = PAGE_SIZE

	const status = RUN_STATUS_FILTER_STATUSES[activeFilter]
	const runType = RUN_TYPE_FILTER_PROCESS_TYPES[activeTypeFilter]

	useEffect(() => {
		setCurrentPage(1)
	}, [activeFilter, activeTypeFilter])

	const setQueryParam = (key: string, value?: string) => {
		setCurrentPage(1)
		setSearchParams(prev => {
			const next = new URLSearchParams(prev)
			applyFilterToSearchParams(next, key, value)
			return next
		})
	}

	const handleFilterChange = (nextFilter: RunHistoryFilter) =>
		setQueryParam(
			RUN_HISTORY_QUERY_PARAM.STATUS,
			RUN_STATUS_FILTER_STATUSES[nextFilter],
		)

	const handleTypeFilterChange = (nextFilter: RunTypeFilter) =>
		setQueryParam(
			RUN_HISTORY_QUERY_PARAM.TYPE,
			nextFilter === RUN_TYPE_FILTER.ALL ? undefined : nextFilter,
		)

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
