# PAVE Fees API

Built a fees system that uses Encore for the API and Temporal for workflow stuff. Basically you create bills, add items to them, and when you close the bill it calculates everything using workflows.

## What does it do?

Pretty straightforward - you can create bills for customers (supports USD and GEL), add line items to open bills, close them when done, and the system handles progressive fee calculation via Temporal workflows. Also lets you query bills by customer and status.

## Tech Stack

Using **Encore** because it makes API development simple. **Temporal** handles the workflow orchestration so fee calculation is reliable even if things crash. **PostgreSQL** stores everything and **Docker Compose** runs the Temporal stack.

## Getting Started

You'll need Go 1.21+, the Encore CLI, and Docker.

**Install Encore:**
```bash
curl -L https://encore.dev/install.sh | bash
```

**Fire up the services:**
```bash
docker-compose up -d
```

**Get dependencies and run:**
```bash
go mod download
encore run
```

Now you've got:
- API at http://localhost:4000
- Temporal UI at http://localhost:8080 
- PostgreSQL on 5432
- Temporal server on 7233

## API Usage

**Create a bill:**
```bash
POST /bills
{
  "customer_id": "customer-123",
  "currency": "USD",
  "description": "Monthly services"
}
```

**Add stuff to it:**
```bash
POST /bills/{bill_id}/items
{
  "description": "Consulting services",
  "amount": 5000
}
```

**Close it when done:**
```bash
POST /bills/{bill_id}/close
```

**Get bill details:**
```bash
GET /bills/{bill_id}
```

**List customer bills:**
```bash
GET /customers/{customer_id}/bills?status=OPEN&limit=10&offset=0
```

**Quick notes:** Amounts are in cents (USD) or tetri (GEL) - so $50.00 is 5000. Bills are either OPEN (can add items) or CLOSED (done deal).

## How Temporal Works Here

Each bill gets its own workflow that just sits there listening. When you add items, it gets a signal and accumulates them. When you close the bill, it gets another signal, calculates the final total, and marks everything as done.

## Testing

you need build tags:

```bash
# Run tests (the proper way)
go test -tags=test ./...

# Or run specific ones
go test -tags=test ./fees -run TestBillService
```

## Code Structure

Kept it clean with layers:
- `fees/fees.go` - API endpoints (Encore magic)
- `fees/service.go` - Business logic 
- `fees/repository.go` - Database stuff
- `fees/workflow.go` - Temporal workflows
- `fees/activity.go` - Temporal activities
- `fees/types.go` - Data types and validation

Database schema lives in `fees/migrations/`. Encore handles it automatically but you can reset with `encore db reset` if needed.

## Useful Commands

**Stop everything:**
```bash
docker-compose down
```

**Check what's happening:**
```bash
docker-compose logs -f temporal
docker-compose logs -f postgres
```
