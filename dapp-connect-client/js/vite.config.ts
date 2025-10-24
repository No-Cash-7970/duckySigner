import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],
  test: {
		include: ['src/**/*.{test,spec}.{js,ts}'],
	},
})
