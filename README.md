# UnifyEM

Unify Endpoint Management (UnifyEM) is free open source, self-hosted software made in Canada. It is designed to help individuals, small, and medium-sized organizations effectively monitor, manage, and secure their endpoints. By streamlining the oversight and maintenance of devices across a business, UnifyEM supports critical security, compliance, and audit objectives.

We have chosen the Apache licence to facilitate widespread adoption and encourage contributions from the community. If this presents issues for your organization please contact us for alternate licensing arrangements.

UnifyEM components are written in Go and designed to be simple to deploy and upgrade. The system consists of four components:

- **uem-server**: The main server component that includes an embedded database and follows an API-first approach for seamless integration with other products.
- **uem-agent**: A lightweight agent that manages each computer’s configurations, policies, and security posture. 
- **uem-cli**: A command-line interface enabling control over administrative tasks via the server's API.
- **uem-webui**: A web-based interface (future).

There are also a number of packages in `common` that are shared across the components and available for other open source projects.

## Development status

This software is under active development. **Testing is required prior to production deployment.**

The `main` branch is intended to be stable. All other branches are for development and testing purposes.

Supported operating systems:

- **Linux:** Developed and tested on Ubuntu 24.04 amd64 (agent requires a lot of testing)
- **macOS:** Developed and tested on macOS Sequoia 15, arm64.
- **Windows:** Developed and tested on Windows 11 amd64 and arm64.
- **Future development:** Android, iOS, iPadOS.

## Whois is UnifyEM for?

This software's initial goal is to provide simple, effective centralized endpoint management for small and medium
businesses, including:

- Basic operations support such as adding users, deleting users, and resetting passwords
- Endpoint locking and erasing
- Providing security-related evidence for SOC 2 and ISO 27001 audits

This software is not intended to replace sophisticated platform-specific solutions. For example, if your organization uses exclusively Microsoft Entra ID (formerly known as AzureAD) connected Windows PCs, Intune is a better solution. If you only have Apple devices, please check out Apple's MDM capabilities.

On the other hand, if you have a mix of platforms in your organization, UnifyEM is designed for you. If it doesn't do what you need, please create an issue and let us know.

## Why does UnifyEM exist?

Every application has a story. This software was inspired by more than three decades of cybersecurity experience and propelled by the urgent need to secure endpoints, streamline administration, and provide evidence of compliance for SOC 2 and ISO 27001 audits. It also reflects the primary author's growing frustration with open source projects that gate critical features behind commercial licences, and the increasing costs of software that fail to justify their price tag.

Asking users to send screenshots as evidence that their firewall and screen lock are enabled is irritating, and paying more to automate the process than it costs to provide the employee with Google Workspace or Office 365 is difficult to justify. While there are a few low-cost MDM providers, it is difficult to recommend a cloud-based endpoint management solution that fails to meet basic security requirements.

The author has no intention of adding commercial features to this free open source software. However, if you or your organization find this software useful and wish to sponsor new features, desire paid support, or would like expert assistance with your cybersecurity program, please feel free to contact us via our website at https://tenebris.com.

## Acknowledgements

This project is sponsored by Tenebris Technologies Inc., a Canadian cybersecurity consultancy incorporated in 1996. Eric Jacksch, the company's founder, president, and principal consultant is the primary author.

Contributions and sponsorships are welcome, would be greatfully appreciated, and will be recognized here.

## Cautions

During testing, we highly recommend that the `PROTECTED` option is set to `true` in agent\global\global.go. This will disable the `uninstall` and `wipe` triggers, as well as hopefully preventing some user and computer locking functions. When received from the server, triggers are executed as quickly as possible. They therefore can only be reset (aborted) before the agent's next sync.

The `wipe` trigger is disruptive. It is designed to delete data on the endpoint and make it as difficult as possible to recover.

The `lock` trigger changes the password of the currently logged-in user to a random string and reboots the computer. Assuming the drive is encrypted, this should lock the user out. The agent attempts to send the username and random password to the server on a best-effort basis.

Combining `uninstall` with the `lock` or `wipe` triggers may have unpredictable results.

While the agent attempts to add new users in such a way as to allow them to unlock BitLocker and FileVault, this functionaly has not been thorougly tested. Please consider the requirement for encryption unlocking when adding and deleting users.

The agent sync interval is controlled by uem-server. We recommend setting a short syncing interval during testing so that the agent promptly actions pending requests.

**Software bugs or administrator mistakes could result in the inability to access the endpoint. Back up your data.**

## Contributions

**Testing**: Testing on any version of macOS, Windows, and Linux that are currently supported by their respective vendors is appreciated.

**Documentation**: Documentation assistance is always appreciated.

**Functionality**: UnifyEM is designed to be easily expanded with new functionality. Please see Development.md for associated notes.

**Bug fixes**: Find a bug, squash a bug :)

By submitting any code or documentation to this project, you confirm that you own the necessary rights to do so, agree to license your contribution under the project’s open source terms, and warrant that it is free of any third-party claims or conflicts. If you are not authorized to contribute under these conditions, please refrain from submitting.

## Overview

All communication is originated by uem-agent, uem-cli, and uem-webui to the uem-server over HTTPS.

While the Go libraries fully support HTTPS, at this point of development the preferred approach is for uem-server to listen for HTTP on localhost and use NGINX for HTTPS termination. This allows Certbot to easily obtain and renew certificates for HTTPS.

Agent installation requires a server-specific installation key that contains the server's FQDN and a registration token (enrollment code), similar to how most endpoint security products operate. The agent uses this information to register with the server and obtain unique credentials.

Agents register automatically using the registration token and receive a unique agent ID, an access token, and a refresh token. The agent ID and refresh token are stored in the agent configuration. The access token is kept in memory. If the access token is denied, the agent will request a new one using the refresh token. If the refresh token is denied, the agent will attempt re-registration. If successful, the agent will receive a new unique agent ID.

Each time the agent checks in with the server it retrieves a list of pending commands and sends any queued responses. It also receives configuration information such as check in and status report frequencies. Status reports include information on the agent computer's functional and security status.

If the agent's record is deleted from the server database, access will be denied even though the tokens may still be valid. This will cause the agent to attempt re-registration using the registration token it was provided at installation.

To remove an agent, the preferable method is to send an uninstall command. This will cause the agent to uninstall itself as a service and stop running. However, in the event of a security issue, changing the registration token (`uem-cli regtoken new`) and then deleting the agent record from the server (`uem-cli agent delete <agent ID>`) will prevent the agent from being able to re-register.

## Security Model Summary

The security model is quickly evolving and will be more fully documented at a later date.

Agents register to the server using the installaton key and are given an access key and a refresh key. By default, the refresh key does not expire. The agent stores the refresh key. The access key is kept in RAM and is used to authenticate to the server. If it expires or the agent restarts, the refresh key is used to obtain a new access key. If the configuration is changed such that the refresh key expires, the agent will attempt a new registation. Note that a new registration gives the agent a new identity since it would be unwise to allow an agent to re-register with an unproven identity.

If the CA pinning feature is enabled, when the agent next connects to the server, it retains a hash of the CA public key (the last certificate in a verify chain). From that point forward, it will refuse to connect to the server if the server's SSL/TLS certificate does not chain to the same CA. This allows certificates from services such as LetsEncrypt to be used while providing some MITM attack mitigation.

When the server requests an agent to download an execute a file, it includes an SHA265 hash of the file in the request. When the server instructs the agent to upgrade, it includes the SHA256 hash of a deployment file which, in turn, lists the SHA256 hashes of all agents available for download. The agent will discard any file that can not be verified. (For development and transition this can be disabled in agent/global/global.go.

When updated clients are placed in the download directory, the administrator must initiate a refresh of the deployment file. This can be done using the CLI (`uem-cli files deploy`). Failure to update the hashes in the deployment file will prevent the agents from upgrading unless hash verification is disabled.

A transition is in process to all requests being digitally signed by the server. Once the agent receives a configuration containing the server's public signing key, it will refuse to accept any request that is not digitally signed. (For development purposes this can be disabled in agent/global/global.go)

Administrators authenticate to the server using their username and password, and receive a refresh and access token. The refresh token lifetime for users ("refresh_token_life_users") defaults to 1440 minutes, after which the user will need to re-authenticate. This is configurable. At this point only one administrator is allowed. Expanding this and adding MFA is on the short-term roadmap.

## Build and deploy

Each of the components can be built by changing to their directory and using Go build. For example,

```
git clone https://github.com/UnifyEM/UnifyEM.git
cd UnifyEM
mkdir -p bin
cd server
go build -o ../bin/uem-server
```

Each component is a single statically-linked binary.

**The author's script from his Ubuntu 24 UEM server is included as uem-build.sh. This script compiles uem-server and uem-cli and deploys them, stopping and re-starting uem-server as required. It then builds the various agents and copies them to the default file distribution directory. Note that this script uses sudo. Please review it prior to running it on your ocmputer.**

### uem-server installation

`./uem-server install` will install the server as a service. On Windows, configuration information is stored in the registry and a data directory is created in ProgramData. On macOS and Linux, configuration information is written to /etc/uem-server.conf and a data directory is created in /opt/uem-server.

By default, the server will listen on http://127.0.0.1:8080. If you encounter difficulties, you can temporarily bypass the configured listen address and start uem-server in the foreground (i.e. not as a deamon/service) using `uem-server listen 127.0.0.1:8080` or another suitable address. This is useful in the event that a mistake in the configuration prevents uem-server from starting.

To change the listen URL, the external URL, or other configuration, update them using `uem-cli config server`, stop the service and change the registry or /etc/uem-server.conf file as appropriate.

`./uem-server admin <username> <password>` will create a super administrator account. There are no default accounts. The ability to add and maintain regular administrators via the API will be added in the near future.

Note that attempting to create a super admin while the server is running may fail due to database locking.

The server is currently designed to run with root/admin privileges to allow it to install, etc. The ability to run as a non-root user may be added in the future. To install, the user will need to enter their password (Linux and macOS) or confirm the installation (Windows).

`./uem-server uninstall` will remove the service from the system.

At this point we recommend the uem-server default of listening for HTTP on localhost and NGINX as a proxy that provides TLS termination. This has the benefit of out-of-the-box certbot compatibility.

By default, logs are written to /var/log/uem-server.log on Linux and macOS. On Windows, log events are sent to the Windows Event Log and, but default, also to c:\ProgramData\uem-server\uem-server.log. Logs are rotated daily and by default retained for 30 days. The retention period can be changed in the configuration file/registry.

### uem-cli installation

uem-cli is a command-line interface for administration use only. Authentication and server information from the environment is used to authenticate with the server and obtain an access and refresh token. If a file in the user's home directory named `.uem` exists, it will be loaded into the environment.

The following environment variables are required:

```
UEM_USER: The administrator's username
UEM_PASS: The administrator's password
UEM_SERVER: The protocol, FQDN, and port of the server (i.e. https://uem.example.com:443)
```

Additional administrator accounts, along with managing them via the API, will be added in the near future. Until this occurs, the only admin-level credentials are usernames and passwords set from the uem-server command line.

### uem-agent installation

Reminder: The agent is not yet supported on Linux. Attempting to compile it on Linux will fail. This will be addressed in the near future.

The agent can be installed by running:

```
./uem-agent install <installation token>
```

Note that the registration token is the public URL of the server followed by a slash and a randomly generated token. The registration token can be viewed in the configuration or retrieved from the server's API using `./uem-cli regtoken`.

Note that the same registration token is used by all agents. Changing the registration token will not affect agents that are already registered unless they become deregistered. To generate a new registration token, use `./uem-cli regtoken new`.

For testing purposes, the agent can be installed and immediately uninstalled. It will leave the configuration information in place.

Note: The agent requires root/administrator privileges to perform many functions and therefore tests for elevated privileges on startup. To install, the user will need to enter their password (Linux and macOS) or confirm the installation (Windows).

### uem-webui installation

This component has not yet been developed.

## Copyright and license

Copyright (c) 2024-2025 by Tenebris Technologies Inc. and available for use under Apache License 2.0. Please see the LICENSE file for full information.

## Warranty

THIS SOFTWARE IS PROVIDED “AS IS,” WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
