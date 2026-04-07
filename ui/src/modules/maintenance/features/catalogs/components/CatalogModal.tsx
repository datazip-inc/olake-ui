import { QuestionIcon } from "@phosphor-icons/react"
import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"
import { Button, message, Modal, Select, Spin, Tooltip } from "antd"
import { useEffect, useRef, useState } from "react"
import { useNavigate } from "react-router-dom"

import ArrayFieldTemplate from "@/common/components/form/ArrayFieldTemplate"
import CustomFieldTemplate from "@/common/components/form/CustomFieldTemplate"
import ObjectFieldTemplate from "@/common/components/form/ObjectFieldTemplate"
import { widgets } from "@/common/components/form/widgets"
import {
	ErrorLogsModal,
	TestConnectionFailureModal,
	TestConnectionModal,
	TestConnectionSuccessModal,
} from "@/common/components/modals"
import { transformErrors, TEST_CONNECTION_STATUS } from "@/common/constants"
import { TestConnectionError } from "@/common/types"
import { trimFormDataStrings, handleSpecResponse } from "@/common/utils"

import {
	useCatalogDetails,
	useIcebergDestinations,
	useCatalogSpec,
	useCatalogVersions,
	useCreateCatalog,
	useTestCatalogConnection,
	useUpdateCatalog,
} from "../hooks"
import type { CatalogModalProps, CatalogFormData } from "../types"
import CatalogSuccessModal from "./CatalogSuccessModal"

enum ActiveCatalogModalState {
	TESTING = "testing",
	TEST_SUCCESS = "testSuccess",
	PENDING_CATALOG_SAVE = "pendingCatalogSave",
	TEST_FAILURE = "testFailure",
	SPEC_FAILED = "specFailed",
	CREATION_FAILED = "creationFailed",
	CATALOG_SUCCESS = "catalogSuccess",
}

const getCatalogNameFromFormData = (data: CatalogFormData): string => {
	const { catalog_name } =
		(data as { writer: { catalog_name: string } }).writer ?? ""
	return catalog_name.trim()
}

/** API expects the writer object only, not `{ type, writer }`. */
const getCatalogWriterPayload = (
	data: CatalogFormData,
	olake_imported?: boolean,
): Record<string, unknown> => {
	const writer = (data as { writer?: Record<string, unknown> }).writer
	if (!writer || typeof writer !== "object") {
		throw new Error("Missing catalog writer configuration")
	}
	if (!olake_imported) {
		return writer
	}
	return {
		...writer,
		olake_imported: true,
	}
}

const getLabelWithTooltip = (name: string) => (
	<Tooltip title={name}>
		<span className="block truncate">{name}</span>
	</Tooltip>
)

