import api from "../axios"
import {
	APIResponse,
	Entity,
	EntityBase,
	EntityTestResponse,
} from "../../types"

// Flag to use mock data instead of real API
const useMockData = true

const mockDestinationConnectorSchemas = {
	"Amazon S3": {
		type: "object",
		properties: {
			type: {
				type: "string",
				title: "Output Format",
				description: "Specifies the output file format for writing data",
				enum: ["PARQUET"],
				default: "PARQUET",
			},
			normalization: {
				type: "boolean",
				title: "Enable Normalization",
				description:
					"Indicates whether data normalization (JSON flattening) should be applied before writing data to S3",
				default: false,
			},
			s3_bucket: {
				type: "string",
				title: "S3 Bucket",
				description:
					"The name of the Amazon S3 bucket where your output files will be stored",
			},
			s3_region: {
				type: "string",
				title: "S3 Region",
				description: "The AWS region where the specified S3 bucket is hosted",
				enum: [
					"ap-south-1",
					"us-east-1",
					"us-east-2",
					"us-west-1",
					"us-west-2",
					"ap-southeast-1",
					"ap-southeast-2",
					"ap-northeast-1",
					"eu-central-1",
					"eu-west-1",
				],
				default: "ap-south-1",
			},
			s3_access_key: {
				type: "string",
				title: "AWS Access Key",
				description: "The AWS access key used for authenticating S3 requests",
			},
			s3_secret_key: {
				type: "string",
				title: "AWS Secret Key",
				description: "The AWS secret key used for S3 authentication",
				format: "password",
			},
			s3_path: {
				type: "string",
				title: "S3 Path",
				description:
					"The specific path (or prefix) within the S3 bucket where data files will be written",
				default: "/data",
			},
		},
		required: [
			"type",
			"s3_bucket",
			"s3_region",
			"s3_access_key",
			"s3_secret_key",
			"s3_path",
		],
		uiSchema: {
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			type: {
				"ui:widget": "select",
			},
		},
	},
	"AWS Glue": {
		type: "object",
		properties: {
			type: {
				type: "string",
				title: "Output Format",
				description: "Specifies the output file format",
				enum: ["ICEBERG"],
				default: "ICEBERG",
			},
			catalog_type: {
				type: "string",
				title: "Catalog Type",
				description: "Type of catalog to use",
				enum: ["glue"],
				default: "glue",
			},
			normalization: {
				type: "boolean",
				title: "Enable Normalization",
				description: "Flag to enable or disable data normalization",
				default: false,
			},
			iceberg_s3_path: {
				type: "string",
				title: "Iceberg S3 Path",
				description: "S3 path where the Iceberg data is stored in AWS",
				pattern: "^s3://[^/]+/.+$",
				examples: ["s3://bucket_name/olake_iceberg/test_olake"],
			},
			aws_region: {
				type: "string",
				title: "AWS Region",
				description:
					"AWS region where the S3 bucket and Glue catalog are located",
				enum: [
					"ap-south-1",
					"us-east-1",
					"us-east-2",
					"us-west-1",
					"us-west-2",
					"ap-southeast-1",
					"ap-southeast-2",
					"ap-northeast-1",
					"eu-central-1",
					"eu-west-1",
				],
				default: "ap-south-1",
			},
			aws_access_key: {
				type: "string",
				title: "AWS Access Key",
				description:
					"AWS access key with sufficient permissions for S3 and Glue",
			},
			aws_secret_key: {
				type: "string",
				title: "AWS Secret Key",
				description: "AWS secret key corresponding to the access key",
				format: "password",
			},
			iceberg_db: {
				type: "string",
				title: "Iceberg Database",
				description: "Name of the database to be created in AWS Glue",
			},
			grpc_port: {
				type: "integer",
				title: "grpc Port",
				description: "Port on which the grpc server listens",
				default: 50051,
				minimum: 1,
				maximum: 65535,
			},
			server_host: {
				type: "string",
				title: "Server Host",
				description: "Host address of the grpc server",
				default: "localhost",
			},
		},
		required: [
			"type",
			"catalog_type",
			"iceberg_s3_path",
			"aws_region",
			"aws_access_key",
			"aws_secret_key",
			"iceberg_db",
			"grpc_port",
			"server_host",
		],
		uiSchema: {
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			type: {
				"ui:widget": "select",
			},
		},
	},
	"REST Catalog": {
		type: "object",
		properties: {
			type: {
				type: "string",
				title: "Output Format",
				description: "Specifies the output file format",
				enum: ["ICEBERG"],
				default: "ICEBERG",
			},
			catalog_type: {
				type: "string",
				title: "Catalog Type",
				description: "Type of catalog to use",
				enum: ["rest"],
				default: "rest",
			},
			normalization: {
				type: "boolean",
				title: "Enable Normalization",
				description: "Indicates whether data normalization is applied",
				default: false,
			},
			rest_catalog_url: {
				type: "string",
				title: "REST Catalog URL",
				description: "Endpoint URL for the REST catalog service",
				format: "uri",
				examples: ["http://localhost:8181/catalog"],
			},
			iceberg_s3_path: {
				type: "string",
				title: "Iceberg S3 Path",
				description: "S3 path or storage location for Iceberg data",
				default: "warehouse",
			},
			iceberg_db: {
				type: "string",
				title: "Iceberg Database",
				description: "Name of the Iceberg database to be used",
				default: "olake_iceberg",
			},
		},
		required: [
			"type",
			"catalog_type",
			"rest_catalog_url",
			"iceberg_s3_path",
			"iceberg_db",
		],
		uiSchema: {
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			type: {
				"ui:widget": "select",
			},
		},
	},
	"JDBC Catalog": {
		type: "object",
		properties: {
			type: {
				type: "string",
				title: "Output Format",
				description: "Specifies the output file format",
				enum: ["ICEBERG"],
				default: "ICEBERG",
			},
			catalog_type: {
				type: "string",
				title: "Catalog Type",
				description: "Type of catalog to use",
				enum: ["jdbc"],
				default: "jdbc",
			},
			jdbc_url: {
				type: "string",
				title: "JDBC URL",
				description: "JDBC connection URL for the catalog database",
				format: "uri",
				examples: ["jdbc:postgresql://localhost:5432/iceberg"],
			},
			jdbc_username: {
				type: "string",
				title: "JDBC Username",
				description: "Username for JDBC connection",
			},
			jdbc_password: {
				type: "string",
				title: "JDBC Password",
				description: "Password for JDBC connection",
				format: "password",
			},
			normalization: {
				type: "boolean",
				title: "Enable Normalization",
				description: "Flag to enable or disable data normalization",
				default: false,
			},
			iceberg_s3_path: {
				type: "string",
				title: "Iceberg S3 Path",
				description: "S3 path where Iceberg data is stored",
				pattern: "^s3a?://[^/]+/.+$",
				examples: ["s3a://warehouse"],
			},
			s3_endpoint: {
				type: "string",
				title: "S3 Endpoint",
				description: "S3-compatible storage endpoint URL",
				// format: "uri",
				examples: ["http://localhost:9000"],
			},
			s3_use_ssl: {
				type: "boolean",
				title: "Use SSL",
				description: "Whether to use SSL for S3 connections",
				default: false,
			},
			s3_path_style: {
				type: "boolean",
				title: "Path Style Access",
				description: "Whether to use path-style access for S3",
				default: true,
			},
			aws_access_key: {
				type: "string",
				title: "AWS Access Key",
				description: "AWS access key for S3 access",
			},
			aws_region: {
				type: "string",
				title: "AWS Region",
				description: "AWS region for S3 access",
				enum: [
					"ap-south-1",
					"us-east-1",
					"us-east-2",
					"us-west-1",
					"us-west-2",
					"ap-southeast-1",
					"ap-southeast-2",
					"ap-northeast-1",
					"eu-central-1",
					"eu-west-1",
				],
				default: "ap-south-1",
			},
			aws_secret_key: {
				type: "string",
				title: "AWS Secret Key",
				description: "AWS secret key for S3 access",
				format: "password",
			},
			iceberg_db: {
				type: "string",
				title: "Iceberg Database",
				description: "Name of the Iceberg database",
				default: "olake_iceberg",
			},
		},
		required: [
			"type",
			"catalog_type",
			"jdbc_url",
			"jdbc_username",
			"jdbc_password",
			"iceberg_s3_path",
			"s3_endpoint",
			"aws_access_key",
			"aws_secret_key",
			"iceberg_db",
		],
		uiSchema: {
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			type: {
				"ui:widget": "select",
			},
		},
	},
}

