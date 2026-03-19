import { CaretLeftIcon } from "@phosphor-icons/react"
import clsx from "clsx"

import type { NavModule } from "../nav-config"
import SidebarNavItem from "./SidebarNavItem"

const SidebarModuleGroup: React.FC<{
	mod: NavModule
	isOpen: boolean
	onToggle: () => void
}> = ({ mod, isOpen, onToggle }) => {
	const ModIcon = mod.icon

	return (
		<div>
			<button
				onClick={onToggle}
				aria-expanded={isOpen}
				className="mb-2 flex h-8 w-full items-center justify-between rounded-md px-2 text-[14px] leading-[22px] text-[#595959] hover:bg-[#f5f5f5]"
			>
				<div className="flex items-center gap-[9px]">
					<ModIcon
						size={16}
						className={mod.iconClassName}
					/>
					<span>{mod.moduleLabel}</span>
					{mod.badge && (
						<span className="rounded-full bg-[#f3f5fd] px-2 text-xs text-[#193AE6]">
							{mod.badge}
						</span>
					)}
				</div>
				<CaretLeftIcon
					className={clsx(
						"transition-transform duration-200",
						isOpen ? "-rotate-90" : "-rotate-180",
					)}
					size={14}
				/>
			</button>

			<div
				className={clsx(
					"grid transition-all duration-200 ease-in-out",
					isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
				)}
			>
				<div className="overflow-hidden">
					<div className="px-4 py-2">
						<div className="flex gap-5">
							<div className="mt-1 w-px bg-[#d9d9d9]" />
							<div className="w-full pl-3">
								{mod.items.map(({ path, label, icon }) => (
									<SidebarNavItem
										key={path}
										path={path}
										label={label}
										icon={icon}
										iconSize={14}
										className="mb-3 h-6"
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
