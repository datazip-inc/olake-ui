import { InfoIcon, WarningIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import { useState } from "react"

import { SelectedStream } from "@/modules/ingestion/common/types"
import {
	selectStreamsData,
	useJobStore,
	useStreamSelectionStore,
} from "@/modules/ingestion/features/jobs/stores"
import { StreamDifferenceModalProps } from "@/modules/ingestion/features/jobs/types"
import { coversAllSelectedStreams } from "@/modules/ingestion/features/jobs/utils"

const StreamDifferenceModal = ({
	streamDifference,
	showIndexWarning = false,
	onConfirm,
}: StreamDifferenceModalProps) => {
	const [isLoading, setIsLoading] = useState(false)

	const { showStreamDifferenceModal, setShowStreamDifferenceModal } =
		useJobStore()
	const streamsData = useStreamSelectionStore(selectStreamsData)
	// Nothing is spared, so listing the streams adds no information.
	const allStreamsImpacted = coversAllSelectedStreams(
		streamsData,
		streamDifference,
	)

	const handleCloseModal = () => {
		setShowStreamDifferenceModal(false)
	}

	const handleFinish = async () => {
		setIsLoading(true)
		await onConfirm()
		setIsLoading(false)

		setShowStreamDifferenceModal(false)
	}

	const renderStreamsByNamespace = () => {
		const namespaces = Object.keys(streamDifference.selected_streams)

		return namespaces.map(namespace => {
			const streams = streamDifference.selected_streams[namespace]

			if (!streams || streams.length === 0) return null

			return (
				<ul
					key={namespace}
					className="mb-4 list-disc"
				>
					<li className="mb-2 text-sm font-semibold text-gray-700">
						{namespace}
						<ul className="mt-1 list-disc space-y-1 pl-6">
							{streams.map((stream: SelectedStream, index: number) => (
								<li
									key={`${namespace}-${stream.stream_name}-${index}`}
									className="text-sm font-normal text-gray-600"
								>
									{stream.stream_name}
								</li>
							))}
						</ul>
					</li>
				</ul>
			)
		})
	}

	return (
		<>
			<Modal
				title={
					<div className="flex justify-center">
						<WarningIcon
							weight="fill"
							className="size-12 text-primary"
						/>
					</div>
				}
				open={showStreamDifferenceModal}
				onCancel={handleCloseModal}
				footer={[
					<Button
						key="cancel"
						onClick={handleCloseModal}
					>
						Cancel
					</Button>,
					<Button
						key="edit"
						type="primary"
						onClick={handleFinish}
						loading={isLoading}
						className="bg-primary hover:bg-primary-600"
					>
						Confirm and Finish
					</Button>,
				]}
				centered
				width="30%"
			>
				<div className="mt-2 text-center">
					<h3 className="text-xl font-medium">
						Are you sure you want to continue?
					</h3>
					<p className="mt-4 text-left text-sm text-black">
						{allStreamsImpacted
							? "Modifying stream configurations will clear destination data for all of the impacted streams."
							: "Modifying stream configurations will clear destination data for the impacted streams. Following streams will be impacted:"}
					</p>
					<div className="mt-3 flex w-full items-center justify-center gap-1 text-xs text-red-600">
						<InfoIcon className="size-4" />
						Any ongoing sync will be auto cancelled.
					</div>
				</div>
				{!allStreamsImpacted && (
					<div className="mt-4 max-h-96 overflow-y-auto px-4">
						{renderStreamsByNamespace()}
					</div>
				)}

				{showIndexWarning && (
					<div className="mt-4 flex gap-3 rounded-lg border border-warning/40 bg-warning-light p-4 text-left text-sm text-warning-dark">
						<InfoIcon className="mt-0.5 size-5 shrink-0" />
						<span>
							Indexes will be built for streams using position deletes or
							deletion vectors. The next sync may take some time.
						</span>
					</div>
				)}
			</Modal>
		</>
	)
}

export default StreamDifferenceModal
