import { Button, Modal } from "antd"

import SuccessIcon from "@/assets/success-icon.svg"

type CatalogAddedSuccessModalProps = {
	open: boolean
	onClose: () => void
	onViewTables: () => void
	onViewCatalogs: () => void
}

const CatalogAddedSuccessModal: React.FC<CatalogAddedSuccessModalProps> = ({
	open,
	onClose,
	onViewTables,
	onViewCatalogs,
}) => {
	return (
		<Modal
			open={open}
			onCancel={onClose}
			title={null}
			footer={null}
			width={680}
			centered
			destroyOnHidden
			closable={false}
		>
			<div className="flex h-[620px] flex-col items-center justify-center">
				<img
					src={SuccessIcon}
					alt=""
					aria-hidden
					className="mb-6 size-16"
				/>

				<div className="mb-5 flex w-64 flex-col items-center gap-1 text-center">
					<p className="text-xl font-medium leading-7 text-olake-text">
						Catalog Added Successfully
					</p>
					<p className="text-sm leading-[22px] text-olake-text-secondary">
						You can start optimising tables
					</p>
				</div>

				<div className="flex items-center gap-2">
					<Button onClick={onViewTables}>View Tables</Button>
					<Button onClick={onViewCatalogs}>View Catalogs</Button>
				</div>
			</div>
		</Modal>
	)
}

export default CatalogAddedSuccessModal
