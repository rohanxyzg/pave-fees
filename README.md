# PAVE Fees API

A progressive fee accrual system built with Encore and Temporal. This API manages bills with line items and calculates fees progressively as items are added, using Temporal workflows for reliable processing.

## What it does

- Create bills for customers (USD or GEL currency)
- Add line items to active bills 
- Close bills to finalize totals
- Progressive fee calculation using Temporal workflows
- Query bills by customer and status

## Architecture

- **Encore Framework**: HTTP API with automatic routing and database management
- **Temporal**: Workflow orchestration for progressive fee accrual
- **PostgreSQL**: Bill and line item storage (managed by Encore)
- **Docker Compose**: Temporal server, UI, and PostgreSQL database

## Prerequisites

- Go 1.21+
- [Encore CLI](https://encore.dev/docs/install)
- Docker and Docker Compose

## Setup

1. **Install Encore CLI**
   ```bash
   curl -L https://encore.dev/install.sh | bash
   ```

2. **Start Temporal services**
   ```bash
   docker-compose up -d
   ```
   This starts:
   - PostgreSQL database on port 5432
   - Temporal server on port 7233
   - Temporal Web UI on http://localhost:8080

3. **Install Go dependencies**
   ```bash
   go mod download
   ```

4. **Run the application**
   ```bash
   encore run
   ```
   The API will be available at http://localhost:4000

## Services Running

- **API**: http://localhost:4000 (Encore app)
- **Temporal UI**: http://localhost:8080 (workflow monitoring)
- **PostgreSQL**: localhost:5432 (database)
- **Temporal Server**: localhost:7233 (workflow engine)

## API Endpoints

### Create Bill
```bash
POST /bills
{
  "customer_id": "customer-123",
  "currency": "USD",
  "description": "Monthly services"
}
```

### Add Line Item
```bash
POST /bills/{bill_id}/items
{
  "description": "Consulting services",
  "amount": 5000
}
```

### Close Bill
```bash
POST /bills/{bill_id}/close
```

### Get Bill
```bash
GET /bills/{bill_id}
```

### List Bills
```bash
GET /customers/{customer_id}/bills?status=OPEN&limit=10&offset=0
```

## Currency

Amounts are in cents/tetri:
- USD: $50.00 = 5000 cents
- GEL: ₾50.00 = 5000 tetri

## Bill States

- **OPEN**: Active, can add items
- **CLOSED**: Finalized, total calculated

## Temporal Workflow

Each bill starts a Temporal workflow that:
1. Listens for line item signals
2. Accumulates items until close signal
3. Calculates final total
4. Updates bill status to CLOSED

View workflows at http://localhost:8080 when services are running.

## Testing

```bash
# Run all tests
go test ./...

# Run specific test suite
go test ./fees -run TestBillService
go test ./fees -run TestBillWorkflow
```

## Database

Encore automatically manages PostgreSQL for the application data. Temporal uses the PostgreSQL instance from docker-compose for its own data.

Database schema is in `fees/migrations/`.

To reset application database:
```bash
encore db reset
```

## Development

The codebase follows clean architecture:

- `fees/fees.go` - HTTP handlers (Encore API endpoints)
- `fees/service.go` - Business logic layer  
- `fees/repository.go` - Data access layer
- `fees/workflow.go` - Temporal workflow definitions
- `fees/activity.go` - Temporal activity functions
- `fees/types.go` - Data structures and validation

## Docker Services

Stop all services:
```bash
docker-compose down
```

View service logs:
```bash
docker-compose logs -f temporal
docker-compose logs -f postgres
```

Reset Temporal data:
```bash
docker-compose down -v
docker-compose up -d
```

## Troubleshooting

**Temporal connection issues:**
- Ensure `docker-compose up -d` is running
- Check Temporal server logs: `docker-compose logs temporal`
- Verify port 7233 is available

**Database issues:**
- Check PostgreSQL logs: `docker-compose logs postgres`  
- Run `encore db reset` to recreate application schema
- Check `fees/migrations/` for schema definitions

**Build issues:**
- Ensure you're using `encore run` (not `go run`)
- Make sure all Docker services are up before starting the app

**Port conflicts:**
- PostgreSQL: 5432
- Encore API: 4000
- Temporal Server: 7233
- Temporal UI: 8080

Make sure these ports are available or modify docker-compose.yml as needed.