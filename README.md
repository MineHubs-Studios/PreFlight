# PreFlight 🚀

A CLI tool to streamline project setup and ensure all dependencies are in check.

> :warning: **Disclaimer:** PreFlight is currently in beta development,
meaning that features, functionality, and stability are subject to change on a daily basis.
The project is highly experimental, and users should expect frequent updates, modifications, and potential disruptions.

>We appreciate any feedback or bug reports to help improve the project.

## Overview

PreFlight is a powerful command-line tool designed to validate your project's environment and dependencies before takeoff. It helps developers ensure that all necessary dependencies, tools, and configurations are properly set up before running a project.

## Features

- ✅ Dependency version validation
- 🔍 Package manager detection (npm, composer)
- 📦 Lock file verification
- 🛠️ Environment setup checking
- 💻 Development tool validation

## Requirements

- Go 1.24 or higher
- Access to project's root directory

## Installation

```bash
go install github.com/MineHubs-Studios/PreFlight@latest
```

## Usage

Run PreFlight in your project directory:

```bash
preflight check
```

## Support

If you encounter any problems or have suggestions, please open an issue in the GitHub repository.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
