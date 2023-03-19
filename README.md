# KnockoutCity Proxy

This is an open source server that proxies the KnockoutCity Private server to provide per-user authentication. Part of the bot is a Discord bot to allow members of a community to enroll themselves. 

## Commands

Just copying and pasting the `--help` output here. More info below!

```

Usage:
  kocproxy [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  connected   List connected users
  discord     Run the KOC Server Auth Discord bot
  help        Help about any command
  kick        Kick a connected user
  migrate     Run database migrations to upgrade the auth database
  passwd      Generate a hash for a given password
  proxy       Run the KOC Standalone Server Auth Proxy
  web         Run the status webserver

Flags:
  -c, --config string   config file
  -h, --help            help for kocproxy
  -v, --verbose         verbose logging

Use "kocproxy [command] --help" for more information about a command.
```

## Setup

You need a config file. By default, the configuration file is loaded from `$HOME/.kocproxy`. A configuration file can be explicitly specified with the `-c` command line switch.

An example config file is below:

```yaml
proxy:
    backend:
        secret: abcdefg # the secret password set when launching KnockoutCityServer.exe
        upstream: localhost:23500 # the address + port of the KnockoutCityServer.exe
    port: 23600 # port to listen on ... this is what clients will connect to

discord:
    applicationID: 1234567890123456789 # your bot's application ID
    guildID: 9876543210987654321 # what guild your community lives in
    botChannelID: 1111111111111111111 # channel for bot to post its public welcome message
    adminChannelID: 2222222222222222222 # a channel for admin messages 
    adminRoleID: 3333333333333333333 # role ID for users that are allowed to use admin commands on the bot
    token: sometokenhere # your discord auth token
    authdRoleID: 4444444444444444444 # what role to grant users when they create an account

# this section configures the database for storing users
auth:
    db:
        host: localhost
        port: 9876
        username: auth
        password: auth
        name: auth

# for configuring access to the redis server that KnockoutCity uses
koc:
    redis:
        host: localhost:6380

web:
  port: 8080 # port for plaintext HTTP server (for using letsencrypt, this ultimately needs to be port 80, whether it is here or via proxy)
  tls:
    enabled: true # set to true to enable HTTPS. if this is false, all the fields below are not needed
    port: 8081 # HTTPS server (like the plaintext port above, this should be 443 when finally served to the greater internet)
    domain: example.com# # domain to serve on
    cache: ./certs # where to cache certificates
    test: true # enable test mode to test certificate issuing, this will tell the program to use the letsencrypt staging environment

```

## Create the auth db

The next step is to create the auth db. Since I wrote this without knowing what is in the normal Knockout City db, you can't create it in the same schema that the backend server uses (they both have `users` tables....(TODO: fix this?)). 

Once you have your schema created and your config set up, run `kocproxy migrate -v 1`. This will set up the tables for you. When future revisions of the database are made, you will need to re-run this command with different versions. Running `kocproxy migrate` with no version specified will print the current database revision.

### A note on built in servers

One note that if you are using the PostgreSQL instance that Knockout City uses, you may need to tinker with the `pg_hba.conf` file. If you are running this on the same machine, no config is necessary (listens on port `5434`), it is configured to implicitly trust connections from `localhost` (use username `viper` with any password). If you want to connect from other machines, you're on your own for that.

Something similar is in use for the Redis instance as well. Redis has no password and is configured to trust any connections from localhost. Securing Redis is ... a thing, so I'll leave allowing external connections an exercise to the reader (lol don't do it). Also note that the KOC Redis server runs on port `6380`.

## Running everything

You'll probably want two instances of the app running -- a proxy and a discord bot (TODO: make this one?). 

`./kocproxy proxy` to start in proxy mode and `./kocproxy discord` to start the Discord bot.

When the Discord bot starts up, it creates the slash commands in the Discord guild as well as print the welcome message with buttons for users. When the server is killed gracefully (it hooks `SIGINT` (or whatever the Windows version is)), it will de-register the commands and delete that welcome message. 

## Discord Commands

| Command               | Description                                                                                                                                                         |
|-----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `/admin whois <user>` | Prints information about the given user to the admin channel                                                                                                        |
| `/admin ban <user>`   | Bans the user from connecting to the KO City Server. You will be prompted for the duration and reason. If the user is connected to the server, they will be kicked. |
