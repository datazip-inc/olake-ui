import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"
import { Button, message, Modal, Spin } from "antd"
import { useEffect, useRef, useState } from "react"
import { useNavigate } from "react-router-dom"

import ArrayFieldTemplate from "@/common/components/form/ArrayFieldTemplate"
import CustomFieldTemplate from "@/common/components/form/CustomFieldTemplate"
import ObjectFieldTemplate from "@/common/components/form/ObjectFieldTemplate"
import { widgets } from "@/common/components/form/widgets"
import {
	SpecFailedModal,
	TestConnectionFailureModal,
	TestConnectionModal,
	TestConnectionSuccessModal,
} from "@/common/components/modals"
import { transformErrors, TEST_CONNECTION_STATUS } from "@/common/constants"
import { TestConnectionError } from "@/common/types"
import { trimFormDataStrings, handleSpecResponse } from "@/common/utils"

import {
	useCatalogDetails,
	useCatalogSpec,
	useCatalogVersions,
	useCreateCatalog,
	useTestCatalogConnection,
	useUpdateCatalog,
} from "../hooks"
import type { CatalogModalProps, CatalogFormData } from "../types"
import CatalogSuccessModal from "./CatalogSuccessModal"

type ActiveModal =
	| null
	| "testing"
	| "testSuccess"
	| "pendingCatalogSave"
	| "testFailure"
	| "specFailed"
	| "catalogSuccess"

const getCatalogNameFromFormData = (data: CatalogFormData): string => {
	const { catalog_name } = (data as { writer: { catalog_name: string } }).writer
	return catalog_name.trim()
}

/** API expects the writer object only, not `{ type, writer }`. */
const getCatalogWriterPayload = (
	data: CatalogFormData,
): Record<string, unknown> => {
	const writer = (data as { writer?: Record<string, unknown> }).writer
	if (!writer || typeof writer !== "object") {
		throw new Error("Missing catalog writer configuration")
	}
	return writer
}

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
	const [activeModal, setActiveModal] = useState<ActiveModal>(null)
	const [specError, setSpecError] = useState<string | null>(null)
	const [testConnectionError, setTestConnectionError] =
		useState<TestConnectionError | null>(null)
	const [createdCatalogName, setCreatedCatalogName] = useState("")

	const navigate = useNavigate()
	const { data: versionsData, isLoading: loadingVersions } =
		useCatalogVersions(open)
	const versions = versionsData?.version ?? []
	const latestVersion = versions[0] ?? ""

	const {
		data: specData,
		isLoading: loadingSpec,
		error: specQueryError,
		refetch: refetchSpec,
	} = useCatalogSpec(latestVersion, open)

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
			setSpecError(errMsg)
			setActiveModal("specFailed")
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
			setActiveModal(null)
		}
	}, [open])

	const validateForm = async (): Promise<boolean> => {
		if (schema && formRef.current) {
			return formRef.current.validateForm()
		}

		return true
	}

	const handleConnect = async () => {
		const isValid = await validateForm()
		if (!isValid) return

		const testCatalogData = {
			version: latestVersion,
			config: JSON.stringify(formData),
		}

		setActiveModal("testing")
		try {
			const testResult = await testCatalogMutation.mutateAsync({
				catalog: testCatalogData,
			})

			if (
				testResult.data?.connection_result.status ===
				TEST_CONNECTION_STATUS.SUCCEEDED
			) {
				setActiveModal("testSuccess")
				setTimeout(async () => {
					setActiveModal("pendingCatalogSave")
					if (isEditMode) {
						await updateCatalogMutation.mutateAsync({
							catalogName: catalogName!,
							config: getCatalogWriterPayload(formData) as CatalogFormData,
						})
						setCreatedCatalogName(catalogName!)
						setActiveModal("catalogSuccess")
					} else {
						await createCatalogMutation.mutateAsync(
							getCatalogWriterPayload(formData) as CatalogFormData,
						)
						setCreatedCatalogName(getCatalogNameFromFormData(formData))
						setActiveModal("catalogSuccess")
					}
				}, 1000)
			} else {
				setTestConnectionError({
					message: testResult.data?.connection_result.message || "",
					logs: testResult.data?.logs || [],
				})
				setActiveModal("testFailure")
			}
		} catch {
			setActiveModal(null)
			message.error("Connection test failed. Please try again.")
		}
	}

	const handleViewCatalogs = () => {
		setActiveModal(null)
		onClose()
		onSuccess?.()
	}

	const handleCancel = () => {
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
				open={open && activeModal === "catalogSuccess"}
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
				open={open && activeModal === "testing"}
				connectionType="catalog"
			/>
			<TestConnectionSuccessModal
				open={open && activeModal === "testSuccess"}
				connectionType="catalog"
			/>
			<TestConnectionFailureModal
				open={open && activeModal === "testFailure"}
				onClose={handleCancel}
				onEdit={() => setActiveModal(null)}
				connectionType="catalog"
				testConnectionError={testConnectionError}
			/>
			<SpecFailedModal
				open={open && activeModal === "specFailed"}
				onClose={handleCancel}
				connectionType="catalog"
				error={specError ?? ""}
				onTryAgain={refetchSpec}
			/>
		</>
	)
}

export default CatalogModal
