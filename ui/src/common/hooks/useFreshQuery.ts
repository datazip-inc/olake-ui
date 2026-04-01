import {
	useQuery,
	type UseQueryOptions,
	type QueryKey,
} from "@tanstack/react-query"

/**
 * A `useQuery` wrapper that guarantees consumers always receive fresh data —
 * never a stale cached value from a previous mount or a background refetch in progress.
 */

type FreshQueryOptions<
	TQueryFnData,
	TError,
	TData,
	TQueryKey extends QueryKey,
> = Omit<
	UseQueryOptions<TQueryFnData, TError, TData, TQueryKey>,
	"staleTime" | "gcTime" | "refetchOnMount" | "refetchOnReconnect"
>

export function useFreshQuery<
	TQueryFnData,
	TError = Error,
	TData = TQueryFnData,
	TQueryKey extends QueryKey = QueryKey,
>(options: FreshQueryOptions<TQueryFnData, TError, TData, TQueryKey>) {
	const query = useQuery({
		...options,
		staleTime: 0,
		gcTime: 0,
		refetchOnMount: "always",
		refetchOnReconnect: "always",
		refetchOnWindowFocus: options.refetchOnWindowFocus ?? false,
	})

	return {
		...query,
		data:
			query.isFetchedAfterMount && !query.isFetching ? query.data : undefined,
	}
}
