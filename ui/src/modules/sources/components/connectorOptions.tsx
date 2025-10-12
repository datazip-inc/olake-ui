import { ConnectorOption } from "../../../types"
import { getConnectorImage } from "../../../utils/utils"
import { CONNECTOR_TYPES } from "../../../utils/constants"
import { SourceConnector } from "../../../enums"

export const connectorOptions: ConnectorOption[] = [
	{
		value: SourceConnector.MONGODB,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.MONGODB)}
					alt={CONNECTOR_TYPES.MONGODB}
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-mongodb">MongoDB</span>
			</div>
		),
	},
	{
		value: SourceConnector.POSTGRES,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.POSTGRES)}
					alt={CONNECTOR_TYPES.POSTGRES}
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-postgres">Postgres</span>
			</div>
		),
	},
	{
		value: SourceConnector.MYSQL,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.MYSQL)}
					alt={CONNECTOR_TYPES.MYSQL}
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-mysql">MySQL</span>
			</div>
		),
	},
	{
		value: SourceConnector.ORACLE,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.ORACLE)}
					alt={CONNECTOR_TYPES.ORACLE}
					className="mr-2 h-4 w-5"
				/>
				<span data-testid="connector-option-oracle">Oracle</span>
			</div>
		),
	},
]

export default connectorOptions
