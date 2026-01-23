// Some parts taken from:
// https://daisyui.com/blog/how-to-install-sveltekit-and-daisyui/

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,svelte,js,ts}'],
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
  daisyui: {
    darkTheme: 'dark',
    themes: [
      "light",
      "dark",
    ],
  },
};
