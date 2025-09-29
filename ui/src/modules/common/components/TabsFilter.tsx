import { TabsFilterProps } from "@types/index"

const TabsFilter: React.FC<TabsFilterProps> = ({
	tabs,
	activeTab,
	onChange,
}) => {
	return (
		<div className="border-b border-gray-200">
			<div className="flex">
				{tabs.map(tab => (
					<button
						key={tab.key}
						className={`px-4 py-3 text-sm font-medium ${
							activeTab === tab.key
								? "border-b-2 border-blue-600 text-blue-600"
								: "text-gray-500 hover:text-gray-700"
						}`}
						onClick={() => onChange(tab.key)}
					>
						{tab.label}
					</button>
				))}
			</div>
		</div>
	)
}

export default TabsFilter
