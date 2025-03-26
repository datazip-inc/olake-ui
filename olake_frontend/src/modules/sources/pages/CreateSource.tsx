import { useState, useEffect, useRef } from "react"
import { Link } from "react-router-dom"
import { Input, Radio, Select, Switch } from "antd"
import { useAppStore } from "../../../store"
import {
	ArrowLeft,
	ArrowRight,
	Control,
	GearFine,
	GenderNeuter,
	Notebook,
} from "@phosphor-icons/react"
import TestConnectionModal from "../../common/components/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/components/TestConnectionSuccessModal"
import EntitySavedModal from "../../common/components/EntitySavedModal"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import EntityCancelModal from "../../common/components/EntityCancelModal"
import StepTitle from "../../common/components/StepTitle"

interface CreateSourceProps {
	fromJobFlow?: boolean
	fromJobEditFlow?: boolean
	existingSourceId?: string
	onComplete?: () => void
	stepNumber?: string
	stepTitle?: string
}

const CreateSource: React.FC<CreateSourceProps> = ({
	fromJobFlow = false,
	fromJobEditFlow = false,
	existingSourceId,
	onComplete,
	stepNumber,
	stepTitle,
}) => {
	const [setupType, setSetupType] = useState("new")
	const [connector, setConnector] = useState("MongoDB")
	const [connectionType, setConnectionType] = useState("uri")
	const [connectionUri, setConnectionUri] = useState("")
	const [sourceName, setSourceName] = useState("")
	const [srvEnabled, setSrvEnabled] = useState(false)
	const {
		setShowEntitySavedModal,
		setShowSourceCancelModal,
		setShowTestingModal,
		setShowSuccessModal,
	} = useAppStore()
	const [showAdvanced, setShowAdvanced] = useState(false)
	const [isDocPanelCollapsed, setIsDocPanelCollapsed] = useState(false)
	const iframeRef = useRef<HTMLIFrameElement>(null)

	const [filteredSources, setFilteredSources] = useState<any[]>([])

	const { sources, fetchSources } = useAppStore()

	useEffect(() => {
		if (!sources.length) {
			fetchSources()
		}
	}, [sources.length, fetchSources])

	useEffect(() => {
		if (fromJobEditFlow && existingSourceId) {
			setSetupType("existing")

			const selectedSource = sources.find(s => s.id === existingSourceId)

			if (selectedSource) {
				setSourceName(selectedSource.name)
				setConnector(selectedSource.type)

				if (selectedSource.type === "MongoDB") {
					setConnectionType("uri")
					setConnectionUri("mongodb://username:password@hostname:port/database")
					setSrvEnabled(true)
				} else if (selectedSource.type === "PostgreSQL") {
					setConnectionType("uri")
					setConnectionUri(
						"postgresql://username:password@hostname:port/database",
					)
					setSrvEnabled(false)
				} else if (selectedSource.type === "MySQL") {
					setConnectionType("hosts")
					setConnectionUri("mysql://username:password@hostname:port/database")
					setSrvEnabled(false)
				} else if (selectedSource.type === "Kafka") {
					setConnectionType("hosts")
					setConnectionUri("kafka://hostname:port")
					setSrvEnabled(false)
				}
			}
		}
	}, [fromJobEditFlow, existingSourceId, sources])

	useEffect(() => {
		if (setupType === "existing") {
			setFilteredSources(sources.filter(source => source.type === connector))
		}
	}, [connector, setupType, sources])

	const handleCancel = () => {
		setShowSourceCancelModal(true)
	}

	const handleCreate = () => {
		setTimeout(() => {
			setShowTestingModal(true)
			setTimeout(() => {
				setShowTestingModal(false)
				setShowSuccessModal(true)
				setTimeout(() => {
					setShowSuccessModal(false)
					setShowEntitySavedModal(true)
				}, 2000)
			}, 2000)
		}, 2000)
	}

	const toggleAdvancedConfig = () => {
		setShowAdvanced(!showAdvanced)
	}

	const handleConnectorChange = (value: string) => {
		setConnector(value)

		if (value === "MongoDB") {
			setConnectionType("uri")
			setConnectionUri("")
		} else if (value === "PostgreSQL") {
			setConnectionType("uri")
			setConnectionUri("")
		} else if (value === "MySQL") {
			setConnectionType("hosts")
			setConnectionUri("")
		} else if (value === "Kafka") {
			setConnectionType("hosts")
			setConnectionUri("")
		}

		// Update iframe src when connector changes
		if (iframeRef.current) {
			iframeRef.current.src = `https://olake.io/docs/category/${value.toLowerCase()}`
		}
	}

	const handleExistingSourceSelect = (value: string) => {
		const selectedSource = sources.find(s => s.id === value)

		if (selectedSource) {
			setSourceName(selectedSource.name)
			setConnector(selectedSource.type)

			if (selectedSource.type === "MongoDB") {
				setConnectionType("uri")
				setConnectionUri("mongodb://username:password@hostname:port/database")
				setSrvEnabled(true)
			} else if (selectedSource.type === "PostgreSQL") {
				setConnectionType("uri")
				setConnectionUri(
					"postgresql://username:password@hostname:port/database",
				)
				setSrvEnabled(false)
			} else if (selectedSource.type === "MySQL") {
				setConnectionType("hosts")
				setConnectionUri("mysql://username:password@hostname:port/database")
				setSrvEnabled(false)
			} else if (selectedSource.type === "Kafka") {
				setConnectionType("hosts")
				setConnectionUri("kafka://broker1:9092,broker2:9092/topic")
				setSrvEnabled(false)
			} else if (selectedSource.type === "REST API") {
				setConnectionType("uri")
				setConnectionUri("https://api.example.com/data")
				setSrvEnabled(false)
			}
		}
	}

	const toggleDocPanel = () => {
		setIsDocPanelCollapsed(!isDocPanelCollapsed)
	}

	return (
		<div className="flex h-screen flex-col">
			{!fromJobFlow && (
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to={"/sources"}
						className="flex items-center text-lg font-bold"
					>
						<ArrowLeft className="mr-1 size-6 font-bold" />
					</Link>
					<div className="text-lg font-bold">Create source</div>
				</div>
			)}

			<div className="flex flex-1 overflow-hidden">
				<div className="w-full overflow-auto p-6 pt-0">
					{stepNumber && stepTitle && (
						<StepTitle
							stepNumber={stepNumber}
							stepTitle={stepTitle}
						/>
					)}
					<div className="mb-6 mt-2 rounded-xl border border-gray-200 bg-white p-6">
						<div className="mb-6">
							<div className="mb-4 flex items-center gap-1 text-base font-medium">
								<Notebook className="size-5" />
								Capture information
							</div>

							{!fromJobEditFlow && (
								<div className="mb-4 flex">
									<Radio.Group
										value={setupType}
										onChange={e => setSetupType(e.target.value)}
										className="flex"
									>
										<Radio
											value="new"
											className="mr-8"
										>
											Set up a new source
										</Radio>
										<Radio value="existing">Use an existing source</Radio>
									</Radio.Group>
								</div>
							)}

							{setupType === "new" && !fromJobEditFlow ? (
								<div className="grid grid-cols-2 gap-6">
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<div className="flex items-center">
											<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-white">
												<span>{connector.charAt(0)}</span>
											</div>
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="w-full"
												options={[
													{ value: "MongoDB", label: "MongoDB" },
													{ value: "PostgreSQL", label: "PostgreSQL" },
													{ value: "MySQL", label: "MySQL" },
													{ value: "Kafka", label: "Kafka" },
													{ value: "REST API", label: "REST API" },
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your source:
										</label>
										<Input
											placeholder="Enter the name of your source"
											value={sourceName}
											onChange={e => setSourceName(e.target.value)}
										/>
									</div>
								</div>
							) : (
								<div className="grid grid-cols-2 gap-6">
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<div className="flex items-center">
											<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-white">
												<span>{connector.charAt(0)}</span>
											</div>
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="w-full"
												disabled={fromJobEditFlow}
												options={[
													{ value: "MongoDB", label: "MongoDB" },
													{ value: "PostgreSQL", label: "PostgreSQL" },
													{ value: "MySQL", label: "MySQL" },
													{ value: "Kafka", label: "Kafka" },
													{ value: "REST API", label: "REST API" },
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											{fromJobEditFlow ? "Source:" : "Select existing source:"}
										</label>
										<Select
											placeholder="Select a source"
											className="w-full"
											onChange={handleExistingSourceSelect}
											value={fromJobEditFlow ? existingSourceId : undefined}
											disabled={fromJobEditFlow}
											options={filteredSources.map(s => ({
												value: s.id,
												label: s.name,
											}))}
										/>
									</div>
								</div>
							)}
						</div>
					</div>

					<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6">
						<div className="mb-4 flex items-center">
							<div className="mb-2 flex items-center gap-1">
								<GenderNeuter className="size-5" />
								<div className="text-base font-medium">Endpoint config</div>
							</div>
						</div>

						<div className="mb-6">
							<div className="mb-4 flex">
								<Radio.Group
									value={connectionType}
									onChange={e => setConnectionType(e.target.value)}
									className="flex"
									disabled={fromJobEditFlow}
								>
									<Radio
										value="uri"
										className="mr-8"
									>
										Connection URI
									</Radio>
									<Radio value="hosts">Hosts</Radio>
								</Radio.Group>
							</div>

							{connectionType === "uri" ? (
								<div className="mb-4">
									<label className="mb-2 block text-sm font-medium text-gray-700">
										Connection URI:
									</label>
									<Input
										placeholder={`Enter your ${connector} connection URI`}
										value={connectionUri}
										onChange={e => setConnectionUri(e.target.value)}
										disabled={fromJobEditFlow}
									/>
									{connector === "MongoDB" && (
										<p className="mt-1 text-xs text-gray-500">
											Example:
											mongodb://username:password@hostname:port/database
										</p>
									)}
									{connector === "PostgreSQL" && (
										<p className="mt-1 text-xs text-gray-500">
											Example:
											postgresql://username:password@hostname:port/database
										</p>
									)}
								</div>
							) : (
								<div className="grid grid-cols-2 gap-6">
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Host:
										</label>
										<Input placeholder="Enter host" />
									</div>
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Port:
										</label>
										<Input placeholder="Enter port" />
									</div>
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Username:
										</label>
										<Input placeholder="Enter username" />
									</div>
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Password:
										</label>
										<Input.Password placeholder="Enter password" />
									</div>
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Database:
										</label>
										<Input placeholder="Enter database name" />
									</div>
								</div>
							)}
						</div>
					</div>

					<div className="mb-6 select-none rounded-xl border border-gray-200 bg-white p-6">
						<div
							className="flex cursor-pointer items-center justify-between"
							onClick={toggleAdvancedConfig}
						>
							<div className="flex items-center">
								<div className="flex items-center gap-1">
									<GearFine className="size-5" />
									<div className="font-medium">Advanced configurations</div>
								</div>
							</div>
							<span
								className={`transform transition-transform ${
									showAdvanced ? "rotate-180" : ""
								}`}
							>
								<Control className="size-5" />
							</span>
						</div>

						{showAdvanced && (
							<div className="mt-4 border-t border-gray-200 pt-4">
								{connector === "MongoDB" && (
									<div className="grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Replica set:
											</label>
											<Input placeholder="Enter replica set" />
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												AuthDB:
											</label>
											<Input placeholder="Enter auth database" />
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Read preference:
											</label>
											<Select
												placeholder="Select read preference"
												className="w-full"
												options={[
													{ value: "primary", label: "Primary" },
													{
														value: "primaryPreferred",
														label: "Primary Preferred",
													},
													{ value: "secondary", label: "Secondary" },
													{
														value: "secondaryPreferred",
														label: "Secondary Preferred",
													},
													{ value: "nearest", label: "Nearest" },
												]}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Server RAM:
											</label>
											<Input placeholder="Enter server RAM (MB)" />
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Max threads:
											</label>
											<Input placeholder="Enter max threads" />
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Default mode:
											</label>
											<Select
												placeholder="Select default mode"
												className="w-full"
												options={[
													{ value: "standard", label: "Standard" },
													{ value: "legacy", label: "Legacy" },
												]}
											/>
										</div>
										<div className="col-span-2">
											<div className="flex items-center justify-between">
												<span className="font-medium">SRV</span>
												<Switch
													checked={srvEnabled}
													onChange={setSrvEnabled}
													className={srvEnabled ? "bg-blue-600" : ""}
												/>
											</div>
										</div>
									</div>
								)}

								{connector === "PostgreSQL" && (
									<div className="grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Schema:
											</label>
											<Input
												placeholder="Enter schema name"
												defaultValue="public"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												SSL Mode:
											</label>
											<Select
												placeholder="Select SSL mode"
												className="w-full"
												defaultValue="prefer"
												options={[
													{ value: "disable", label: "Disable" },
													{ value: "prefer", label: "Prefer" },
													{ value: "require", label: "Require" },
													{ value: "verify-ca", label: "Verify CA" },
													{ value: "verify-full", label: "Verify Full" },
												]}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Connection timeout:
											</label>
											<Input
												placeholder="Enter timeout (seconds)"
												defaultValue="30"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Max connections:
											</label>
											<Input
												placeholder="Enter max connections"
												defaultValue="10"
											/>
										</div>
									</div>
								)}

								{connector === "MySQL" && (
									<div className="grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Character set:
											</label>
											<Input
												placeholder="Enter character set"
												defaultValue="utf8mb4"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Collation:
											</label>
											<Input
												placeholder="Enter collation"
												defaultValue="utf8mb4_unicode_ci"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Connection timeout:
											</label>
											<Input
												placeholder="Enter timeout (seconds)"
												defaultValue="30"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Max connections:
											</label>
											<Input
												placeholder="Enter max connections"
												defaultValue="10"
											/>
										</div>
									</div>
								)}

								{connector === "Kafka" && (
									<div className="grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Group ID:
											</label>
											<Input placeholder="Enter consumer group ID" />
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Auto offset reset:
											</label>
											<Select
												placeholder="Select offset reset"
												className="w-full"
												defaultValue="latest"
												options={[
													{ value: "earliest", label: "Earliest" },
													{ value: "latest", label: "Latest" },
												]}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Security protocol:
											</label>
											<Select
												placeholder="Select security protocol"
												className="w-full"
												defaultValue="plaintext"
												options={[
													{ value: "plaintext", label: "PLAINTEXT" },
													{ value: "ssl", label: "SSL" },
													{ value: "sasl_plaintext", label: "SASL_PLAINTEXT" },
													{ value: "sasl_ssl", label: "SASL_SSL" },
												]}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Session timeout:
											</label>
											<Input
												placeholder="Enter timeout (ms)"
												defaultValue="30000"
											/>
										</div>
									</div>
								)}

								{connector === "REST API" && (
									<div className="grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Authentication type:
											</label>
											<Select
												placeholder="Select authentication type"
												className="w-full"
												defaultValue="none"
												options={[
													{ value: "none", label: "None" },
													{ value: "basic", label: "Basic Auth" },
													{ value: "bearer", label: "Bearer Token" },
													{ value: "oauth2", label: "OAuth 2.0" },
												]}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Request timeout:
											</label>
											<Input
												placeholder="Enter timeout (seconds)"
												defaultValue="30"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Retry attempts:
											</label>
											<Input
												placeholder="Enter retry attempts"
												defaultValue="3"
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Retry delay:
											</label>
											<Input
												placeholder="Enter retry delay (seconds)"
												defaultValue="5"
											/>
										</div>
									</div>
								)}
							</div>
						)}
					</div>
				</div>

				<DocumentationPanel
					docUrl={`https://olake.io/docs/category/${connector.toLowerCase()}`}
					isMinimized={isDocPanelCollapsed}
					onToggle={toggleDocPanel}
					showResizer={true}
				/>
			</div>

			{/* Footer */}
			{!fromJobFlow && !fromJobEditFlow && (
				<div className="flex justify-between border-t border-gray-200 bg-white p-4">
					<button
						onClick={handleCancel}
						className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
					>
						Cancel
					</button>
					<button
						className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
						onClick={handleCreate}
					>
						Next
						<ArrowRight className="size-4 text-white" />
					</button>
				</div>
			)}

			<TestConnectionModal />

			<TestConnectionSuccessModal />

			<EntitySavedModal
				type="source"
				onComplete={onComplete}
				fromJobFlow={fromJobFlow || false}
			/>
			<EntityCancelModal
				type="source"
				navigateTo={
					fromJobEditFlow ? "jobs" : fromJobFlow ? "jobs/new" : "sources"
				}
			/>
		</div>
	)
}

export default CreateSource
