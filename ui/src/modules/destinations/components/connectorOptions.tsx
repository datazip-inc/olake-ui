import { ConnectorOption } from "../../../types"
import AWSS3 from "../../../assets/AWSS3.svg"
import ApacheIceBerg from "../../../assets/ApacheIceBerg.svg"
import { CONNECTOR_TYPES } from "../../../utils/constants"
import { DestinationConnector } from "../../../enums"

export const connectorOptions: ConnectorOption[] = [
	{
		value: DestinationConnector.AMAZON_S3,
		label: (
			<div className="flex items-center">
				<img
					src={AWSS3}
					alt={CONNECTOR_TYPES.AMAZON_S3}
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-s3">Amazon S3</span>
			</div>
		),
	},
	{
		value: DestinationConnector.APACHE_ICEBERG,
		label: (
			<div className="flex items-center">
				<img
					src={ApacheIceBerg}
					alt={CONNECTOR_TYPES.APACHE_ICEBERG}
					className="mr-2 size-5"
				/>
				<span data-testid="connector-option-iceberg">Apache Iceberg</span>
			</div>
		),
	},
]
