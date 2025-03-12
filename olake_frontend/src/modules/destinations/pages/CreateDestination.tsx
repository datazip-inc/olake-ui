import { useState, useEffect } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Select } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, GenderNeuter, Notebook } from "@phosphor-icons/react"
import TestConnectionModal from "../../common/components/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/components/TestConnectionSuccessModal"
import EntitySavedModal from "../../common/components/EntitySavedModal"

interface CreateDestinationProps {
	fromJobFlow?: boolean
	onComplete?: () => void
	stepNumber?: number
	stepTitle?: string
}

const CreateDestination: React.FC<CreateDestinationProps> = ({
	fromJobFlow,
	onComplete,
	stepNumber,
	stepTitle,
}) => {
	const navigate = useNavigate()
	const [setupType, setSetupType] = useState("new")
	const [connector, setConnector] = useState("Amazon S3")
	const [authType, setAuthType] = useState("iam")
	const [iamInfo, setIamInfo] = useState("")
	const [awsAccessKeyId, setAwsAccessKeyId] = useState("")
	const [awsSecretKey, setAwsSecretKey] = useState("")
	const [s3BucketName, setS3BucketName] = useState("")
	const [s3BucketPath, setS3BucketPath] = useState("")
	const [region, setRegion] = useState("")
	const [destinationName, setDestinationName] = useState("")
	const [filteredDestinations, setFilteredDestinations] = useState<any[]>([])

	const {
		destinations,
		fetchDestinations,
		setShowTestingModal,
		setShowSuccessModal,
		setShowEntitySavedModal,
	} = useAppStore()

	useEffect(() => {
		if (!destinations.length) {
			fetchDestinations()
		}
	}, [destinations.length, fetchDestinations])

	useEffect(() => {
		if (setupType === "existing") {
			setFilteredDestinations(
				destinations.filter(destination => destination.type === connector),
			)
		}
	}, [connector, setupType, destinations])

	const handleCancel = () => {
		if (fromJobFlow) {
			navigate("/jobs/new")
		} else {
			navigate("/destinations")
		}
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

	const handleConnectorChange = (value: string) => {
		setConnector(value)

		if (value === "Amazon S3") {
			setAuthType("iam")
			setIamInfo("")
			setAwsAccessKeyId("")
			setAwsSecretKey("")
			setS3BucketName("")
			setS3BucketPath("")
			setRegion("")
		} else if (value === "Snowflake") {
			setAuthType("keys")
			setAwsAccessKeyId("")
			setAwsSecretKey("")
		} else if (value === "BigQuery") {
			setAuthType("keys")
			setAwsAccessKeyId("")
			setAwsSecretKey("")
		} else if (value === "Redshift") {
			setAuthType("iam")
			setIamInfo("")
		}
	}

	const handleExistingDestinationSelect = (value: string) => {
		const selectedDestination = destinations.find(d => d.id === value)

		if (selectedDestination) {
			setDestinationName(selectedDestination.name)
			setConnector(selectedDestination.type)

			if (selectedDestination.type === "Amazon S3") {
				setAuthType("iam")
				setIamInfo("mock-iam-info")
				setS3BucketName("mock-bucket")
				setS3BucketPath("/mock/path")
				setRegion("us-west-2")
			} else if (selectedDestination.type === "Snowflake") {
				setAuthType("keys")
				setAwsAccessKeyId("mock-snowflake-account")
				setAwsSecretKey("mock-snowflake-password")
			} else if (selectedDestination.type === "BigQuery") {
				setAuthType("keys")
				setAwsAccessKeyId("mock-bigquery-project")
				setAwsSecretKey("mock-bigquery-key")
			} else if (selectedDestination.type === "Redshift") {
				setAuthType("iam")
				setIamInfo("mock-redshift-iam")
				setRegion("us-east-1")
			}
		}
	}

	return (
		<div className="flex h-screen flex-col">
			{/* Header */}
			{!fromJobFlow && (
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to={"/destinations"}
						className="flex items-center text-lg font-bold"
					>
						<ArrowLeft className="mr-1 size-6 font-bold" />
					</Link>
					<div className="text-xl font-bold">Create destination</div>
				</div>
			)}

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden">
				{/* Left content */}
				<div className="w-full overflow-auto p-6 pt-0">
					{stepNumber && stepTitle && (
						<div className="mb-4 flex flex-col gap-2">
							<div className="flex items-center gap-2 text-sm text-blue-600">
								<div className="size-2 rounded-full border border-blue-600 outline outline-2 outline-blue-600"></div>
								<span className="text-[#8A8A8A]">Step {stepNumber}</span>
							</div>
							<h1 className="text-xl font-medium">{stepTitle}</h1>
						</div>
					)}
					<div className="mb-6 mt-6 rounded-xl border border-gray-200 bg-white p-6">
						<div>
							<div className="mb-4 flex items-center gap-1 text-base font-medium">
								<Notebook className="size-5" />
								Capture information
							</div>
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
										Set up a new destination
									</Radio>
									<Radio value="existing">Use an existing destination</Radio>
								</Radio.Group>
							</div>

							{setupType === "new" ? (
								<div className="grid grid-cols-2 gap-6">
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<div className="flex items-center">
											<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
												<span>
													{connector === "Amazon S3"
														? "S"
														: connector.charAt(0)}
												</span>
											</div>
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="w-full"
												options={[
													{ value: "Amazon S3", label: "Amazon S3" },
													{ value: "Snowflake", label: "Snowflake" },
													{ value: "BigQuery", label: "BigQuery" },
													{ value: "Redshift", label: "Redshift" },
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your destination:
										</label>
										<Input
											placeholder="Enter the name of your destination"
											value={destinationName}
											onChange={e => setDestinationName(e.target.value)}
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
											<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
												<span>
													{connector === "Amazon S3"
														? "S"
														: connector.charAt(0)}
												</span>
											</div>
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="w-full"
												options={[
													{ value: "Amazon S3", label: "Amazon S3" },
													{ value: "Snowflake", label: "Snowflake" },
													{ value: "BigQuery", label: "BigQuery" },
													{ value: "Redshift", label: "Redshift" },
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Select existing destination:
										</label>
										<Select
											placeholder="Select a destination"
											className="w-full"
											onChange={handleExistingDestinationSelect}
											options={filteredDestinations.map(d => ({
												value: d.id,
												label: d.name,
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
							{connector === "Amazon S3" && (
								<>
									<div className="mb-4 flex">
										<Radio.Group
											value={authType}
											onChange={e => setAuthType(e.target.value)}
											className="flex"
										>
											<Radio
												value="iam"
												className="mr-8"
											>
												IAM
											</Radio>
											<Radio value="keys">Access keys</Radio>
										</Radio.Group>
									</div>

									{authType === "iam" ? (
										<div className="mb-4">
											<label className="mb-2 block text-sm font-medium text-gray-700">
												IAM info:
											</label>
											<Input
												placeholder="Enter your IAM info"
												value={iamInfo}
												onChange={e => setIamInfo(e.target.value)}
											/>
										</div>
									) : (
										<div className="mb-4 grid grid-cols-2 gap-6">
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													AWS access key ID:
												</label>
												<Input
													placeholder="Enter your AWS access key ID"
													value={awsAccessKeyId}
													onChange={e => setAwsAccessKeyId(e.target.value)}
												/>
											</div>
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													AWS secret key:
												</label>
												<Input.Password
													placeholder="Enter your AWS secret key"
													value={awsSecretKey}
													onChange={e => setAwsSecretKey(e.target.value)}
												/>
											</div>
										</div>
									)}

									<div className="mb-4 grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												S3 bucket name:
											</label>
											<Input
												placeholder="Enter your S3 bucket name"
												value={s3BucketName}
												onChange={e => setS3BucketName(e.target.value)}
											/>
										</div>
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												S3 bucket path:
											</label>
											<Input
												placeholder="Enter your S3 bucket path"
												value={s3BucketPath}
												onChange={e => setS3BucketPath(e.target.value)}
											/>
										</div>
									</div>

									<div className="mb-4">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Region:
										</label>
										<Select
											placeholder="Select AWS region"
											className="w-full"
											value={region || undefined}
											onChange={value => setRegion(value)}
											options={[
												{ value: "us-east-1", label: "US East (N. Virginia)" },
												{ value: "us-east-2", label: "US East (Ohio)" },
												{
													value: "us-west-1",
													label: "US West (N. California)",
												},
												{ value: "us-west-2", label: "US West (Oregon)" },
												{ value: "eu-west-1", label: "EU (Ireland)" },
												{ value: "eu-central-1", label: "EU (Frankfurt)" },
												{ value: "ap-south-1", label: "Asia Pacific (Mumbai)" },
												{
													value: "ap-northeast-1",
													label: "Asia Pacific (Tokyo)",
												},
											]}
										/>
									</div>
								</>
							)}

							{/* Add other connector configurations here */}
						</div>
					</div>
				</div>

				{/* Documentation panel */}
				<div className="h-[calc(100vh-120px)] w-1/4 overflow-hidden border-l border-gray-200 bg-white">
					<div className="flex items-center border-b border-gray-200 p-4">
						<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-white">
							<span className="font-bold">
								{connector === "Amazon S3" ? "S" : connector.charAt(0)}
							</span>
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

			{/* Footer */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<Button
					onClick={handleCancel}
					className="border border-[#F5222D] text-[#F5222D]"
				>
					Cancel
				</Button>
				<Button
					type="primary"
					className="font-thin"
					onClick={handleCreate}
				>
					Create →
				</Button>
			</div>

			<TestConnectionModal />

			<TestConnectionSuccessModal />

			<EntitySavedModal
				type="destination"
				onComplete={onComplete}
				fromJobFlow={fromJobFlow || false}
			/>
		</div>
	)
}

export default CreateDestination
