/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

//goland:noinspection GoUnusedConst
const (
	RoleNone = iota
	RoleTest
	RoleAgent
	RoleUser
	RoleAuditor
	RoleAdmin
	RoleSuperAdmin
)

var (
	RolesAll = []int{RoleTest, RoleAgent, RoleUser, RoleAuditor, RoleAdmin, RoleSuperAdmin}
)
