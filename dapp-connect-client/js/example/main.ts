import './style.css'
import typescriptLogo from './typescript.svg'
import viteLogo from '/vite.svg'
import { setupCounter } from '../example/counter.ts'

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div class="max-w-7xl p-8 text-center self-center">
    <a href="https://vite.dev" target="_blank">
      <img src="${viteLogo}" class="h-24 p-6 will" alt="Vite logo" />
    </a>
    <a href="https://www.typescriptlang.org/" target="_blank">
      <img src="${typescriptLogo}" class="h-24 p-6 will" alt="TypeScript logo" />
    </a>
    <h1>Vite + TypeScript</h1>
    <div class="p-8">
      <button id="counter" type="button"></button>
    </div>
    <p class="text-gray-500">
      Click on the Vite and TypeScript logos to learn more
    </p>
  </div>
`

setupCounter(document.querySelector<HTMLButtonElement>('#counter')!)
