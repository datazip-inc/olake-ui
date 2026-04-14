import { Modal } from "antd"

import { DestinationSuccess } from "@/assets"

const TestConnectionSuccessModal = ({
	open,
	connectionType = "source",
}: {
	open: boolean
	connectionType?: "source" | "destination" | "catalog"
}) => {
	const labelMap = {
		source: "Source",
		destination: "Destination",
		catalog: "Catalog",
	} as const
	const label = labelMap[connectionType]
	return (
		<Modal
			open={open}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-7 py-6">
				<img src={DestinationSuccess} />
				<div className="flex flex-col items-center">
					<p className="text-xs text-olake-text-tertiary">Successful</p>
					<h2 className="text-lg font-medium">
						{label} test connection is successful
					</h2>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionSuccessModal
