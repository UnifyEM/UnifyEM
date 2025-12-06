/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package global

import (
	"runtime"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

const (
	ConfigServerSet             = "server_config"
	ConfigLogFile               = "log_file"
	ConfigLogStdout             = "log_stdout"
	ConfigLogRetention          = "log_retention"
	ConfigListen                = "listen"
	ConfigExternalULR           = "external_url"
	ConfigDataPath              = "data_path"
	ConfigFilesPath             = "files_path"
	ConfigDBPath                = "db_path"
	ConfigHTTPTimeout           = "http_timeout"
	ConfigHTTPIdleTimeout       = "http_idle_timeout"
	ConfigMaxConcurrent         = "max_concurrent"
	ConfigPenaltyBoxMin         = "penalty_box_min"
	ConfigPenaltyBoxMax         = "penalty_box_max"
	ConfigHandlerTimeout        = "handler_timeout"
	ConfigAccessTokenLife       = "access_token_life"
	ConfigRefreshTokenLifeUsers = "refresh_token_life_users"
	ConfigAuthorizedAdminIPs    = "authorized_admin_ips"
	ConfigRequestRetries        = "request_retries"
	ConfigRequestRetryDelay     = "request_retry_delay"
	ConfigAgentRetention        = "agent_retention_days"
	ConfigEventRetention        = "event_retention_days"
	ConfigRequestRetention      = "request_retention_days"

	ConfigPrivate                = "server_private"
	ConfigRegToken               = "reg_token"
	ConfigJWTKey                 = "jwt_key"
	ConfigRefreshTokenLifeAgents = "refresh_token_life_agents"
	ConfigServerECPrivateSig     = "ec_private_sig"
	ConfigServerECPublicSig      = "ec_public_sig"
	ConfigServerECPrivateEnc     = "ec_private_enc"
	ConfigServerECPublicEnc      = "ec_public_enc"
)

// setDefaults makes sure the sets exist, sets default values, and constraints
func setDefaults(c interfaces.Config) (interfaces.Parameters, interfaces.Parameters) {

	// Server configuration set
	sc := c.NewSet(ConfigServerSet)
	sc.SetConstraint(ConfigLogFile, 0, 0, "")                          // no log file by default
	sc.SetConstraint(ConfigLogStdout, 0, 0, true)                      // by default log to stdout
	sc.SetConstraint(ConfigLogRetention, 1, 0, 365)                    // days
	sc.SetConstraint(ConfigListen, 0, 0, "127.0.0.1:8080")             // listen address
	sc.SetConstraint(ConfigExternalULR, 0, 0, "http://127.0.0.1:8080") // external URL (should be FQDN for production)
	sc.SetConstraint(ConfigDataPath, 0, 0, "")                         // data path (base directory for data)
	sc.SetConstraint(ConfigFilesPath, 0, 0, "")                        // files path
	sc.SetConstraint(ConfigDBPath, 0, 0, "")                           // database path
	sc.SetConstraint(ConfigHTTPTimeout, 0, 0, 30)                      // seconds
	sc.SetConstraint(ConfigHTTPIdleTimeout, 0, 0, 30)                  // seconds
	sc.SetConstraint(ConfigMaxConcurrent, 0, 0, 100)                   // number of concurrent connections, others will wait
	sc.SetConstraint(ConfigPenaltyBoxMin, 0, 0, 1000)                  // Minimum penalty box time in milliseconds
	sc.SetConstraint(ConfigPenaltyBoxMax, 0, 0, 5000)                  // Maximum penalty box time in milliseconds
	sc.SetConstraint(ConfigHandlerTimeout, 0, 0, 30)                   // seconds
	sc.SetConstraint(ConfigAccessTokenLife, 0, 0, 720)                 // minutes
	sc.SetConstraint(ConfigRefreshTokenLifeUsers, 0, 0, 1440)          // minutes
	sc.SetConstraint(ConfigAuthorizedAdminIPs, 0, 0, "127.0.0.1")      // default to localhost
	sc.SetConstraint(ConfigRequestRetries, 0, 0, 3)                    // default to 3 retries
	sc.SetConstraint(ConfigRequestRetryDelay, 0, 0, 600)               // seconds
	sc.SetConstraint(ConfigAgentRetention, 1, 0, 365)                  // days
	sc.SetConstraint(ConfigEventRetention, 1, 0, 365)                  // days
	sc.SetConstraint(ConfigRequestRetention, 1, 0, 365)                // days

	// Protected configuration items
	sp := c.NewSet(ConfigPrivate)
	sp.SetConstraint(ConfigJWTKey, 0, 0, "")
	sp.SetConstraint(ConfigRegToken, 0, 0, "")
	sp.SetConstraint(ConfigRefreshTokenLifeAgents, 0, 0, 0)
	sp.SetConstraint(ConfigServerECPrivateSig, 0, 0, "")
	sp.SetConstraint(ConfigServerECPublicSig, 0, 0, "")
	sp.SetConstraint(ConfigServerECPrivateEnc, 0, 0, "")
	sp.SetConstraint(ConfigServerECPublicEnc, 0, 0, "")

	// Return the sets
	return sc, sp
}

// DefaultLog is used to create a log location if the usual approach fails
func DefaultLog() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData\\" + LogName + "\\" + LogName + ".log"
	}
	return "/var/log/" + LogName + ".log"
}
