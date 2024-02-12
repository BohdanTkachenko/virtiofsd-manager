# virtiofsd-manager

![GitHub License](https://img.shields.io/github/license/BohdanTkachenko/virtiofsd-manager)
![GitHub Release](https://img.shields.io/github/v/release/BohdanTkachenko/virtiofsd-manager)
[![Go](https://github.com/BohdanTkachenko/virtiofsd-manager/actions/workflows/go.yml/badge.svg)](https://github.com/BohdanTkachenko/virtiofsd-manager/actions/workflows/go.yml)
[![GoReleaser](https://github.com/BohdanTkachenko/virtiofsd-manager/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/BohdanTkachenko/virtiofsd-manager/actions/workflows/goreleaser.yml)

virtiofsd-manager is a straightforward command-line tool written in Go, designed to simplify the management of systemd unit files for virtiofsd. With it, you can easily create, enable, disable, and remove services for sharing directories with virtual machines, streamlining your virtual file system (VFS) management process.

## Features

- Simple Installation: Get up and running with a single command.
- Easy to Use: Manage virtiofsd services with intuitive command-line options.
- Automated Service Management: Quickly create, enable, disable, or remove systemd services.

## Installation

Get virtiofsd-manager set up on your system using one of the following methods:

### One-liner Installation

For a quick installation, execute:

```sh
curl -sL https://github.com/BohdanTkachenko/virtiofsd-manager/raw/main/scripts/install.sh | bash
```

Note: Always review scripts before running them for security purposes.

### Manually

1. Navigate to the [latest release](https://github.com/BohdanTkachenko/virtiofsd-manager/releases/tag/latest) page
2. Select and download the package suitable for your operating system and architecture.
3. Install the downloaded package using your system's package manager, or unpack the archive if you prefer a more hands-on approach.

## Building from Source

1. Clone this repo to your machine:

    ```sh
    git clone git@github.com:BohdanTkachenko/virtiofsd-manager.git
    cd virtiofsd-manager
    ```

2. Build with [GoReleaser](https://goreleaser.com/install):

    ```sh
    goreleaser --snapshot --skip=publish --clean
    ```

## Service File Naming Convention

**virtiofsd-manager** employs a standardized naming convention for systemd service files to ensure uniqueness and filesystem compatibility. This convention makes use of the VM's identifier and the shared directory's path, structured as follows:

```
virtiofsd-{vm_id}-{safe_share_path}.service
```

Where:

- `{vm_id}` represents the unique identifier of the virtual machine.
- `{safe_share_path}` is a sanitized version of the shared directory path. In this transformation:
  - Leading and trailing slashes (`/`) are stripped off.
  - All remaining slashes (`/`) within the path are converted to double underscores (`__`).

### Example:

Given a virtual machine with an ID of `123` and a shared directory located at `/data/vm123`, the service file name generated by **virtiofsd-manager** would be:

```
virtiofsd-123-data__vm123.service
```

## Usage

### Sharing Directories with VMs

Create a new systemd service to share a specific directory with your VM:

```sh
sudo virtiofsd-manager install --vm_id=123 --path=/data/vm123
```

This command generates a .service file in /etc/systemd/system/, facilitating the sharing process.

### Managing Sharing Services

**Stop, disable, and remove** a sharing service:

```sh
sudo virtiofsd-manager remove --vm_id=123 --path=/data/vm123
```

To **enable and start** services for a VM, making shared directories accessible:

```sh
sudo virtiofsd-manager enable --vm_id=123
```

**Stop and disable** services to revoke access temporarily:

```sh
sudo virtiofsd-manager disable --vm_id=123
```

### Generating VFS CLI Arguments for QEMU

```sh
sudo virtiofsd-manager get-vfs-args --vm_id=123
```

Output example:

```
-chardev socket,id=data__vm123,path=/run/virtiofsd/123-data__vm123.sock -device vhost-user-fs-pci,chardev=data__vm123,tag=data__vm123
```
