// Session storage utilities for managing browser session state

const RELEASES_FETCHED_SESSION_KEY = "releases_fetched_session"

// Checks if releases have been fetched in the current browser session
// returns true if releases were already fetched this session, false otherwise
export const hasFetchedReleasesThisSession = (): boolean => {
	return sessionStorage.getItem(RELEASES_FETCHED_SESSION_KEY) !== null
}

// Marks that releases have been fetched in the current browser session
export const markReleasesFetchedThisSession = (): void => {
	sessionStorage.setItem(RELEASES_FETCHED_SESSION_KEY, "true")
}

export const clearReleasesFetchedSession = (): void => {
	sessionStorage.removeItem(RELEASES_FETCHED_SESSION_KEY)
}