const CatalogModal: React.FC<CatalogModalProps> = ({
	open,
	onClose,
	onSuccess,
	catalogName,
}) => {
	const isEditMode = !!catalogName

	const formRef = useRef<any>(null)
	const [formData, setFormData] = useState<CatalogFormData>({})
	const [schema, setSchema] = useState<any>(null)
	const [uiSchema, setUiSchema] = useState<any>(null)
	const [activeModal, setActiveModal] =
		useState<ActiveCatalogModalState | null>(null)
	const [specError, setSpecError] = useState<string | null>(null)
	const [creationError, setCreationError] = useState<string | null>(null)
	const [testConnectionError, setTestConnectionError] =
		useState<TestConnectionError | null>(null)
	const [createdCatalogName, setCreatedCatalogName] = useState("")
	const [selectedIcebergDestinationId, setSelectedIcebergDestinationId] =
		useState<string | null>(null)
	const timeoutRef = useRef<NodeJS.Timeout | null>(null)

	useEffect(() => {
		return () => {
			if (timeoutRef.current) {
				clearTimeout(timeoutRef.current)
			}
		}
	}, [])

	const navigate = useNavigate()
	const { data: versionsData, isLoading: loadingVersions } =
		useCatalogVersions(open)
	const {
		data: icebergDestinations = [],
		isLoading: loadingIcebergDestinations,
	} = useIcebergDestinations(open && !isEditMode)
	const icebergDestinationOptions = icebergDestinations.map(destination => ({
		value: destination.id.toString(),
		label: getLabelWithTooltip(destination.name),
	}))
	const versions = versionsData?.version ?? []
	const latestVersion = versions[0] ?? ""

	const {
		data: specData,
		isLoading: loadingSpec,
		error: specQueryError,
		refetch: refetchSpec,
	} = useCatalogSpec(latestVersion, isEditMode, open)

	const {
		data: catalogDetails,
		isLoading: loadingDetails,
		isError: isDetailsError,
		refetch: refetchDetails,
	} = useCatalogDetails(catalogName ?? "")

	const createCatalogMutation = useCreateCatalog()
	const updateCatalogMutation = useUpdateCatalog()
	const testCatalogMutation = useTestCatalogConnection()

	useEffect(() => {
		if (specData) {
			handleSpecResponse(specData, setSchema, setUiSchema, "destination")
		}
	}, [specData])

	useEffect(() => {
		if (specQueryError) {
			setSchema({})
			setUiSchema({})
			const errMsg =
				specQueryError instanceof Error
					? specQueryError.message
					: "Failed to fetch spec. Please try again."
			message.error(errMsg)
			setSpecError(errMsg)
			setActiveModal(ActiveCatalogModalState.SPEC_FAILED)
		}
	}, [specQueryError])

	useEffect(() => {
		if (isEditMode && catalogDetails) {
			setFormData(catalogDetails)
		}
	}, [isEditMode, catalogDetails])

	useEffect(() => {
		if (!open) {
			setFormData({})
			setSelectedIcebergDestinationId(null)
			setActiveModal(null)
		}
	}, [open])

	const handleIcebergDestinationSelect = (value: string) => {
		setSelectedIcebergDestinationId(value)
		const selectedDestination = icebergDestinations.find(
			destination => destination.id.toString() === value,
		)
		if (!selectedDestination) return

		// fill form data with destination config
		try {
			setFormData(JSON.parse(selectedDestination.config))
		} catch {
			message.error("Failed to load destination config")
		}
	}

	const validateForm = async (): Promise<boolean> => {
		if (schema && formRef.current) {
			return formRef.current.validateForm()
		}

		return false
	}

	const clearPendingTimeout = () => {
		if (timeoutRef.current) {
			clearTimeout(timeoutRef.current)
			timeoutRef.current = null
		}
	}

	const handleConnect = async () => {
		const isValid = await validateForm()
		if (!isValid) return

		const testCatalogData = {
			version: latestVersion,
			config: JSON.stringify(formData),
		}

		setActiveModal(ActiveCatalogModalState.TESTING)
		try {
			const testResult = await testCatalogMutation.mutateAsync({
				catalog: testCatalogData,
			})
			const testSucceeded =
				testResult.data?.connection_result.status ===
				TEST_CONNECTION_STATUS.SUCCEEDED
			if (!testSucceeded) {
				setTestConnectionError({
					message: testResult.data?.connection_result.message || "",
					logs: testResult.data?.logs || [],
				})
				setActiveModal(ActiveCatalogModalState.TEST_FAILURE)
				return
			}

			setActiveModal(ActiveCatalogModalState.TEST_SUCCESS)
			clearPendingTimeout()
			timeoutRef.current = setTimeout(() => {
				void (async () => {
					if (!timeoutRef.current) return
					timeoutRef.current = null
					setActiveModal(ActiveCatalogModalState.PENDING_CATALOG_SAVE)
					try {
						if (isEditMode) {
							await updateCatalogMutation.mutateAsync({
								catalogName: catalogName!,
								config: getCatalogWriterPayload(formData) as CatalogFormData,
							})
							setCreatedCatalogName(catalogName!)
						} else {
							await createCatalogMutation.mutateAsync(
								getCatalogWriterPayload(
									formData,
									!!selectedIcebergDestinationId,
								) as CatalogFormData,
							)
							setCreatedCatalogName(getCatalogNameFromFormData(formData))
						}
						setActiveModal(ActiveCatalogModalState.CATALOG_SUCCESS)
					} catch (e) {
						const err = e as {
							response?: { data?: { message?: string } }
							message?: string
						}
						setCreationError(
							err?.response?.data?.message ||
								err?.message ||
								(isEditMode
									? "Failed to update the catalog"
									: "Failed to create the catalog"),
						)
						setActiveModal(ActiveCatalogModalState.CREATION_FAILED)
					}
				})()
			}, 1000)
		} catch {
			setActiveModal(null)
			message.error("Test connection failed. Please try again.")
		}
	}

	const handleViewCatalogs = () => {
		clearPendingTimeout()
		setActiveModal(null)
		onClose()
		onSuccess?.()
	}

	const handleCancel = () => {
		clearPendingTimeout()
		setActiveModal(null)
		onClose()
	}

	const isLoading =
		loadingVersions || loadingSpec || (isEditMode && loadingDetails)
	const canSubmit =
		!!schema && !!latestVersion && !isLoading && !(isEditMode && isDetailsError)

	return (
		<>
			<Modal
				open={open && activeModal === null}
				onCancel={handleCancel}
				title={
					<span className="text-xl font-medium leading-7 text-olake-text">
						{isEditMode ? "Edit Catalog" : "Add New Catalog"}
					</span>
				}
				footer={null}
				width={680}
				centered
				destroyOnHidden
			>
				{!isEditMode &&
					(loadingIcebergDestinations || icebergDestinations.length > 0) && (
						<div className="mt-4 rounded-md">
							<div className="flex items-center gap-1">
								<p className="text-sm font-medium leading-[22px] text-olake-text">
									Import Catalog from destination
								</p>
								<Tooltip title="Select a destination to auto-fill the catalog with its credentials">
									<QuestionIcon
										size={14}
										className="cursor-help text-olake-text-tertiary"
									/>
								</Tooltip>
							</div>
							{loadingIcebergDestinations ? (
								<div className="mt-2">
									<Spin size="small" />
								</div>
							) : (
								<div className="mt-2">
									<Select
										className="w-1/2 [&_.ant-select-selection-item]:truncate"
										value={selectedIcebergDestinationId}
										onChange={handleIcebergDestinationSelect}
										options={icebergDestinationOptions}
										placeholder="Select a destination"
										disabled={isEditMode}
									/>
								</div>
							)}
						</div>
					)}
				<div className="min-h-[280px]">
					{isEditMode && isDetailsError ? (
						<div className="flex min-h-[280px] flex-col items-center justify-center gap-1 text-center">
							<p className="text-xl font-medium leading-7 text-olake-heading-strong">
								Failed to load catalog details
							</p>
							<p className="text-sm leading-[22px] text-olake-body">
								Unable to fetch the catalog configuration. Please try again.
							</p>
							<Button
								type="primary"
								className="mt-3"
								onClick={() => refetchDetails()}
							>
								Retry
							</Button>
						</div>
					) : isLoading ? (
						<div className="flex min-h-[280px] items-center justify-center">
							<Spin tip="Loading schema..." />
						</div>
					) : (
						schema && (
							<div className="py-6">
								<Form
									ref={formRef}
									schema={schema}
									transformErrors={transformErrors}
									templates={{
										ObjectFieldTemplate,
										FieldTemplate: CustomFieldTemplate,
										ArrayFieldTemplate,
										ButtonTemplates: {
											SubmitButton: () => null,
										},
									}}
									widgets={widgets}
									formData={formData}
									onChange={e => {
										const trimmedData = trimFormDataStrings(e.formData)
										setFormData(trimmedData)
									}}
									uiSchema={uiSchema}
									validator={validator}
									showErrorList={false}
									omitExtraData
									liveOmit
								/>
							</div>
						)
					)}
				</div>

				<div className="flex items-center pt-5">
					<div className="flex gap-3">
						<Button
							type="primary"
							onClick={handleConnect}
							disabled={!canSubmit}
						>
							{isEditMode ? "Save Changes" : "Connect"}
						</Button>
						<Button onClick={handleCancel}>Cancel</Button>
					</div>
				</div>
			</Modal>

			<CatalogSuccessModal
				open={open && activeModal === ActiveCatalogModalState.CATALOG_SUCCESS}
				isEditMode={isEditMode}
				onClose={handleViewCatalogs}
				onViewCatalogs={handleViewCatalogs}
				onViewTables={() => {
					setActiveModal(null)
					onClose()
					navigate(
						`/maintenance/tables?catalog=${encodeURIComponent(createdCatalogName)}`,
					)
				}}
			/>

			<TestConnectionModal
				open={open && activeModal === ActiveCatalogModalState.TESTING}
				connectionType="catalog"
			/>
			<TestConnectionSuccessModal
				open={open && activeModal === ActiveCatalogModalState.TEST_SUCCESS}
				connectionType="catalog"
			/>
			<TestConnectionFailureModal
				open={open && activeModal === ActiveCatalogModalState.TEST_FAILURE}
				onClose={handleCancel}
				onEdit={() => setActiveModal(null)}
				connectionType="catalog"
				testConnectionError={testConnectionError}
			/>
			<ErrorLogsModal
				open={open && activeModal === ActiveCatalogModalState.SPEC_FAILED}
				onClose={handleCancel}
				title="Catalog Spec Load Failed"
				error={specError ?? ""}
				onAction={() => {
					setActiveModal(null)
					refetchSpec()
				}}
				actionButtonText="Try Again"
			/>
			<ErrorLogsModal
				open={open && activeModal === ActiveCatalogModalState.CREATION_FAILED}
				onClose={handleCancel}
				title={
					isEditMode ? "Failed to Update Catalog" : "Failed to Create Catalog"
				}
				error={creationError ?? ""}
				onAction={() => {
					setActiveModal(null)
					handleConnect()
				}}
				actionButtonText="Edit Catalog"
			/>
		</>
	)
}

export default CatalogModal
