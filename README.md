# PreFlight 🚀

A CLI tool to streamline project setup and ensure all dependencies are in check.

> ⚠️ **Disclaimer:** PreFlight is currently in beta development, meaning that features, functionality, and stability are subject to change frequently. This project is highly experimental, and users should expect regular updates, modifications, and potential disruptions.
>
> We appreciate any **feedback or bug reports** to help improve the project.

---

## 📌 Overview

PreFlight is a powerful command-line tool designed to **validate, check, and fix** your project's environment and dependencies before starting development.

It helps developers ensure that **all required tools and configurations** are set up correctly before running a project.

---

## ✨ Features

#### 🔍 Check Command (`preflight check`)
- Ensures your system matches the project's expected setup.
- Supports **Go, PHP, Composer, Node.js, npm, pnpm, and Yarn**.
- **EOL (End of Life) Detection** for **PHP and Node.js versions**.
- Configurable **timeout** (`--timeout=<seconds>`) for dependency checks.
- Supports **filtering by package manager** using `--pm=composer,php,node`.

#### 🔧 Fix Command (`preflight fix`)
- Automatically **installs missing dependencies**.
- Supports **Composer (PHP)** and package managers (**npm, pnpm, Yarn**).
- Allows **force reinstallation** of dependencies using `--force`.

#### 📋 List Command (`preflight list`)
- Displays all installed dependencies for:
	- **Composer (PHP)**
	- **npm, pnpm, Yarn (JavaScript)**
	- **Go Modules**
- Supports **filtering by package manager** using `--pm=node,composer`.

---

### 🔄 **Dependency Management**
- **Detects missing dependencies** and suggests fixes.
- **Ensures correct versions** of required tools and libraries.
- **Verifies lock files**:
	- `composer.lock`
	- `package-lock.json`
	- `pnpm-lock.yaml`
	- `yarn.lock`

---

### ⚙️ **Customization & Flags**

| Flag                | Description                                                   |
|---------------------|---------------------------------------------------------------|
| `--pm=<managers>`  | Filter by package manager (e.g., `--pm=php,composer,node`).   |
| `--force`          | Force reinstall dependencies (for `fix` command).             |
| `--timeout=<sec>`  | Set timeout for dependency checks (for `check` command).      |

---

## 📌 Requirements

- **Go 1.24** or higher
- **Access to the project's root directory**

## 🚀 Installation

```sh
go install github.com/MineHubs-Studios/PreFlight@latest
```

## 💡 Support

If you encounter any problems or have suggestions, please open an issue.

## 🤝 Contributing
We welcome contributions! Follow these steps:

1. **Fork** the repository
2. **Create** your feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add some amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
