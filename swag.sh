#!/bin/sh
################################################################################
# Copyright (c) 2024-2025 Tenebris Technologies Inc.                           #
# Please see the LICENSE file for details                                      #
################################################################################

swag init --v3.1 -g main.go -d server --parseDependency --parseInternal --parseDepth 5
