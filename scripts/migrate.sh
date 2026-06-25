#!/usr/bin/env bash
# 数据库迁移占位脚本。
# 方案A：GORM AutoMigrate（在应用启动或独立 migrate 命令中执行）。
# 方案B：将 migrations/*.sql 用 golang-migrate 等工具按序执行。
set -euo pipefail

echo "TODO: 执行数据库迁移（GORM AutoMigrate 或 migrations/*.sql）"