export const destinationService = {
	getDestinations: async () => {
		try {
			const response = await api.get<APIResponse<Entity[]>>(
				"/api/v1/project/123/destinations",
			)

			const destinations: Entity[] = response.data.data.map(item => {
				const config = JSON.parse(item.config)

				return {
					...item,
					config,
					status: "active",
				}
			})

			return destinations
		} catch (error) {
			console.error("Error fetching sources from API:", error)
			throw error
		}
	},

	// Create new destination
	createDestination: async (
		destination: Omit<Destination, "id" | "createdAt">,
	) => {
		const response = await api.post<EntityBase>(
			"/api/v1/project/123/destinations",
			destination,
		)
		return response.data
	},

	// Update destination
	updateDestination: async (id: string, destination: any) => {
		try {
			const response = await api.put<APIResponse<any>>(
				`/api/v1/project/123/destinations/${id}`,
				{
					name: destination.name,
					type: destination.type,
					version: destination.version,
					config: destination.config,
				},
			)
			return response.data
		} catch (error) {
			console.error("Error updating destination:", error)
			throw error
		}
	},

	// Delete destination
	deleteDestination: async (id: number) => {
		await api.delete(`/api/v1/project/123/destinations/${id}`)
		return
	},

	getConnectorSchema: async (connectorType: string) => {
		try {
			if (useMockData) {
				// Return mock schema for development
				const schema =
					mockDestinationConnectorSchemas[
						connectorType as keyof typeof mockDestinationConnectorSchemas
					]
				// If schema doesn't exist, return empty object
				if (!schema) {
					return {}
				}
				return { ...schema }
			}

			// For production: fetch schema from API
			const response = await api.get(`/connectors/${connectorType}/schema`)
			return response.data
		} catch (error) {
			console.error(`Error fetching schema for ${connectorType}:`, error)
			throw error
		}
	},

	// Test destination connection
	testDestinationConnection: async (destination: EntityTestResponse) => {
		try {
			const response = await api.post<APIResponse<EntityTestResponse>>(
				"/api/v1/project/123/destinations/test",
				{
					type: destination.type.toLowerCase(),
					version: "1.0.0",
					config: destination.config,
				},
			)
			return {
				success: response.data.success,
				message: response.data.message,
			}
		} catch (error) {
			console.error("Error testing destination connection:", error)
			return {
				success: false,
				message:
					error instanceof Error ? error.message : "Unknown error occurred",
			}
		}
	},

	getDestinationVersions: async (type: string) => {
		const response = await api.get<APIResponse<{ version: string[] }>>(
			`/api/v1/project/123/destinations/versions/?type=${type}`,
		)
		return response.data
	},

	getDestinationSpec: async (type: string, catalog: string | null) => {
		if (!catalog) {
			catalog = "none"
		}
		if (type.toLowerCase() === "amazon s3") {
			type = "s3"
		} else if (type.toLowerCase() === "apache iceberg") {
			type = "iceberg"
		}
		if (catalog.toLowerCase() === "aws glue") {
			catalog = "glue"
		} else if (catalog.toLowerCase() === "rest catalog") {
			catalog = "rest"
		} else if (catalog.toLowerCase() === "jdbc catalog") {
			catalog = "jdbc"
		} else if (catalog.toLowerCase() === "hive catalog") {
			catalog = "hive"
		}

		const response = await api.post<APIResponse<any>>(
			`/api/v1/project/123/destinations/spec`,
			{
				type: type.toLowerCase(),
				version: "latest",
				catalog: catalog,
			},
		)
		return response.data
	},
}
