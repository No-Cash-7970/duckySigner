# Ducky Signer Prototype

Prototype for a simple desktop Algorand wallet that signs things…maybe.

> [!WARNING]
> This is an experimental prototype. Do not put keys for accounts that you care about into this wallet. Use it for testing purposes only.

## Installation for development

This project uses [Wails v3-alpha](https://v3alpha.wails.io/). The Wails v3 source code is expected to be in the same parent directory as the project directory with the name `wails`.

```text
parent_dir/
├─ duckySigner/
├─ wails/
…
```

First, install Wails:

```bash
git clone https://github.com/wailsapp/wails.git
cd wails
git checkout v3-alpha
cd v3/cmd/wails3
go install
```

Then move back into the parent directory:

```bash
cd ..
```

Lastly, install the project:

```bash
git clone https://github.com/No-Cash-7970/duckySigner.git
cd duckySigner
go mod download
```

## Upgrading backend dependencies

The easy way to update the project's Go dependencies is to use the [go-mod-upgrade](https://github.com/oligot/go-mod-upgrade) tool. If go-mod-upgrade is not installed yet, install it:

```bash
go install github.com/oligot/go-mod-upgrade@latest
```

Then run go-mod-upgrade to interactively select packages to upgrade:

```bash
go-mod-upgrade
```

Alternatively, upgrade all packages without go-mod-upgrade:

```bash
go get -u
go mod tidy
```

### Keeping Wails up to date

Because this project is currently using the alpha version of Wails v3, Wails needs to be updated constantly. Run the following to update Wails:

```bash
cd wails
git checkout v3-alpha
cd v3/cmd/wails3
go install
```

## Upgrading frontend dependencies

The frontend is somewhat separate from the backend. It is a TypeScript/Javascript sub-project that uses Node.js and Yarn. Upgrade the frontend dependencies by navigating to the `frontend` directory and using yarn to upgrade:

```bash
cd frontend
yarn upgrade-interactive
```

Refer to the installation instructions of the Wails v3 documentation for more information: <https://v3alpha.wails.io/getting-started/installation/#installation_1>.

## Building project

Run the following to build the project:

```bash
wails3 build
```

The output is placed in the `bin` directory.

## Dev mode

Wails v3 provides a "dev mode" the watches for changes and automatically rebuilds the project when there is a change. Activate dev mode by running:

```bash
wails3 dev
```

Use <kbd>Ctrl</kbd>+<kbd>C</kbd> (<kbd>Cmd</kbd>+<kbd>C</kbd> on Mac) to exit dev mode.

## Testing

Unit tests are always in the same directory as the module or class that is being tested.

To run all tests once:

```bash
wails3 task test:all
```

To run only backend tests once:

```bash
wails3 task test
```

To run backend tests in watch mode

```bash
wails3 task test:watch
```

To run frontend tests once:

```bash
wails3 task test:frontend
```

Or run the frontend tests using `yarn`:

```bash
cd frontend
yarn test
```

## Other useful tasks

Look in `Taskfile.yml` for a complete list of tasks available for this project that can be run using `wails3 task TASK_NAME`.

## File structure

This project largely follows the file structure of a [Wails v3-alpha](https://v3alpha.wails.io/) project.

- The frontend UI code is in the `frontend` directory.
- The modified KMD code is in the `kmd` directory.
- Wails services used connect the frontend actions to the backend are in the `services` directory.
- Other miscellaneous backend code (if any) is put in the `internal` directory.
