import { ConnectorOption } from "../../../types"
import AWSS3 from "../../../assets/AWSS3.svg"
import ApacheIceBerg from "../../../assets/ApacheIceBerg.svg"
import { CONNECTOR_TYPES } from "../../../utils/constants"

export const connectorOptions: ConnectorOption[] = [
	{
		value: CONNECTOR_TYPES.AMAZON_S3,
		label: (
			<div className="flex items-center">
				<img
					src={AWSS3}
					alt={CONNECTOR_TYPES.AMAZON_S3}
					className="mr-2 size-5"
				/>
				<span>{CONNECTOR_TYPES.AMAZON_S3}</span>
			</div>
		),
	},
	{
		value: CONNECTOR_TYPES.APACHE_ICEBERG,
		label: (
			<div className="flex items-center">
				<img
					src={ApacheIceBerg}
					alt={CONNECTOR_TYPES.APACHE_ICEBERG}
					className="mr-2 size-5"
				/>
				<span>{CONNECTOR_TYPES.APACHE_ICEBERG}</span>
			</div>
		),
	},
]
