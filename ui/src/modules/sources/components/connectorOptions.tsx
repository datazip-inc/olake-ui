import { ConnectorOption } from "../../../types"
import { getConnectorImage } from "../../../utils/utils"
import { CONNECTOR_TYPES } from "../../../utils/constants"

export const connectorOptions: ConnectorOption[] = [
	{
		value: CONNECTOR_TYPES.MONGODB,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.MONGODB)}
					alt={CONNECTOR_TYPES.MONGODB}
					className="mr-2 size-5"
				/>
				<span>{CONNECTOR_TYPES.MONGODB}</span>
			</div>
		),
	},
	{
		value: CONNECTOR_TYPES.POSTGRES,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.POSTGRES)}
					alt={CONNECTOR_TYPES.POSTGRES}
					className="mr-2 size-5"
				/>
				<span>{CONNECTOR_TYPES.POSTGRES}</span>
			</div>
		),
	},
	{
		value: CONNECTOR_TYPES.MYSQL,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.MYSQL)}
					alt={CONNECTOR_TYPES.MYSQL}
					className="mr-2 size-5"
				/>
				<span>{CONNECTOR_TYPES.MYSQL}</span>
			</div>
		),
	},
	{
		value: CONNECTOR_TYPES.ORACLE,
		label: (
			<div className="flex items-center">
				<img
					src={getConnectorImage(CONNECTOR_TYPES.ORACLE)}
					alt={CONNECTOR_TYPES.ORACLE}
					className="mr-2 h-4 w-5"
				/>
				<span>{CONNECTOR_TYPES.ORACLE}</span>
			</div>
		),
	},
]

export default connectorOptions
