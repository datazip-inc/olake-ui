import { Input, Select } from "antd"
import StepTitle from "../../common/components/StepTitle"
import { getFrequencyValue } from "../../../utils/utils"

interface JobConfigurationProps {
	jobName: string
	setJobName: React.Dispatch<React.SetStateAction<string>>
	replicationFrequency: string
	setReplicationFrequency: React.Dispatch<React.SetStateAction<string>>
	replicationFrequencyValue: string
	setReplicationFrequencyValue: React.Dispatch<React.SetStateAction<string>>
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
	replicationFrequencyValue,
	setReplicationFrequencyValue,
	stepNumber = 4,
	stepTitle = "Job Configuration",
}) => {
	const handleFrequencyChange = (selectedUnit: string) => {
		setReplicationFrequency(getFrequencyValue(selectedUnit))
	}

	const handleValueChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const newValue = e.target.value
		setReplicationFrequencyValue(newValue)
		const unit = getFrequencyValue(replicationFrequency)
		setReplicationFrequency(unit)
	}

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
						<label className="mb-2 block text-sm font-medium">
							Job name:<span className="text-red-500">*</span>
						</label>
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
								onChange={handleValueChange}
								className="w-2/5"
							/>

							<Select
								className="w-3/5"
								value={getFrequencyValue(replicationFrequency)}
								onChange={handleFrequencyChange}
							>
								<Select.Option value="minutes">Minutes</Select.Option>
								<Select.Option value="hours">Hours</Select.Option>
								<Select.Option value="days">Days</Select.Option>
								<Select.Option value="weeks">Weeks</Select.Option>
								<Select.Option value="months">Months</Select.Option>
								<Select.Option value="years">Years</Select.Option>
							</Select>
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}

export default JobConfiguration
