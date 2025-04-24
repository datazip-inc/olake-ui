import { Input, Select, Switch } from "antd"
import StepTitle from "../../common/components/StepTitle"
import { useState } from "react"

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
	stepNumber = 4,
	stepTitle = "Job Configuration",
}) => {
	const [replicationFrequencyValue, setReplicationFrequencyValue] =
		useState("1")
	return (
		<div className="w-full p-6">
			{stepNumber && stepTitle && (
				<StepTitle
					stepNumber={stepNumber}
					stepTitle={stepTitle}
				/>
			)}
			<div className="rounded-xl border border-[#D9D9D9] p-4">
				<div className="mb-2 grid grid-cols-2 gap-6">
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
						<div className="flex w-full items-center gap-2">
							<Input
								value={replicationFrequencyValue}
								defaultValue={replicationFrequencyValue}
								onChange={e => setReplicationFrequencyValue(e.target.value)}
								className="w-2/5"
							/>

							<Select
								className="w-3/5"
								value={replicationFrequency}
								onChange={setReplicationFrequency}
							>
								<Select.Option value="seconds">Seconds</Select.Option>
								<Select.Option value="minutes">Minutes</Select.Option>
								<Select.Option value="hours">Hours</Select.Option>
								<Select.Option value="months">Months</Select.Option>
								<Select.Option value="years">Years</Select.Option>
							</Select>
						</div>
					</div>
				</div>
			</div>

			<div className="mb-6 mt-6">
				<label className="mb-2 block text-sm font-medium">
					When the source schema changes, I want to:
				</label>
				<div className="rounded-xl border border-gray-200 p-4">
					<div className="flex items-center justify-between">
						<div>
							<span className="font-medium">Propagate field changes only</span>
							<p className="mt-1 text-sm text-[#575757]">
								Only column changes will be propagated. Incompatible schema
								changes will be detected, but not propagated.
							</p>
						</div>
						<Switch
							checked={schemaChangeStrategy === "propagate"}
							onChange={checked =>
								setSchemaChangeStrategy(checked ? "propagate" : "none")
							}
							className={
								schemaChangeStrategy === "propagate"
									? "bg-blue-600"
									: "bg-gray-300"
							}
						/>
					</div>
				</div>
			</div>
		</div>
	)
}

export default JobConfiguration
