import { Modal } from "antd"

import TestConnectionGif from "@/assets/TestConnectionGif.gif"

const TestConnectionModal = ({
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
			<div className="flex flex-col items-center justify-center gap-1 py-8">
				<img
					src={TestConnectionGif}
					className="max-w-[70%]"
				/>
				<div className="flex flex-col items-center">
					<p className="text-olake-text-tertiary">Please wait...</p>
					<div className="text-xl font-medium text-olake-text">
						Testing {label} connection
					</div>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionModal
