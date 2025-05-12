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
		} else if (catalog === "Hive Catalog") {
			return "iceberg/catalog/hive"
		}
	}
}

export const getStatusClass = (status: string) => {
	switch (status) {
		case "success":
			return "text-[#52C41A] bg-[#F6FFED]"
		case "failed":
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
	if (catalog === "AWS Glue") {
		return "glue"
	} else if (catalog === "REST Catalog") {
		return "rest"
	} else if (catalog === "JDBC Catalog") {
		return "jdbc"
	} else if (catalog === "Hive Catalog") {
		return "hive"
	}
}

export const getStatusLabel = (status: string) => {
	switch (status) {
		case "success":
			return "Success"
		case "failed":
			return "Failed"
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
		default:
			return "MySQL"
	}
}
