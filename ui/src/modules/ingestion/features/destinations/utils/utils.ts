export const getConnectorDocumentationPath = (
	connector: string,
	catalog: string | null,
) => {
	switch (connector) {
		case "Amazon S3":
			return "s3/config"
		case "Apache Iceberg":
			switch (catalog) {
				case "glue":
					return "iceberg/catalog/glue"
				case "rest":
					return "iceberg/catalog/rest"
				case "jdbc":
					return "iceberg/catalog/jdbc"
				case "hive":
					return "iceberg/catalog/hive"
				default:
					return "iceberg/catalog/glue"
			}
		default:
			return undefined
	}
}
