export enum SourceConnector {
	MongoDB = "MongoDB",
	Postgres = "Postgres",
	MySQL = "MySQL",
	Oracle = "Oracle",
	Kafka = "Kafka",
	S3 = "S3",
}

export enum DestinationConnector {
	AmazonS3 = "Amazon S3",
	ApacheIceberg = "Apache Iceberg",
}

export enum CatalogType {
	Glue = "glue",
	JDBC = "jdbc",
	Hive = "hive",
	Rest = "rest",
}
