import { ConnectorOption } from "../../../types"
import { getConnectorImage } from "../../../utils/utils"

const connectorOptions: ConnectorOption[] = [
	{
		value: "MongoDB",
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage("MongoDB")}
					alt="MongoDB"
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-mongodb">MongoDB</span>
			</div>
		),
	},
	{
		value: "Postgres",
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage("Postgres")}
					alt="Postgres"
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-postgres">Postgres</span>
			</div>
		),
	},
	{
		value: "MySQL",
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage("MySQL")}
					alt="MySQL"
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-mysql">MySQL</span>
			</div>
		),
	},
	{
		value: "Oracle",
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage("Oracle")}
					alt="Oracle"
					className="mr-2 h-4 w-5"
				/>
				<span data-testid="connector-option-oracle">Oracle</span>
			</div>
		),
	},
]

export default connectorOptions
