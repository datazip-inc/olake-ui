export const useActionAccess = () => ({
	canAccess: (feature?: string) => {
		void feature
		return true
	},
})
