import { SavedJobDraft } from "../types"

const SAVED_JOBS_KEY = "savedJobs"

export const savedJobsService = {
	getAll(): SavedJobDraft[] {
		try {
			return JSON.parse(
				localStorage.getItem(SAVED_JOBS_KEY) ?? "[]",
			) as SavedJobDraft[]
		} catch {
			return []
		}
	},

	upsert(draft: SavedJobDraft): void {
		try {
			const jobs = this.getAll()
			const idx = jobs.findIndex(j => j.id === draft.id)
			if (idx !== -1) {
				jobs[idx] = draft
			} else {
				jobs.push(draft)
			}
			localStorage.setItem(SAVED_JOBS_KEY, JSON.stringify(jobs))
		} catch {
			// storage unavailable or quota exceeded
		}
	},

	remove(id: string): SavedJobDraft[] {
		try {
			const updated = this.getAll().filter(j => j.id !== id)
			localStorage.setItem(SAVED_JOBS_KEY, JSON.stringify(updated))
			return updated
		} catch {
			return this.getAll()
		}
	},
}
