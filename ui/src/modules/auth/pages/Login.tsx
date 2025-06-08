import { useState } from "react"
import { Form, Input, Button, Card, Alert } from "antd"
import { User, LockKey } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { useNavigate } from "react-router-dom"
import { LoginArgs } from "../../../types"

const Login: React.FC = () => {
	const { login } = useAppStore()
	const navigate = useNavigate()
	const [loading, setLoading] = useState(false)
	const [error, setError] = useState<string | null>(null)

	const onFinish = async (values: LoginArgs) => {
		setLoading(true)
		setError(null)
		try {
			await login(values.username, values.password)
			navigate("/jobs")
		} catch (err) {
			const errorMessage =
				err instanceof Error
					? err.message
					: "Failed to login. Please try again."
			setError(errorMessage)
		} finally {
			setLoading(false)
		}
	}

	return (
		<div className="flex h-screen items-center justify-center bg-gray-100">
			<Card className="w-full max-w-md shadow-md">
				<div className="mb-6 flex flex-col gap-2 text-center">
					<div className="text-3xl font-bold">Login</div>
					<div className="text-gray-400">
						Please enter your credentials to continue
					</div>
				</div>

				{error && (
					<Alert
						message="Login Failed"
						description={error}
						type="error"
						showIcon
						className="mb-4"
					/>
				)}

				<Form
					name="login"
					initialValues={{ remember: true }}
					onFinish={onFinish}
					layout="vertical"
				>
					<Form.Item
						name="username"
						rules={[
							{ required: true, message: "Please input your username!" },
							{ min: 3, message: "Username must be at least 3 characters" },
						]}
					>
						<Input
							prefix={
								<User
									className="site-form-item-icon text-gray-400"
									weight="bold"
									size={18}
								/>
							}
							placeholder="Username"
							size="large"
						/>
					</Form.Item>

					<Form.Item
						name="password"
						rules={[
							{ required: true, message: "Please input your password!" },
							{ min: 6, message: "Password must be at least 6 characters" },
						]}
					>
						<Input.Password
							prefix={
								<LockKey
									className="site-form-item-icon text-gray-400"
									weight="bold"
									size={18}
								/>
							}
							placeholder="Password"
							size="large"
						/>
					</Form.Item>

					<Form.Item>
						<Button
							type="primary"
							htmlType="submit"
							loading={loading}
							className="w-full"
							size="large"
						>
							Log in
						</Button>
					</Form.Item>
				</Form>
			</Card>
		</div>
	)
}

export default Login
