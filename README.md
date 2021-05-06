# Utility for working with AWS EC2 development virtual machines

This is a small app for starting and stopping an EC2 instance, and fetching its public DNS name so that you can easily SSH into it (useful for VSCode Remote SSH). It currently uses STS and an MFA token to log into an IAM user account before interacting with EC2.

The driver script `awsdev.sh` demonstrates usage with environment variables to set up most of the configuration. The only mandtory argument is the MFA token code.

## Setting up an EC2 instance

Create a new instance in EC2's console (I like m6g instances running Ubuntu) and download the PEM file to enable SSH login. Add the machine to the SSH config to enable SSH.

```ssh-config
Host vscode-remote
  HostName ec2-XX-XX-XX-XX.eu-west-1.compute.amazonaws.com
  User ubuntu
  IdentityFile ~/.cert/vscode-remote-001.pem
```

Log into the machine and set up some basic utilities:

```shell
ssh vscode-remote
sudo apt update
sudo apt dist-upgrade
sudo apt install build-essential
```

Copy Git config (`~/.gitconfig`) over from local machine.

Copy NPM config (`~/.npmrc`) over from local machine.

Create an SSH key and add to GitHub. [https://docs.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent]

Create a GPG key, add to GitHub and configure in Git client [https://docs.github.com/en/github/authenticating-to-github/generating-a-new-gpg-key]

Add to the bottom of the bashrc file:

```shell
eval "$(nodenv init -)"
export EDITOR="vi"

eval "$(ssh-agent -s)" &>/dev/null
ssh-add ~/.ssh/id_github
```

Clone the repo and get set up.

```shell
git clone git@github.com/example/example.git
```

## Build instructions

```shell
go build aws-dev-vm.go
```

## Usage

```shell
awsdev -t TOKEN_CODE [start|stop|describe]
```

The public DNS address can be added to your SSH config as needed.
