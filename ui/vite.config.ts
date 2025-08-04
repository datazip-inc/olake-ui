import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import path from "path"

// https://vite.dev/config/
export default defineConfig({
	plugins: [react()],
	resolve: {
		alias: {
			"@": path.resolve(__dirname, "src"),
		},
	},
	server: {
		port: 5173,
		proxy: {
			"/api": {
				target: "http://localhost:8000",
				changeOrigin: true,
				secure: false,
				rewrite: path => path,
			},
			"/login": {
				target: "http://localhost:8000",
				changeOrigin: true,
				secure: false,
			},
			"/signup": {
				target: "http://localhost:8000",
				changeOrigin: true,
				secure: false,
			},
			"/auth": {
				target: "http://localhost:8000",
				changeOrigin: true,
				secure: false,
			},
		},
	},
	// Handle client-side routing
	appType: "spa",
})
