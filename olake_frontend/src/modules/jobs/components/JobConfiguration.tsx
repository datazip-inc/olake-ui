import { Input, Radio, Select, Switch } from "antd"

interface JobConfigurationProps {
	jobName: string
	setJobName: React.Dispatch<React.SetStateAction<string>>
	replicationFrequency: string
	setReplicationFrequency: React.Dispatch<React.SetStateAction<string>>
	schemaChangeStrategy: string
	setSchemaChangeStrategy: React.Dispatch<React.SetStateAction<string>>
	notifyOnSchemaChanges: boolean
	setNotifyOnSchemaChanges: React.Dispatch<React.SetStateAction<boolean>>
	stepNumber?: number | string
	stepTitle?: string
}

const JobConfiguration: React.FC<JobConfigurationProps> = ({
	jobName,
	setJobName,
	replicationFrequency,
	setReplicationFrequency,
	schemaChangeStrategy,
	setSchemaChangeStrategy,
	notifyOnSchemaChanges,
	setNotifyOnSchemaChanges,
	stepNumber = 4,
	stepTitle = "Job Configuration",
}) => {
	return (
		<div className="w-full p-6">
			{stepNumber && stepTitle && (
				<div className="mb-4 flex flex-col gap-2">
					<div className="flex items-center gap-2">
						<div className="size-2 rounded-full border border-[#203FDD] outline outline-2 outline-[#203FDD]"></div>
						<span>Step {stepNumber}</span>
					</div>
					<h1 className="text-xl font-medium">{stepTitle}</h1>
				</div>
			)}

			<div className="rounded-xl border border-[#D9D9D9] p-4">
				<div className="mb-6 grid grid-cols-2 gap-6">
					<div>
						<label className="mb-2 block text-sm font-medium">Job name:</label>
						<Input
							placeholder="Enter your job name"
							value={jobName}
							onChange={e => setJobName(e.target.value)}
						/>
					</div>
					<div>
						<label className="mb-2 block text-sm font-medium">
							Replication frequency:
						</label>
						<Select
							className="w-full"
							value={replicationFrequency}
							onChange={setReplicationFrequency}
						>
							<Select.Option value="hourly">Hourly</Select.Option>
							<Select.Option value="daily">Daily</Select.Option>
							<Select.Option value="weekly">Weekly</Select.Option>
							<Select.Option value="monthly">Monthly</Select.Option>
						</Select>
					</div>
				</div>
			</div>

			<div className="mb-6 mt-6">
				<label className="mb-2 block text-sm font-medium">
					When the source schema changes, I want to:
				</label>
				<div className="rounded-xl border border-gray-200 p-4">
					<Radio.Group
						value={schemaChangeStrategy}
						onChange={e => setSchemaChangeStrategy(e.target.value)}
						className="w-full"
					>
						<div className="mb-2">
							<Radio
								value="propagate"
								className="mb-2 flex w-full items-start"
							>
								<span className="font-medium">
									Propagate field changes only
								</span>
								<p className="mt-1 pl-2 text-sm">
									Only column changes will be propagated. Incompatible schema
									changes will be detected, but not propagated.
								</p>
							</Radio>
						</div>
					</Radio.Group>
				</div>
			</div>

			<div className="flex items-center justify-between rounded-xl border border-gray-200 p-6">
				<span className="font-medium">
					Be notified when schema changes occur
				</span>
				<Switch
					checked={notifyOnSchemaChanges}
					onChange={setNotifyOnSchemaChanges}
					className={notifyOnSchemaChanges ? "bg-blue-600" : ""}
				/>
			</div>
		</div>
	)
}

export default JobConfiguration
