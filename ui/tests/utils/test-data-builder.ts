/**
 * Test Data Builder Utility
 *
 * Provides methods to generate unique, consistent test data
 * for sources, destinations, and jobs to avoid conflicts
 * and ensure test isolation.
 */

export class TestDataBuilder {
	static uniqueName(prefix: string): string {
		const timestamp = Date.now()
		return `${prefix}_${timestamp}`
	}

	static getUniqueSourceName(connector: string = "postgres"): string {
		return this.uniqueName(`e2e_${connector}_source`)
	}

	static getUniqueDestinationName(connector: string = "iceberg"): string {
		return this.uniqueName(`e2e_${connector}_dest`)
	}

	static getUniqueJobName(
		sourceConnector: string,
		destinationConnector: string,
		catalogType?: string,
	): string {
		return catalogType
			? `${sourceConnector}_${destinationConnector}_${catalogType}_job`
			: `${sourceConnector}_${destinationConnector}_job`
	}

	static timestamp(): number {
		return Date.now()
	}
}
