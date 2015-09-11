# dropletconn
A simple golang base CLI app to list and connect to your DigitalOcean droplets

To use, `go get github.com/theonejb/dropletconn` and `go install github.com/theonejb/dropletconn`. `dropletconn` is the
name of the genrated binary. Available commands:
 - `config`: Generate config file that stores the API token and other settings. This needs to be generated before the rest of
 the commands can be used
 - `list <FILTER EXPRESSION>`: Lists all droplets from your account. You can optionally pass a number of filter expressions.
 If you do, only droplets whose names or IPs contain at least one of the given fitler expressions will be listed
 - `connect NAME`: Connect to the droplet with the given name

Both `list` and `connect` accept an optional `--force-update` flag. By default, the list of droplets is cached for a configurable duration (as set in
the config file). Passing this flag forces an update of this list before running the command.
