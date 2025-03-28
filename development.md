# Development Notes

## Adding agent commands:

1. Define the command in common/schema/commands/commands.go
   - Add a const to name the command
   - In init() add the command along with a list of required an and optional arguments
   - Since maps are not ordered, allArgN(x) adds a list of arguments from opt1 to optx. This allows sending commands that require parameters to remain in the correct order.


2. Add the commend in cli/functions/cmd/cmd.go. As long as there is one command and arguments, it can simply be added as a subcommand to cmd. This will allow the command to be used with "uem-cli cmd <subcommand>" where subcommand is the command specified above along with any arguments provided in key=value pairs. For example, if you create a command `woof` with mandatory arguments `at` and `agent_id` then the following command will be accepted: 

   `uem-cli cmd woof at=truck agent_id=0123-456789-00000-00000`

   Note that agent_id and request_id are automatically added as optional parameters to every command. However, an agent_id is required for the server to send the command to an agent.


3. Recompile uem-server. This will allow the new command to validate in the postCmd handler and be queued for agent. If a command does not pass validation, the server will refuse to send it to an agent.


4. Implement the command in UEM agent:
   - Create a new Go package in agent/functions. The ping and status packages provide examples.
   - The new command must implement the `CmdHander` interface:
  ```
   type CmdHandler interface {
     Cmd(schema.AgentRequest) (schema.AgentResponse, error)
   }
  ```
   - If your new command could be potentially dangerous, add code to abort if global.PROTECTED is true.
   - Add the new command package in the New() function in agent/functions/functions.go. Note that the key must be the same string as defined in step 1 above. Using the constant is preferable.
   - Note that every log event has a unique integer as the first argument to assist with debugging.
   - Disabling a command in the agent can be achieved by commenting out the `c.addHandler` line in agent/functions/functions.go Main()