# tssh - Convenient CLI client for teleport

`tssh` is a wrapper around the Teleport CLI client, designed to simplify connections to servers by enabling autocompletion and fuzzy searching. This project aims to streamline the process of connecting to servers via the command line, making it more intuitive and user-friendly.

![tssh](./preview.gif)

## Installation

### Prerequisites

- [Teleport CLI](https://goteleport.com/docs/installation/)
- [Golang](https://go.dev/dl/)

### Clone the Repository

```sh
git clone https://github.com/Firebain/tssh
cd tssh
```

### Install Binary

```sh
go install
```

## Usage

### Pre-requisites

`tssh` uses the same authentication method as `tsh`. Ensure you have logged in using `tsh` at least once:

```sh
tsh login --proxy=teleport.example.com
```

### Basic Commands

Simply run `tssh` by calling it in the console:

```sh
tssh
```

### Changing SSH User

To change the SSH user while running `tssh`, press `Ctrl+U`.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE.txt) file for more details.
