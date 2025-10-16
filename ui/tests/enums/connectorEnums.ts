export enum SourceConnector {
	MongoDB = "MongoDB",
	Postgres = "Postgres",
	MySQL = "MySQL",
	Oracle = "Oracle",
}

export enum DestinationConnector {
	AmazonS3 = "Amazon S3",
	ApacheIceberg = "Apache Iceberg",
}

export enum ConnectorType {
	SOURCE = "source",
	DESTINATION = "destination",
}

export enum CatalogType {
	Glue = "glue",
	JDBC = "jdbc",
	Hive = "hive",
	Rest = "rest",
}
