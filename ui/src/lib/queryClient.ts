import { MutationCache, QueryClient } from "@tanstack/react-query"

export const queryClient = new QueryClient({
	mutationCache: new MutationCache({
		onSuccess: (_data, _variables, _context, mutation) => {
			queryClient.invalidateQueries({
				queryKey: mutation.options.mutationKey,
			})
		},
	}),
})
