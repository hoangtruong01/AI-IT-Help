#!/usr/bin/env bash
# ==============================================================================
# EOMP — Chaos Engineering & Resilience Testing CLI (Linux / macOS / CI)
# ==============================================================================

set -e

show_help() {
    echo "================================================================"
    echo "  EOMP - Phase 14 Chaos Engineering & Resilience CLI"
    echo "================================================================"
    echo ""
    echo "Usage: ./scripts/chaos.sh [command]"
    echo ""
    echo "Commands:"
    echo "  simulate-db-down    Pause PostgreSQL container"
    echo "  restore-db          Unpause PostgreSQL container"
    echo "  simulate-rabbit-jam Stop RabbitMQ container"
    echo "  restore-rabbit      Start RabbitMQ container"
    echo "  run-all-chaos       Execute full automated chaos drill cycle"
    echo ""
}

simulate_db_down() {
    echo "Injecting Chaos: Pausing PostgreSQL database container..."
    docker pause eomp-postgres || docker pause eomp-prod-postgres || true
    echo "[CHAOS] PostgreSQL is now OFFLINE."
}

restore_db() {
    echo "Restoring PostgreSQL database container..."
    docker unpause eomp-postgres || docker unpause eomp-prod-postgres || true
    echo "[RECOVERED] PostgreSQL is ONLINE."
}

simulate_rabbit_jam() {
    echo "Injecting Chaos: Stopping RabbitMQ container..."
    docker stop eomp-rabbitmq || docker stop eomp-prod-rabbitmq || true
    echo "[CHAOS] RabbitMQ is now STOPPED."
}

restore_rabbit() {
    echo "Restoring RabbitMQ container..."
    docker start eomp-rabbitmq || docker start eomp-prod-rabbitmq || true
    echo "[RECOVERED] RabbitMQ is ONLINE."
}

run_all_chaos() {
    echo "=== Starting Full Automated Chaos Drill Cycle ==="
    simulate_db_down
    sleep 3
    restore_db
    sleep 2
    simulate_rabbit_jam
    sleep 3
    restore_rabbit
    echo "All Chaos scenarios completed successfully."
}

case "$1" in
    simulate-db-down)    simulate_db_down ;;
    restore-db)          restore_db ;;
    simulate-rabbit-jam) simulate_rabbit_jam ;;
    restore-rabbit)      restore_rabbit ;;
    run-all-chaos)       run_all_chaos ;;
    *)                   show_help ;;
esac
