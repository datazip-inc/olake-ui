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
				<span>MongoDB</span>
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
				<span>Postgres</span>
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
				<span>MySQL</span>
			</div>
		),
	},
]

export default connectorOptions
