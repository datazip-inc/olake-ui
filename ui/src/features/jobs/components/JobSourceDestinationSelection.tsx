import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Select, Button, Spin } from "antd"
import {
	PlusIcon,
	PencilSimpleIcon,
	GitCommitIcon,
} from "@phosphor-icons/react"
import { useSources } from "@/features/sources/hooks"
import { useDestinations } from "@/features/destinations/hooks"
import { getConnectorInLowerCase } from "@/common/utils"
import {
	sourceConnectorOptions,
	destinationConnectorOptions,
} from "@/common/components/connectorOptions"
import { useJobConfigurationStore } from "../stores"

const JobSourceDestinationSelection: React.FC = () => {
	const navigate = useNavigate()
	const {
		selectedSource,
		selectedDestination,
		setSelectedSource,
		setSelectedDestination,
		isEditMode,
	} = useJobConfigurationStore()

	const [sourceConnector, setSourceConnector] = useState<string | null>(
		"MongoDB",
	)
	const [destinationConnector, setDestinationConnector] = useState<
		string | null
	>("Amazon S3")

	const { data: sourcesData, isLoading: isLoadingSources } = useSources()
	const { data: destinationsData, isLoading: isLoadingDestinations } =
		useDestinations()

	const sourceOptions = (sourcesData ?? [])
		.filter((s: any) => {
			if (!sourceConnector) return true
			return s.type === getConnectorInLowerCase(sourceConnector)
		})
		.map((s: any) => ({
			value: s.id,
			label: s.name,
		}))

	const destinationOptions = (destinationsData ?? [])
		.filter((d: any) => {
			if (!destinationConnector) return true
			return d.type === getConnectorInLowerCase(destinationConnector)
		})
		.map((d: any) => ({
			value: d.id,
			label: d.name,
		}))

	// Pre-fill local source connector dropdown state if a source is already selected
	useEffect(() => {
		if (selectedSource && !sourceConnector) {
			const matchingOption = sourceConnectorOptions.find(
				opt => getConnectorInLowerCase(opt.value) === selectedSource.type,
			)
			if (matchingOption) {
				setSourceConnector(matchingOption.value)
			}
		}
	}, [selectedSource, sourceConnector])

	// Pre-fill local destination connector dropdown state if a destination is already selected
	useEffect(() => {
		if (selectedDestination && !destinationConnector) {
			const matchingOption = destinationConnectorOptions.find(
				opt => getConnectorInLowerCase(opt.value) === selectedDestination.type,
			)
			if (matchingOption) {
				setDestinationConnector(matchingOption.value)
			}
		}
	}, [selectedDestination, destinationConnector])

	return (
		<div className="mt-5 flex flex-col gap-2 rounded-xl border border-[#D9D9D9] p-6">
			<div className="mb-4 flex items-center gap-2">
				<GitCommitIcon className="size-5" />
				<span className="text-base font-medium text-gray-900">
					Select Source & Destination
				</span>
			</div>
			<div className="flex gap-4 gap-x-6">
				{/* Source Selection */}
				<div className="flex w-full flex-col items-end gap-y-3 border-r pr-6">
					<div className="flex w-full gap-3">
						<div className="flex-1">
							<label className="mb-2 block text-sm font-medium">
								Source Connector:<span className="text-red-500">*</span>
							</label>
							<Select
								className="w-full"
								value={sourceConnector ?? undefined}
								onChange={val => {
									setSourceConnector(val)
									if (!isEditMode) setSelectedSource(null)
								}}
								options={sourceConnectorOptions}
								placeholder="Select a connector"
								data-testid="source-connector-select"
								disabled={isEditMode}
							/>
						</div>
						<div className="flex-1">
							<div className="items-between mb-2 flex justify-between text-sm font-medium">
								<div className="inline-flex items-center">
									Select existing source:
									<span className="text-red-500">*</span>
								</div>
							</div>
							{isLoadingSources ? (
								<div className="flex h-8 items-center">
									<Spin size="small" />
								</div>
							) : (
								<Select
									className="w-full"
									value={selectedSource?.id ?? undefined}
									onChange={val => {
										const source = sourcesData?.find((s: any) => s.id === val)
										setSelectedSource(source || null)
									}}
									options={sourceOptions}
									placeholder="Select a source"
									data-testid="existing-source"
									disabled={isEditMode}
								/>
							)}
						</div>
					</div>
					<div>
						{isEditMode ? (
							<Button
								type="default"
								icon={<PencilSimpleIcon className="size-4" />}
								onClick={() => navigate(`/sources/${selectedSource?.id}`)}
								className="flex items-center gap-1"
								disabled={!selectedSource}
							>
								Edit Source
							</Button>
						) : (
							<Button
								type="default"
								icon={<PlusIcon className="size-4" />}
								onClick={() => navigate("/sources/new")}
								className="flex items-center gap-1"
							>
								New Source
							</Button>
						)}
					</div>
				</div>

				{/* Destination Selection */}
				<div className="flex w-full flex-col items-end gap-3">
					<div className="flex w-full gap-3">
						<div className="flex-1">
							<label className="mb-2 block text-sm font-medium">
								Destination Connector:<span className="text-red-500">*</span>
							</label>
							<Select
								className="w-full"
								value={destinationConnector ?? undefined}
								onChange={val => {
									setDestinationConnector(val)
									if (!isEditMode) setSelectedDestination(null)
								}}
								options={destinationConnectorOptions}
								placeholder="Select a connector"
								data-testid="destination-connector-select"
								disabled={isEditMode}
							/>
						</div>
						<div className="flex-1">
							<div className="items-between mb-2 flex justify-between text-sm font-medium">
								<div className="inline-flex items-center">
									Select existing destination:
									<span className="text-red-500">*</span>
								</div>
							</div>
							{isLoadingDestinations ? (
								<div className="flex h-8 items-center">
									<Spin size="small" />
								</div>
							) : (
								<Select
									className="w-full"
									value={selectedDestination?.id ?? undefined}
									onChange={val => {
										const dest = destinationsData?.find(
											(d: any) => d.id === val,
										)
										setSelectedDestination(dest || null)
									}}
									options={destinationOptions}
									placeholder="Select a destination"
									data-testid="existing-destination"
									disabled={isEditMode}
								/>
							)}
						</div>
					</div>
					<div>
						{isEditMode ? (
							<Button
								type="default"
								icon={<PencilSimpleIcon className="size-4" />}
								onClick={() =>
									navigate(`/destinations/${selectedDestination?.id}`)
								}
								className="flex items-center gap-1"
								disabled={!selectedDestination}
							>
								Edit Destination
							</Button>
						) : (
							<Button
								type="default"
								icon={<PlusIcon className="size-4" />}
								onClick={() => navigate("/destinations/new")}
								className="flex items-center gap-1"
							>
								New Destination
							</Button>
						)}
					</div>
				</div>
			</div>
		</div>
	)
}

export default JobSourceDestinationSelection
