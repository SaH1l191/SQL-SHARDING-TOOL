#!/bin/bash
# Shard Management Commands

COMPOSE_FILE="docker-compose.shards.yml"

case "$1" in
  start)
    echo "Starting all shard containers..."
    docker-compose -f $COMPOSE_FILE up -d
    echo "Waiting for shards to be healthy..."
    sleep 5
    docker-compose -f $COMPOSE_FILE ps
    ;;
  
  stop)
    echo "Stopping all shard containers..."
    docker-compose -f $COMPOSE_FILE down
    ;;
  
  status)
    echo "Shard container status:"
    docker-compose -f $COMPOSE_FILE ps
    ;;
  
  logs)
    echo "Showing logs for all shards..."
    docker-compose -f $COMPOSE_FILE logs -f
    ;;
  
  clean)
    echo "WARNING: This will delete all shard data!"
    read -p "Are you sure? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
      docker-compose -f $COMPOSE_FILE down -v
      echo "All shards and data removed."
    else
      echo "Cancelled."
    fi
    ;;
  
  test)
    echo "Testing shard connections..."
    # Extract shard count from compose file
    SHARD_COUNT=$(grep -c "postgres-shard-" $COMPOSE_FILE)
    BASE_PORT=5433
    for i in $(seq 0 $((SHARD_COUNT - 1))); do
      PORT=$((BASE_PORT + i))
      if pg_isready -h localhost -p $PORT -U temp_user > /dev/null 2>&1; then
        echo "✓ Shard $i (port $PORT): ONLINE"
      else
        echo "✗ Shard $i (port $PORT): OFFLINE"
      fi
    done
    ;;
  
  *)
    echo "Usage: $0 {start|stop|status|logs|clean|test}"
    echo ""
    echo "Commands:"
    echo "  start  - Start all shard containers"
    echo "  stop   - Stop all shard containers"
    echo "  status - Show container status"
    echo "  logs   - View container logs"
    echo "  clean  - Remove all containers and data (DANGER)"
    echo "  test   - Test connection to all shards"
    ;;
esac
