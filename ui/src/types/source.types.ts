import { JobBasic } from "./job.types"

export interface Source {
	id: string
	name: string
	type: string
	status: "active" | "inactive" | "saved"
	createdAt: Date
	config?: any // Configuration data specific to the connector type
	associatedJobs?: JobBasic[]
}
