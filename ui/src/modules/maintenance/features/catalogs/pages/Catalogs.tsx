import {
	DotsThreeIcon,
	MagnifyingGlassIcon,
	PencilSimpleIcon,
	PlusIcon,
	TrashIcon,
} from "@phosphor-icons/react"
import { Button, Dropdown, Input } from "antd"
import type { MenuProps } from "antd/es/menu"
import { useState } from "react"

import { DataTable, PageErrorState, Tag } from "@/common/components"
import type { ColumnDef } from "@/common/components"
import { usePaginatedSearch } from "@/common/hooks"

import { CatalogModal } from "../components"
import { useCatalogs, useDeleteCatalog } from "../hooks"
import type { Catalog } from "../types"

const catalogSearchFn = (row: Catalog, term: string): boolean =>
	row.name.toLowerCase().includes(term)

const Catalogs: React.FC = () => {
	const [openActionRow, setOpenActionRow] = useState<string | null>(null)
	const [modalOpen, setModalOpen] = useState(false)
	const [activeCatalogName, setActiveCatalogName] = useState<
		string | undefined
	>(undefined)
	const deleteCatalogMutation = useDeleteCatalog()

	const closeModal = () => {
		setModalOpen(false)
		setActiveCatalogName(undefined)
	}
	const { data: catalogRows = [], isLoading, isError, refetch } = useCatalogs()

	const {
		searchTerm,
		setSearchTerm,
		currentPage,
		setCurrentPage,
		paginatedRows,
		totalPages,
	} = usePaginatedSearch<Catalog>({
		rows: catalogRows,
		searchFn: catalogSearchFn,
	})

	const getMenuItems = (row: Catalog): MenuProps["items"] => [
		{
			key: `edit-${row.id}`,
			icon: <PencilSimpleIcon size={16} />,
			label: <span className="text-sm leading-[22px]">Edit Catalog</span>,
			onClick: () => {
				setOpenActionRow(null)
				setActiveCatalogName(row.name)
				setModalOpen(true)
			},
		},
		{
			key: `delete-${row.id}`,
			onClick: () => {
				setOpenActionRow(null)
				deleteCatalogMutation.mutate(row.name)
			},
			icon: (
				<TrashIcon
					size={16}
					className="text-olake-error"
				/>
			),
			label: (
				<span className="text-sm leading-[22px] text-olake-error">
					Delete Catalog
				</span>
			),
		},
	]

	const columns: ColumnDef<Catalog>[] = [
		{
			key: "actions",
			header: "Actions",
			width: 9,
			render: row => (
				<Dropdown
					menu={{ items: getMenuItems(row) }}
					trigger={["click"]}
					open={openActionRow === row.id}
					onOpenChange={isOpen => setOpenActionRow(isOpen ? row.id : null)}
				>
					<Button className="size-8 border-0 p-0">
						<DotsThreeIcon size={16} />
					</Button>
				</Dropdown>
			),
		},
		{
			key: "name",
			header: "Catalog Name",
			width: 60,
			render: row => (
				<div className="flex items-center gap-2">
					<p className="text-sm leading-6 text-olake-text">{row.name}</p>
					{row.olakeCreated && <Tag>OLake</Tag>}
				</div>
			),
		},
		{
			key: "type",
			header: "Type",
			width: 18,
			render: row => (
				<span className="text-sm leading-6 text-olake-text">{row.type}</span>
			),
		},
		{
			key: "createdOn",
			header: "Created on",
			width: 10,
			render: row => (
				<span className="text-sm leading-6 text-olake-text">
					{row.createdOn}
				</span>
			),
		},
	]

	return (
		<div className="min-h-full bg-white px-6 pt-6">
			<div className="w-full">
				<div>
					<h1 className="text-xl font-medium leading-7 text-olake-text">
						Catalogs
					</h1>
					<p className="mt-1 text-sm leading-[22px] text-olake-text">
						Add, Import, or Edit catalogs
					</p>
				</div>

				<div className="mt-6 flex h-9 items-center gap-6">
					<div className="flex h-9 w-[479px] overflow-hidden rounded-md border border-olake-border">
						<Input
							value={searchTerm}
							onChange={e => setSearchTerm(e.target.value)}
							placeholder="Search Catalogs"
						/>
						<Button
							type="text"
							className="h-9 w-8 rounded-none border-l border-olake-border p-0"
							icon={<MagnifyingGlassIcon size={16} />}
						/>
					</div>

					<Button
						type="primary"
						icon={<PlusIcon size={16} />}
						onClick={() => setModalOpen(true)}
					>
						New Catalog
					</Button>
				</div>
			</div>

			<div className="mt-6 w-full">
				{isError ? (
					<PageErrorState
						title="Failed to load catalogs"
						description="Please check your connection and try again."
						onRetry={() => {
							refetch()
						}}
					/>
				) : (
					<DataTable
						columns={columns}
						rows={paginatedRows}
						rowKey={row => row.id}
						loading={isLoading}
						pagination={
							totalPages > 1
								? {
										currentPage,
										totalPages,
										onPageChange: setCurrentPage,
									}
								: undefined
						}
						emptyStateConfig={{
							title: "No Catalogs Available",
							subtitle:
								"Create a new catalog to view tables for optimizations.",
							onRefetch: () => setModalOpen(true),
							refetchLabel: "New Catalog",
						}}
					/>
				)}
			</div>

			<CatalogModal
				open={modalOpen}
				catalogName={activeCatalogName}
				onClose={closeModal}
				onSuccess={closeModal}
			/>
		</div>
	)
}

export default Catalogs
