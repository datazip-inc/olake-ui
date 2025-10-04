import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Form, Input, Button, Card, message } from "antd"
import { UserIcon, LockKeyIcon } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import { LoginArgs } from "../../../types"

const Login: React.FC = () => {
	const { login } = useAppStore()
	const navigate = useNavigate()
	const [loading, setLoading] = useState(false)
	const [form] = Form.useForm()

	const onFinish = async (values: LoginArgs) => {
		try {
			setLoading(true)
			await login(values.username, values.password)
			navigate("/jobs")
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "An unexpected error occurred"
			message.error({
				content: errorMessage,
				duration: 3,
				className: "font-medium",
			})
			form.resetFields()
		}
		setLoading(false)
	}

	return (
		<div className="flex h-screen items-center justify-center bg-gray-100">
			<Card className="w-full max-w-md shadow-md">
				<div className="mb-6 flex flex-col gap-2 text-center">
					<div className="text-3xl font-bold">Login</div>
					<div className="text-[#9ca3af]">
						Please enter your credentials to continue
					</div>
				</div>

				<Form
					form={form}
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
								<UserIcon
									className="site-form-item-icon text-gray-500"
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
								<LockKeyIcon
									className="site-form-item-icon text-gray-500"
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
