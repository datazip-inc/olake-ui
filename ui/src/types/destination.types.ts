import { JobBasic } from "./job.types"

export interface Destination {
	id: string
	name: string
	type: string
	catalog?: string
	status: "active" | "inactive" | "saved"
	createdAt: Date
	associatedJobs?: JobBasic[]
}
