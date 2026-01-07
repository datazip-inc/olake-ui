import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useAppStore } from "../../../store"
import { Job } from "../../../types"

interface UseJobDetailsOptions {
	jobId: string | undefined
	onJobFetched?: (job: Job) => void
}

/** Custom hook to fetch job details by ID and calling optional callback */
export const useJobDetails = ({
	jobId,
	onJobFetched,
}: UseJobDetailsOptions) => {
	const navigate = useNavigate()
	const { fetchSelectedJob, selectedJob, setSelectedJobId } = useAppStore()

	useEffect(() => {
		const fetchJob = async () => {
			if (!jobId) {
				navigate("/jobs")
				return
			}

			try {
				const job = await fetchSelectedJob(jobId)
				if (job) {
					setSelectedJobId(job.id.toString())

					// Call the optional callback if provided
					onJobFetched?.(job)
				}
			} catch (error) {
				console.error("Error fetching job:", error)
				navigate("/jobs")
			}
		}

		fetchJob()
	}, [jobId])

	return selectedJob
}
