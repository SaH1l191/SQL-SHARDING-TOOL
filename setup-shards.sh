#!/bin/bash
# Multi-Shard PostgreSQL Setup Script
# This script creates multiple PostgreSQL containers to simulate a sharded database setup

# ============================================================================
# CONFIGURATION - Modify these values as needed
# ============================================================================

# Number of shards to create
SHARD_COUNT=2

# Base port number (shards will use ports BASE_PORT, BASE_PORT+1, BASE_PORT+2, etc.)
BASE_PORT=5433

# PostgreSQL credentials (same for all shards)
DB_NAME=temp_db
DB_USER=temp_user
DB_PASSWORD=temp_password

# PostgreSQL version
PG_VERSION=18

# ============================================================================
# DOCKER COMPOSE FILE GENERATION
# ============================================================================

cat > docker-compose.shards.yml << 'EOF'
version: '3.8'

services:
EOF

# Generate service entries for each shard
for i in $(seq 0 $((SHARD_COUNT - 1))); do
    PORT=$((BASE_PORT + i))
    cat >> docker-compose.shards.yml << EOF
  postgres-shard-${i}:
    image: postgres:${PG_VERSION}-alpine
    container_name: postgres-shard-${i}
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    ports:
      - "${PORT}:5432"
    volumes:
      - postgres_shard_${i}_data:/var/lib/postgresql/data
    networks:
      - shard-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5

EOF
done

# Add volumes and networks section
cat >> docker-compose.shards.yml << 'EOF'
volumes:
EOF

for i in $(seq 0 $((SHARD_COUNT - 1))); do
    echo "  postgres_shard_${i}_data:" >> docker-compose.shards.yml
done

cat >> docker-compose.shards.yml << 'EOF'

networks:
  shard-network:
    driver: bridge
EOF

# ============================================================================
# CREATE SHARD CONNECTION CONFIG FILE
# ============================================================================

cat > shard-connections.txt << EOF
# Shard Connection Configuration
# Use these details when configuring shards in the SQL-Sharding-Tool UI
#
# Instructions:
# 1. Create a project in the UI
# 2. Add shards (one per shard below)
# 3. For each shard, click "Connection" and enter the corresponding details
#
# Generated for ${SHARD_COUNT} shards

EOF

for i in $(seq 0 $((SHARD_COUNT - 1))); do
    PORT=$((BASE_PORT + i))
    cat >> shard-connections.txt << EOF
# Shard ${i}
Host: localhost
Port: ${PORT}
Database: ${DB_NAME}
User: ${DB_USER}
Password: ${DB_PASSWORD}

EOF
done

# ============================================================================
# UTILITY COMMANDS
# ============================================================================

cat > shard-commands.sh << 'EOF'
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
EOF

chmod +x shard-commands.sh

# ============================================================================
# SUMMARY OUTPUT
# ============================================================================

cat << EOF

╔══════════════════════════════════════════════════════════════════════════╗
║                   MULTI-SHARD SETUP COMPLETE                               ║
╚══════════════════════════════════════════════════════════════════════════╝

Generated Files:
────────────────
1. docker-compose.shards.yml  - Docker Compose configuration for ${SHARD_COUNT} shards
2. shard-connections.txt     - Connection details for each shard
3. shard-commands.sh         - Management script for shards

Shard Configuration:
────────────────────
Shards:     ${SHARD_COUNT}
Base Port:  ${BASE_PORT}
Database:   ${DB_NAME}
User:       ${DB_USER}
Password:   ${DB_PASSWORD}

Quick Start:
────────────
1. Start shards:     ./shard-commands.sh start
2. Check status:     ./shard-commands.sh status
3. Test connections: ./shard-commands.sh test

UI Configuration Steps:
───────────────────────
1. Create a project in SQL-Sharding-Tool
2. Add ${SHARD_COUNT} shards (Overview tab → Shards → Add Shard)
3. For each shard, click "Connection" button
4. Enter details from shard-connections.txt:
   - Shard 0 uses port ${BASE_PORT}
   - Shard 1 uses port $((BASE_PORT + 1))
   - etc.
5. Activate each shard after configuring connection

Cleanup:
────────
./shard-commands.sh clean  # Removes all containers and data

EOF
