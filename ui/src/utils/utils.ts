import MongoDB from "../assets/Mongo.svg"
import Postgres from "../assets/Postgres.svg"
import MySQL from "../assets/MySQL.svg"
import AWSS3 from "../assets/AWSS3.svg"
import ApacheIceBerg from "../assets/ApacheIceBerg.svg"

export const getConnectorImage = (connector: string) => {
	const lowerConnector = connector.toLowerCase()

	if (lowerConnector === "mongodb") {
		return MongoDB
	} else if (lowerConnector === "postgres") {
		return Postgres
	} else if (lowerConnector === "mysql") {
		return MySQL
	} else if (lowerConnector === "s3" || lowerConnector === "amazon") {
		return AWSS3
	} else if (
		lowerConnector === "iceberg" ||
		lowerConnector === "apache iceberg"
	) {
		return ApacheIceBerg
	}

	// Default placeholder
	return MongoDB
}

export const getConnectorName = (connector: string, catalog: string | null) => {
	if (connector === "Amazon S3") {
		return "s3/config"
	} else if (connector === "Apache Iceberg") {
		if (catalog === "AWS Glue") {
			return "iceberg/catalog/glue"
		} else if (catalog === "REST Catalog") {
			return "iceberg/catalog/rest"
		} else if (catalog === "JDBC Catalog") {
			return "iceberg/catalog/jdbc"
		} else if (catalog === "Hive Catalog" || catalog === "HIVE Catalog") {
			return "iceberg/catalog/hive"
		}
	}
}

export const getStatusClass = (status: string) => {
	switch (status.toLowerCase()) {
		case "success":
		case "completed":
			return "text-[#52C41A] bg-[#F6FFED]"
		case "failed":
		case "cancelled":
			return "text-[#F5222D] bg-[#FFF1F0]"
		case "running":
			return "text-[#0958D9] bg-[#E6F4FF]"
		case "scheduled":
			return "text-[rgba(0,0,0,88)] bg-[#f0f0f0]"
		default:
			return "text-[rgba(0,0,0,88)] bg-[#f0f0f0]"
	}
}

export const getConnectorInLowerCase = (connector: string) => {
	if (connector === "Amazon S3") {
		return "s3"
	} else if (connector === "Apache Iceberg") {
		return "iceberg"
	} else if (connector === "MongoDB") {
		return "mongodb"
	} else if (connector === "Postgres") {
		return "postgres"
	} else if (connector === "MySQL") {
		return "mysql"
	} else {
		return connector.toLowerCase()
	}
}

export const getCatalogInLowerCase = (catalog: string) => {
	if (catalog === "AWS Glue" || catalog === "glue") {
		return "glue"
	} else if (catalog === "REST Catalog" || catalog === "rest") {
		return "rest"
	} else if (catalog === "JDBC Catalog" || catalog === "jdbc") {
		return "jdbc"
	} else if (catalog === "Hive Catalog" || catalog === "hive") {
		return "hive"
	}
}

export const getStatusLabel = (status: string) => {
	switch (status) {
		case "success":
			return "Success"
		case "failed":
			return "Failed"
		case "cancelled":
			return "Cancelled"
		case "running":
			return "Running"
		case "scheduled":
			return "Scheduled"
		default:
			return status
	}
}

export const getConnectorLabel = (type: string): string => {
	switch (type) {
		case "mongodb":
			return "MongoDB"
		case "postgres":
			return "Postgres"
		case "mysql":
			return "MySQL"
		default:
			return "MongoDB"
	}
}

export const getFrequencyValue = (frequency: string) => {
	if (frequency.includes(" ")) {
		const parts = frequency.split(" ")
		const unit = parts[1].toLowerCase()

		if (unit.includes("hour")) return "hours"
		if (unit.includes("minute")) return "minutes"
		if (unit.includes("day")) return "days"
		if (unit.includes("week")) return "weeks"
		if (unit.includes("month")) return "months"
		if (unit.includes("year")) return "years"
	}

	switch (frequency) {
		case "hourly":
		case "hours":
			return "hours"
		case "daily":
		case "days":
			return "days"
		case "weekly":
		case "weeks":
			return "weeks"
		case "monthly":
		case "months":
			return "months"
		case "yearly":
		case "years":
			return "years"
		case "minutes":
			return "minutes"
		default:
			return "hours"
	}
}

export const removeSavedJobFromLocalStorage = (jobId: string) => {
	const savedJobs = localStorage.getItem("savedJobs")
	if (savedJobs) {
		const jobs = JSON.parse(savedJobs)
		const filteredJobs = jobs.filter((job: any) => job.id !== jobId)
		localStorage.setItem("savedJobs", JSON.stringify(filteredJobs))
	}
}

export const getReplicationFrequency = (replicationFrequency: string) => {
	if (replicationFrequency.includes(" ")) {
		const parts = replicationFrequency.split(" ")
		const value = parts[0]
		const unit = parts[1].toLowerCase()

		if (unit.includes("minute")) return `${value} minutes`
		if (unit.includes("hour")) return "hourly"
		if (unit.includes("day")) return "daily"
		if (unit.includes("week")) return "weekly"
		if (unit.includes("month")) return "monthly"
		if (unit.includes("year")) return "yearly"
	}

	if (replicationFrequency === "minutes") {
		return "minutes"
	} else if (replicationFrequency === "hours") {
		return "hourly"
	} else if (replicationFrequency === "days") {
		return "daily"
	} else if (replicationFrequency === "weeks") {
		return "weekly"
	} else if (replicationFrequency === "months") {
		return "monthly"
	} else if (replicationFrequency === "years") {
		return "yearly"
	}
}

export const getLogLevelClass = (level: string) => {
	switch (level) {
		case "debug":
			return "text-blue-600 bg-[#F0F5FF]"
		case "info":
			return "text-[#531DAB] bg-[#F9F0FF]"
		case "warning":
		case "warn":
			return "text-[#FAAD14] bg-[#FFFBE6]"
		case "error":
		case "fatal":
			return "text-red-500 bg-[#FFF1F0]"
		default:
			return "text-gray-600"
	}
}

export const getLogTextColor = (level: string) => {
	switch (level) {
		case "warning":
		case "warn":
			return "text-[#FAAD14]"
		case "error":
		case "fatal":
			return "text-[#F5222D]"
		default:
			return "text-[#000000"
	}
}
