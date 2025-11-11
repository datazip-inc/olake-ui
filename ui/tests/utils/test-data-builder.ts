/**
 * Test Data Builder Utility
 *
 * Provides methods to generate unique, consistent test data
 * for sources, destinations, and jobs to avoid conflicts
 * and ensure test isolation.
 */

import { CatalogType, DestinationConnector, SourceConnector } from "../enums"

export class TestDataBuilder {
	static readonly SOURCE_CONNECTOR_LABELS: Record<SourceConnector, string> = {
		[SourceConnector.Postgres]: "postgres",
		[SourceConnector.MongoDB]: "mongodb",
		[SourceConnector.MySQL]: "mysql",
		[SourceConnector.Oracle]: "oracle",
		[SourceConnector.Kafka]: "kafka",
	}

	static readonly DESTINATION_CONNECTOR_LABELS: Record<
		DestinationConnector,
		string
	> = {
		[DestinationConnector.ApacheIceberg]: "iceberg",
		[DestinationConnector.AmazonS3]: "s3",
	}

	static readonly CATALOG_TYPE_LABELS: Record<CatalogType, string> = {
		[CatalogType.Glue]: "glue",
		[CatalogType.JDBC]: "jdbc",
		[CatalogType.Hive]: "hive",
		[CatalogType.Rest]: "rest",
	}

	static uniqueName(prefix: string): string {
		const timestamp = Date.now()
		return `${prefix}_${timestamp}`
	}

	static getUniqueSourceName(connector: SourceConnector): string {
		return this.uniqueName(
			`e2e_${this.SOURCE_CONNECTOR_LABELS[connector]}_source`,
		)
	}

	static getUniqueDestinationName(
		connector: DestinationConnector,
		catalogType?: CatalogType,
	): string {
		return catalogType
			? this.uniqueName(
					`e2e_${this.DESTINATION_CONNECTOR_LABELS[connector]}_${this.CATALOG_TYPE_LABELS[catalogType]}_dest`,
				)
			: this.uniqueName(
					`e2e_${this.DESTINATION_CONNECTOR_LABELS[connector]}_dest`,
				)
	}

	static getUniqueJobName(
		sourceConnector: SourceConnector,
		destinationConnector: DestinationConnector,
		catalogType?: CatalogType,
	): string {
		return catalogType
			? `${this.SOURCE_CONNECTOR_LABELS[sourceConnector]}_${this.DESTINATION_CONNECTOR_LABELS[destinationConnector]}_${this.CATALOG_TYPE_LABELS[catalogType]}_job`
			: `${this.SOURCE_CONNECTOR_LABELS[sourceConnector]}_${this.DESTINATION_CONNECTOR_LABELS[destinationConnector]}_job`
	}

	static timestamp(): number {
		return Date.now()
	}
}
