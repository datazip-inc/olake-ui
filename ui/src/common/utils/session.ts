/**
 * Small wrapper around sessionStorage with JSON (de)serialization
 */

const get = <T>(key: string): T | null => {
	try {
		const value = sessionStorage.getItem(key)
		return value ? (JSON.parse(value) as T) : null
	} catch (error) {
		console.error(`sessionStore: failed to read "${key}"`, error)
		return null
	}
}

const set = <T>(key: string, value: T): void => {
	try {
		sessionStorage.setItem(key, JSON.stringify(value))
	} catch (error) {
		console.error(`sessionStore: failed to write "${key}"`, error)
	}
}

export const sessionStore = { get, set }
