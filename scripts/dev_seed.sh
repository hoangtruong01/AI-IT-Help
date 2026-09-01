#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
env_file="$repo_root/.env"

if [ ! -f "$env_file" ]; then
  echo ".env is required so the seed command can verify APP_ENV." >&2
  exit 1
fi

read_setting() {
  sed -n "s/^$1=//p" "$env_file" | tail -n 1 | tr -d '"\r'
}

app_env=$(read_setting APP_ENV)
case "$app_env" in
  development|test) ;;
  *) echo "Refusing to seed APP_ENV='$app_env'. Only development or test is allowed." >&2; exit 1 ;;
esac

db_user=$(read_setting POSTGRES_USER)
[ -n "$db_user" ] || db_user=eomp

seed_database() {
  database=$1
  file=$2
  echo "Seeding $database from $file..."
  docker compose --project-directory "$repo_root" exec -T postgres \
    psql -v ON_ERROR_STOP=1 -U "$db_user" -d "$database" \
    < "$repo_root/scripts/dev-seeds/$file"
}

seed_database "$(read_setting ASSET_DB_NAME)" asset.sql
seed_database "$(read_setting HELPDESK_DB_NAME)" helpdesk.sql
seed_database "$(read_setting WORKFLOW_DB_NAME)" workflow.sql
seed_database "$(read_setting KNOWLEDGE_DB_NAME)" knowledge.sql
seed_database "$(read_setting NOTIFICATION_DB_NAME)" notification.sql

echo "Development seed completed for APP_ENV=$app_env."
