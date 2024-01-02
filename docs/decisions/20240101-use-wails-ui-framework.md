# Use Wails UI Framework

- Status: draft
- Deciders: No-Cash-7970
- Date: 2024-01-01

## Context and Problem Statement

What framework(s) should be used to build the desktop wallet app's graphical user interface (GUI)?

## Decision Drivers

- **Ease of use:** Prefer a framework that has a short learning curve and is flexible enough to accommodate fluctuating requirements
- **Appearance:** Must be a framework that easily allow to make a modern and good-looking UI
- **Feature set:** Prefer a framework with a wide array of UI element to choose from
- **Documentation and help:** Prefer a framework with clear documentation and an active community that can help with questions
- **Programming Language(s):** Must be a framework that can be written in one or more the chosen languages (Refer to: [Build Using Go and TypeScript](20240101-build-using-go-and-typescript.md)). Prefer a framework in a single language, but a framework that is divided into a backend language and a UI language is acceptable as long it allows the developer (No-Cash-7970) to use her past development experience.
- **Cross-platform:** Must be able to be compiled on Linux, Mac and Windows
- **Free and open source:** Must be a framework where the code can be examined and the license allows for commercial use

## Considered Options

- Wails
- Gio UI
- Fyne
- Electron
- Tauri

## Decision Outcome

Chose Wails because it can utilize Go's performance and cross-platform ability with the flexibility and massive number of TypeScript and JavaScript libraries and tools. Despite the less-than-stellar documentation, the community in the Wails Discord server should be able to assist with issues that may arise. In addition to Wails, Next.js will be used as the TypeScript/JavaScript because the developer is familiar with the framework.

**Certainty**: Medium. It is not clear if using Wails will help produce the desired results within a reasonable amount of time.

## Pros and Cons of the Options

### Wails

Wails is a desktop app framework for building with Go and web technologies (HTML, Javascript/TypeScript, CSS). According to its documentation it is a "lightweight and fast Electron alternative for Go."

Website: <https://wails.io/>

- Pro: The developer can utilize her knowledge in frontend web development to build a good-looking modern UI
- Pro: Utilizes Go, which offers great cross-platform support
- Pro: There is an active Discord server with helpful people who can answer questions
- Pro: Free and open source under MIT license
- Con: The documentation is messy and many things are not explained clearly.
- Con: Many of the tutorials available are outdated and not very useful
- Con: Some useful features, such as notifications, are not yet supported.

### Gio UI

Gio is a cross-platform UI library for Go. It is the most popular option for desktop apps written in Go.

Website: <https://gioui.org/>

- Pro: Supports a large number of platforms, including WebAssembly
- Pro: Free and open source under MIT and Unlicense licenses
- Pro: Well documented with plenty of tutorials
- Pro: UI elements are good-looking and modern
- Con: Has a larger learning curve because of its architecture (e.g. "immediate mode", “retained mode”)
- Con: Seems somewhat inflexible

### Fyne

Fyne is a toolkit for developing native desktop apps in Go. It is one of the more popular options for developing desktop apps in Go.

Website: <https://fyne.io/>

- Pro: The learning curve is moderate
- Pro: UI elements are good-looking and modern
- Pro: Well documented
- Pro: Free and open source under BSD 3-Clause License
- Pro: Cross-platform, including support for mobile
- Con: The appearance of the UI elements seems to be inflexible, which results in most of the apps using Fyne looking similar

### Electron

Electron is a massively popular framework for building desktop apps using web technologies.

Website: <https://www.electronjs.org/>

- Pro: The developer can utilize her knowledge in frontend web development to build a good-looking modern UI
- Pro: Plenty of documentation that is clear and well-organized, as well as plenty of tutorials
- Pro: Designed to be cross-platform
- Pro: Free and open source under MIT license
- Con: Many complaints about apps built with Electron consuming a lot of computing resources.

### Tauri

Tauri is a framework for building performant and secure desktop apps with Rust and web technologies

Website: <https://tauri.app/>

- Pro: The developer can utilize her knowledge in frontend web development to build a good-looking modern UI
- Pro: Has great cross-platform support
- Pro: Designed to be cross-platform
- Pro: Free and open source under MIT license
- Pro: Emphasis on security, which is important for a desktop wallet
- Pro: Plenty of documentation that is clear and well-organized
- Con: Uses Rust as the backend. Rust has a steep learning curve that would perhaps require months to overcome.

## Links

- Relates to [Build Using Go and TypeScript](20240101-build-using-go-and-typescript.md)
- [Answer to "Comparison with wails" GitHub discussion](https://github.com/tauri-apps/tauri/discussions/3521#discussioncomment-3472966)
- [Comparison of web-to-desktop frameworks](https://github.com/Elanis/web-to-desktop-framework-comparison)
