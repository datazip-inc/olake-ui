export const getConnectorLabel = (type: string): string => {
	switch (type) {
		case "mongodb":
		case "MongoDB":
			return "MongoDB"
		case "postgres":
		case "Postgres":
			return "Postgres"
		case "mysql":
		case "MySQL":
			return "MySQL"
		case "oracle":
		case "Oracle":
			return "Oracle"
		case "kafka":
			return "Kafka"
		case "s3":
		case "S3":
			return "S3"
		case "db2":
			return "DB2"
		case "mssql":
			return "MSSQL"
		default:
			return "Default"
	}
}
