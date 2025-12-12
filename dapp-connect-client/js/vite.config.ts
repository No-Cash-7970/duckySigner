/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import { resolve } from 'path'
import dts from 'unplugin-dts/vite'

export default defineConfig({
  plugins: [
    dts({
      tsconfigPath: 'tsconfig.build.json',
    })
  ],
  // @ts-expect-error
  test: {
		include: ['src/**/*.{test,spec}.{js,ts}'],
	},
  build: {
    lib: {
      entry: resolve(__dirname, 'src/main.ts'),
			name: 'DuckyConnect',
			fileName: 'index',
    },
    rollupOptions: {
      external: [],
    },
  },
})
