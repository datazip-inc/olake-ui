const REST_CATALOG_SLUGS: Record<string, string> = {
	rest: "generic",
	lakekeeper: "lakekeeper",
	nessie: "nessie",
	s3tables: "s3-tables",
	unity: "unity",
	polaris: "polaris",
	biglake: "biglake",
}

const AUTH_TYPE_SLUGS: Record<string, Record<string, string>> = {
	lakekeeper: { None: "none", OAuth2: "oauth2", Token: "token" },
	nessie: { None: "none", OAuth2: "oauth2", Token: "token" },
	polaris: { OAuth2: "oauth2", Token: "token" },
	unity: {
		"Personal Access Token (PAT)": "personal-access-token",
		"OAuth2 M2M": "oauth2-m2m",
		"OAuth2 U2M": "oauth2-u2m",
		"Token Federation": "token-federation",
	},
}

export const getConnectorDocumentationPath = (
	connector: string,
	catalog: string | null,
	authType?: string | null,
) => {
	switch (connector) {
		case "Amazon S3":
			return "s3/config"
		case "Apache Iceberg":
			switch (catalog) {
				case "jdbc":
					return "iceberg/catalog/jdbc"
				case "hive":
					return "iceberg/catalog/hive"
				case "glue":
					return "iceberg/catalog/glue"
				default: {
					if (!catalog) return "iceberg/catalog/glue"

					const catalogSlug = REST_CATALOG_SLUGS[catalog]
					if (!catalogSlug) return "iceberg/catalog/glue"

					const params = new URLSearchParams({ "rest-catalog": catalogSlug })
					const authSlug = authType
						? AUTH_TYPE_SLUGS[catalog]?.[authType]
						: undefined
					if (authSlug) params.set("rest-auth-type", authSlug)

					return `iceberg/catalog/rest/?${params.toString()}`
				}
			}
		default:
			return undefined
	}
}
