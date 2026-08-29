import { SignOutIcon } from "@phosphor-icons/react"

const SidebarLogout: React.FC<{ onLogout: () => void }> = ({ onLogout }) => (
	<button
		onClick={onLogout}
		className="flex h-8 w-full items-center gap-[9px] rounded-md px-2 text-[14px] leading-[22px] text-olake-body hover:bg-olake-surface-muted"
	>
		<SignOutIcon size={16} />
		<span>Logout</span>
	</button>
)

export default SidebarLogout
