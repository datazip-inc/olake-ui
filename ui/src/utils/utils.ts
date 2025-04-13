// import mongodb and postgres and mysql
import MongoDB from "../assets/Mongo.svg"
import Postgres from "../assets/Postgres.svg"
import MySQL from "../assets/MySQL.svg"
import AWSS3 from "../assets/AWSS3.svg"
import ApacheIceberg from "../assets/ApacheIceBerg.svg"

export const getConnectorImage = (connector: string) => {
	const lowerConnector = connector.toLowerCase()

	if (lowerConnector.includes("mongo")) {
		return MongoDB
	} else if (lowerConnector.includes("postgres")) {
		return Postgres
	} else if (lowerConnector.includes("mysql")) {
		return MySQL
	} else if (
		lowerConnector.includes("s3") ||
		lowerConnector.includes("amazon")
	) {
		return AWSS3
	} else if (
		lowerConnector.includes("iceberg") ||
		lowerConnector.includes("apache") ||
		lowerConnector.includes("glue") ||
		lowerConnector.includes("jdbc")
	) {
		return ApacheIceberg
	}

	// Default placeholder
	return MongoDB
}
