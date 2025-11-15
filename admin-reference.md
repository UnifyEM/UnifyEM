# Admin Reference

UnifyEM consists of three components: A server with an HTTP API, an agent for installation on each endpoint, and an
admin interface. At this time administrators use a Command Line Interface (CLI) to interact with the server via the API.
In the future a web-ui will be created, and the API-first design facilitates integration.

From an administrator's perspective, it is helpful to understand that within UnifyEM there are three types of commands:

**Agent requests** are queued in the server database. Each time the agent sends a sync request, the server includes a
list of requests in the HTTP response payload. The agent then executes each request and creates an agent response. A
list of responses is kept in memory until the next sync, at which time the agent includes the list of responses in the
HTTP sync request. While responses are pending, the agent will attempt a sync every 60 seconds.

**Agent triggers** represent immediate actions to be taken by the agent. These include setting the agent in lost mode,
lockout mode, uninstalling the agent, and wiping the drive.

**Admin commands** are used to interact with the server via the API. Some admin commands generate agent requests or set
agent triggers. Others set server parameters or query information from the server's database.

# Agent Requests

Agent requests consist of a command, required parameters, and optional arguments. The server also includes metadata
including the identity of the requester and a unique request ID that allows responses from the agent to be associated
with the request.

The following requests are currently supported:

```
download_execute url=<url> [arg1=<arg> arg2=<arg>...]

execute cmd=<program> [arg1=<arg> ...]

ping

reboot

shutdown

status

tags <agent ID>

tag-add <agent ID>

tag-remove <agent ID>

upgrade

user_add agent_id=<agent ID> user=<user> password=<password> [admin=<true | false>]

user_admin agent_id=<agent ID> user=<user> admin=<true | false>

user_password agent_id=<agent ID> user=<user> password=<password>

user_list agent_id=<agent ID>

user_lock agent_id=<agent ID> user=<user> [shutdown=yes]

user_unlock agent_id=<agent ID> user=<user>
```

# Agent Triggers

Agent triggers are sent as a JSON object with three boolean values.

**Lost** is intended to help trace and locate a lost or stolen device. When lost mode is activated, the agent attempts
to sync once per minute. In the future, it will attempt to send additional location information.

**Uninstall** will cause the agent to attempt to uninstall itself.

**Wipe** will cause the agent to attempt to delete all data, and/or take other steps to render data on the device
inaccessible. While there are no guarantees, this trigger is intended to destroy data and once received by the agent
cannot be reversed.

When an `uninstall` or `wipe` trigger is received, the agent will attempt to send an acknowledgment to the server prior
to executing the trigger.

Note: If the administrator's intent is to deny a user access to a company-owned device, resetting the user's password
and rebooting the device may be a safer approach. Another option is to disable the user's account, but care must be used
to ensure that encrypted drives can be accessed.

### Admin Commands

Formal API documentation will be produced in the near future. In the interim, a list of API endpoints can be found in
common/schema/api.go

Admin comments use the server's API. For CLI syntax, use `./uem-cli --help` or `uem-cli.exe --help` on Windows.

`uem-cli agent <subcommand> <args>` is used to obtain information about agents, setting their name, adding and removing
tags, and setting (possibly resetting) triggers.

`uem-cli cmd <subcommand> <args>` is used to send agent-specific requests , specify agent_id, or a tag to apply the
command to.

`uem-cli config <agents | server> <get | set> [args]` is used to set and retrieve server configuration parameters.

`uem-cli events <subcommand> <args>` provides access to event logs. At this time specifying an agent_id argument is
required.

`uem-cli help` displays help for the CLI or a command.

`uem-cli ping` is used to test authentication and communication with the server.

`uem-cli regtoken [new]` retrieve the registration token or generate a new one.

`uem-cli report` requests reports from the agent. (More work is required on report generation.)

`uem-cli request` is used to query the server for information about agent requests and delete them. Note that each time
`uem-cli cmd` is used to create an agent request, a unique request ID is returned. `uem-cli request get <request-id>`
can be used to query the status of the request including any response received from the agent.

`uem-cli verstion` displays version, copyright, and legal information.
