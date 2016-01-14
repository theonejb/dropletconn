# dropletconn
List and connect to your Digital Ocean droplets instantly (*without a .ssh/config*)

## Quick Start

    go get github.com/theonejb/dropletconn
    go install github.com/theonejb/dropletconn
    dropletconn config
    dropletconn list
    dropletconn connect <NAME OF DROPLET>


## Installing and Configuring dropletconn
![Install and Configure dropletconn](https://dl.dropboxusercontent.com/u/14307524/config-safe.png)

## Listing your droplets
![List droplets using dropletconn](https://dl.dropboxusercontent.com/u/14307524/list-safe.png)

## Connecting to a droplet
![Connect to a droplet using dropletconn](https://dl.dropboxusercontent.com/u/14307524/connect-safe.png)

## Usage
To use, `go get github.com/theonejb/dropletconn` and `go install github.com/theonejb/dropletconn`. `dropletconn` is the
name of the genrated binary. I personally have it aliased to `dc` using `export dc=dropletconn` in my `.zshrc` file since
I use it atleast 20 times a day to connect to various servers at work.

You will also need to generate a token from [Digital Ocean API Tokens](https://cloud.digitalocean.com/settings/applications)
that `dropletconn` will use to get a list of droplets available in your account. For safety, use a **Read** only scoped token.

Available commands and their usage is described here. Some commands have a short version as well, which is what you see after the OR pipe (`|`) in their help text below.
 - `config`: Generate config file that stores the API token and other settings. This needs to be generated before the rest of
 the commands can be used
 - `list | l [<FILTER EXPRESSION>]..`: Lists all droplets from your account. You can optionally pass a number of filter expressions.
 If you do, only droplets whose names or IPs contain at least one of the given fitler expressions will be listed
 - `connect | c NAME`: Connect to the droplet with the given name
 - `run | r <FILTER EXPRESSION> <COMMAND>`: Runs the given command on all droplets matching the filter expression. The filter expression is required, and only one filter
 expression can be given

You can pass an optional `--force-update` flag. By default, the list of droplets is cached for a configurable duration (as set in
the config file). Passing this flag forces an update of this list before running the command.

The `list` command also accepts an options `--list-public-ip` flag. If this flag is used *only* the public IP of the nodes is printed, nothing else.
This is incase you want a list of all IPs in your DO account. I needed this to create a Fabric script.

**Note**: The way flags are parsed, you have to list your flags *before* your commands. For example, you can not do `dropletconn list --list-public-ip`.
Instead, you need to do `dropletconn --list-public-ip list`. Same for the `--force-update` flag.

To enable completion of droplet names, source the included Zsh completion file. Credit for that script goes to James Coglan. I copied it from his blog
(https://blog.jcoglan.com/2013/02/12/tab-completion-for-your-command-line-apps/).
