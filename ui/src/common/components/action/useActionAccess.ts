type ActionAccess = {
	canAccess: (access: string) => boolean
}

export const useActionAccess = (): ActionAccess => ({
	canAccess: () => true,
})
