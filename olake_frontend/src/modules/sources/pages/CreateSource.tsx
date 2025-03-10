import { useState, useEffect } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Select, Switch, message, Modal } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"
import DestinationSuccess from "../../../assets/DestinationSuccess.png"
interface CreateSourceProps {
	fromJobFlow?: boolean
	onComplete?: () => void
}

const CreateSource: React.FC<CreateSourceProps> = ({
	fromJobFlow,
	onComplete,
}) => {
	const navigate = useNavigate()
	const [setupType, setSetupType] = useState("new")
	const [connector, setConnector] = useState("MongoDB")
	const [connectionType, setConnectionType] = useState("uri")
	const [connectionUri, setConnectionUri] = useState("")
	const [sourceName, setSourceName] = useState("")
	const [srvEnabled, setSrvEnabled] = useState(false)
	const [showAdvanced, setShowAdvanced] = useState(false)
	const [showTestingModal, setShowTestingModal] = useState(false)
	const [showSuccessModal, setShowSuccessModal] = useState(false)
	const [showSourceSavedModal, setShowSourceSavedModal] = useState(false)

	const [filteredSources, setFilteredSources] = useState<any[]>([])

	const { sources, fetchSources, addSource } = useAppStore()

	useEffect(() => {
		if (!sources.length) {
			fetchSources()
		}
	}, [sources.length, fetchSources])

	useEffect(() => {
		if (setupType === "existing") {
			setFilteredSources(sources.filter(source => source.type === connector))
		}
	}, [connector, setupType, sources])

	const handleCancel = () => {
		if (fromJobFlow) {
			navigate("/jobs/new")
		} else {
			navigate("/sources")
		}
	}

	const handleCreate = () => {
		setShowTestingModal(true)

		setTimeout(() => {
			setShowTestingModal(false)
			setShowSuccessModal(true)

			setTimeout(() => {
				setShowSuccessModal(false)
				setShowSourceSavedModal(true)

				setTimeout(() => {
					const sourceData = {
						name:
							sourceName ||
							`${connector}_source_${Math.floor(Math.random() * 1000)}`,
						type: connector,
						status: "active" as const,
					}

					addSource(sourceData)
						.then(() => {
							setShowSourceSavedModal(false)
							if (fromJobFlow) {
								if (onComplete) {
									onComplete()
								} else {
									navigate("/jobs/new")
								}
							} else {
								navigate("/sources")
							}
						})
						.catch(error => {
							message.error("Failed to create source")
							console.error(error)
						})
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

	return (
		<div className="flex h-screen flex-col">
			<div className="border-b border-gray-200 p-6 pb-0">
				<Link
					to={fromJobFlow ? "/jobs/new" : "/sources"}
					className="mb-4 flex items-center text-blue-600"
				>
					<ArrowLeft
						size={16}
						className="mr-1"
					/>{" "}
					{fromJobFlow ? "Back to Job Creation" : "Create source"}
				</Link>
			</div>

			<div className="flex flex-1 overflow-hidden">
				<div className="w-3/4 overflow-auto p-6 pt-0">
					<div className="mb-6 rounded-lg border border-gray-200 bg-white p-6">
						<div className="mb-4 flex items-center">
							<div className="mr-2 flex h-5 w-5 items-center justify-center rounded bg-gray-200 text-gray-600">
								<span className="text-xs">📋</span>
							</div>
							<h3 className="text-lg font-medium">Capture information</h3>
						</div>

						<div className="mb-6">
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

							{setupType === "new" ? (
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
											Select existing source:
										</label>
										<Select
											placeholder="Select a source"
											className="w-full"
											onChange={handleExistingSourceSelect}
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

					<div className="mb-6 rounded-lg border border-gray-200 bg-white p-6">
						<div className="mb-4 flex items-center">
							<div className="mr-2 flex h-5 w-5 items-center justify-center rounded bg-gray-200 text-gray-600">
								<span className="text-xs">🔌</span>
							</div>
							<h3 className="text-lg font-medium">Endpoint config</h3>
						</div>

						<div className="mb-6">
							<div className="mb-4 flex">
								<Radio.Group
									value={connectionType}
									onChange={e => setConnectionType(e.target.value)}
									className="flex"
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

					<div className="mb-6 rounded-lg border border-gray-200 bg-white p-6">
						<div
							className="flex cursor-pointer items-center justify-between"
							onClick={toggleAdvancedConfig}
						>
							<div className="flex items-center">
								<div className="mr-2 flex h-5 w-5 items-center justify-center rounded bg-gray-200 text-gray-600">
									<span className="text-xs">⚙️</span>
								</div>
								<h3 className="text-lg font-medium">Advanced configurations</h3>
							</div>
							<span
								className={`transform transition-transform ${
									showAdvanced ? "rotate-180" : ""
								}`}
							>
								▼
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

				<div className="h-[calc(100vh-120px)] w-1/4 overflow-hidden border-l border-gray-200 bg-white">
					<div className="flex items-center border-b border-gray-200 p-4">
						<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-white">
							<span className="font-bold">{connector.charAt(0)}</span>
						</div>
						<span className="text-lg font-bold">{connector}</span>
					</div>

					<iframe
						src="https://olake.io/docs/category/mongodb"
						className="h-[calc(100%-64px)] w-full"
						title="Documentation"
					/>
				</div>
			</div>

			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<Button
					danger
					onClick={handleCancel}
				>
					Cancel
				</Button>
				<Button
					type="primary"
					className="bg-blue-600"
					onClick={handleCreate}
				>
					Create →
				</Button>
			</div>

			<Modal
				open={showTestingModal}
				footer={null}
				closable={false}
				centered
				width={400}
			>
				<div className="flex flex-col items-center justify-center py-8">
					<div className="mb-4 flex items-center justify-center">
						<div className="flex h-12 w-12 items-center justify-center rounded-full border-2 border-blue-600">
							<div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"></div>
						</div>
					</div>
					<p className="mb-2 text-gray-500">Please wait...</p>
					<h2 className="font-semibol d text-xl">Testing your connection</h2>
				</div>
			</Modal>

			<Modal
				open={showSuccessModal}
				footer={null}
				closable={false}
				centered
				width={400}
			>
				<div className="flex flex-col items-center justify-center py-8">
					<img
						src={DestinationSuccess}
						className="h-12 w-12"
					/>{" "}
					<p className="mb-2 font-medium text-green-500">Successful</p>
					<h2 className="mb-2 text-xl font-semibold">
						Your test connection is successful
					</h2>
				</div>
			</Modal>

			<Modal
				open={showSourceSavedModal}
				footer={null}
				closable={false}
				centered
				width={400}
			>
				<div className="flex flex-col items-center justify-center py-8">
					<div className="mb-4">
						<div className="h-8 w-8 text-center">✨</div>
					</div>
					<h2 className="mb-4 text-xl font-semibold">
						Source is connected and saved successfully
					</h2>
					<div className="mb-4 flex w-full items-center rounded-lg bg-gray-100 px-4 py-2">
						<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-blue-600 text-white">
							<span>S</span>
						</div>
						<span>&lt;Source-Name&gt;</span>
						<span className="ml-auto text-green-500">Success</span>
					</div>
					<div className="flex space-x-4">
						<Button
							onClick={() => {
								setShowSourceSavedModal(false)
								if (fromJobFlow) {
									if (onComplete) {
										onComplete()
									} else {
										navigate("/jobs/new")
									}
								} else {
									navigate("/sources")
								}
							}}
						>
							{fromJobFlow ? "Back to Job Creation" : "Sources"}
						</Button>
						{!fromJobFlow && (
							<Button
								type="primary"
								className="bg-blue-600"
								onClick={() => {
									setShowSourceSavedModal(false)
									navigate("/jobs/new")
								}}
							>
								Create a job →
							</Button>
						)}
					</div>
				</div>
			</Modal>
		</div>
	)
}

export default CreateSource
