import { ConnectorOption } from "../../../types"
import AWSS3 from "../../../assets/AWSS3.svg"
import ApacheIceBerg from "../../../assets/ApacheIceBerg.svg"

export const connectorOptions: ConnectorOption[] = [
	{
		value: "Amazon S3",
		label: (
			<div className="flex items-center">
				<img
					src={AWSS3}
					alt="AWS S3"
					className="mr-2 size-5"
				/>
				<span>Amazon S3</span>
			</div>
		),
	},
	{
		value: "Apache Iceberg",
		label: (
			<div className="flex items-center">
				<img
					src={ApacheIceBerg}
					alt="Apache Iceberg"
					className="mr-2 size-5"
				/>
				<span>Apache Iceberg</span>
			</div>
		),
	},
]
