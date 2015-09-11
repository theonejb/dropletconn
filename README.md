# dropletconn
A simple golang base CLI app to list and connect to your DigitalOcean droplets

Generate a DigitalOcean APIv2 token and paste it in ~/.dropletconn.token. Then, you can run:
 - `dropletconn list FILTER_STRING_1 .. FILTER_STRING_N`
 - `dropletconn connect EXACT_NODE_NAME`

All strings used to match droplet names (filter string, node names) are case insensitive.

`dropletconn` caches the API response from DigitalOcean for 5 minutes.  You can pass the
`--force-update` option to force update the cache to all of the commands.
