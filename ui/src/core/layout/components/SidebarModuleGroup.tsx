import { CaretLeftIcon } from "@phosphor-icons/react"
import clsx from "clsx"

import { Tag } from "@/common/components"

import type { NavModule } from "../nav-config"
import SidebarNavItem from "./SidebarNavItem"

interface GroupedNavModule extends NavModule {
	moduleLabel: string
	icon: React.ElementType
}

interface SidebarModuleGroupProps {
	mod: GroupedNavModule
	isOpen: boolean
	onToggle: () => void
}

const SidebarModuleGroup: React.FC<SidebarModuleGroupProps> = ({
	mod,
	isOpen,
	onToggle,
}) => {
	const ModIcon = mod.icon
	const HeaderAction = mod.headerAction

	return (
		<div>
			{HeaderAction ? (
				<div className="mb-2 flex h-8 w-full items-center gap-1 rounded-md px-2 text-[14px] leading-[22px] text-[#595959] hover:bg-[#f5f5f5]">
					<button
						type="button"
						onClick={onToggle}
						aria-expanded={isOpen}
						className="flex min-w-0 flex-1 items-center gap-[9px]"
					>
						<ModIcon
							size={16}
							className={mod.iconClassName}
						/>
						<span className="truncate">{mod.moduleLabel}</span>
						{mod.badge && <Tag>{mod.badge}</Tag>}
					</button>
					<HeaderAction />
				</div>
			) : (
				<button
					onClick={onToggle}
					aria-expanded={isOpen}
					className="mb-2 flex h-8 w-full items-center justify-between rounded-md px-2 text-[14px] leading-[22px] text-[#595959] hover:bg-[#f5f5f5]"
				>
					<div className="flex min-w-0 items-center gap-[9px]">
						<ModIcon
							size={16}
							className={mod.iconClassName}
						/>
						<span className="truncate">{mod.moduleLabel}</span>
						{mod.badge && <Tag>{mod.badge}</Tag>}
					</div>
					<CaretLeftIcon
						className={clsx(
							"transition-transform duration-200",
							isOpen ? "-rotate-90" : "-rotate-180",
						)}
						size={14}
					/>
				</button>
			)}

			<div
				className={clsx(
					"grid transition-all duration-200 ease-in-out",
					isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
				)}
			>
				<div className="overflow-hidden">
					<div className="py-2 pl-1">
						<div className="flex gap-2">
							<div className="mt-1 w-px shrink-0 bg-[#d9d9d9]" />
							<div className="min-w-0 flex-1 pl-3">
								{mod.items.map(({ path, label, icon }) => (
									<SidebarNavItem
										key={path}
										path={path}
										label={label}
										icon={icon}
										iconSize={14}
										className="mb-3 h-6 px-1"
									/>
								))}
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}

export default SidebarModuleGroup
