import { FolderDashedIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"

type CatalogNotAvailableModalProps = {
	open: boolean
	onClose: () => void
	catalogName: string
}

const CatalogNotAvailableModal: React.FC<CatalogNotAvailableModalProps> = ({
	open,
	onClose,
	catalogName,
}) => {
	return (
		<Modal
			open={open}
			onCancel={onClose}
			footer={null}
			closable={false}
			centered
			width={632}
			destroyOnHidden
			styles={{
				content: {
					padding: 0,
					overflow: "hidden",
					borderRadius: 20,
				},
				body: {
					padding: 0,
				},
			}}
		>
			<div className="flex h-[360px] flex-col items-center bg-white pt-20">
				<div className="flex w-[520px] max-w-full flex-col items-center gap-[14px] px-6 text-center">
					<FolderDashedIcon
						size={32}
						weight="regular"
						className="text-olake-icon-muted"
					/>

					<div className="flex w-full flex-col items-center gap-1">
						<p className="text-xl font-medium leading-7 text-olake-text">
							Catalog not available
						</p>
						<p className="w-full whitespace-nowrap text-sm leading-[22px] text-olake-text-secondary">
							Selected catalog is missing, please check your source
						</p>
					</div>
				</div>

				<div className="mt-[18px] flex w-fit items-center justify-center gap-2 rounded-md border border-dashed border-olake-border px-2 py-1">
					<FolderDashedIcon
						size={20}
						weight="regular"
						className="text-olake-icon-muted"
					/>
					<span className="whitespace-nowrap text-sm leading-[22px] text-olake-text-secondary">
						{catalogName}
					</span>
				</div>

				<div className="mt-[49px]">
					<Button onClick={onClose}>Understood</Button>
				</div>
			</div>
		</Modal>
	)
}

export default CatalogNotAvailableModal
