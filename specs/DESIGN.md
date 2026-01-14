# Budgeting App - Design Document

## Overview

A simple, local-first budgeting application for expense tracking and financial visualization. Built with Go backend and HTMX frontend for a responsive, server-rendered experience.

## Core Principles

- **Simplicity First**: No authentication, single bucket approach, minimal configuration
- **Local Operation**: Designed for personal use on local computer or trusted network
- **Server-Rendered UI**: HTMX for dynamic interactions without JavaScript complexity
- **SQLite Persistence**: Lightweight, file-based database for portability

## Architecture

### Tech Stack
- **Backend**: Go (standard library + SQLite driver)
- **Frontend**: HTMX for dynamic interactions, standard HTML/CSS
- **Database**: SQLite3
- **Charts**: Chart.js for interactive visualizations
- **Currency**: USD (amounts stored as integers in cents, displayed with formatting)

### Application Structure
```
budgeting/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── database/
│   │   ├── migrations/       # SQL migration files
│   │   ├── db.go            # Database connection & setup
│   │   └── queries.go       # SQL queries
│   ├── models/
│   │   ├── expense.go       # Expense model & logic
│   │   ├── income.go        # Income model & logic
│   │   ├── category.go      # Category model & logic
│   │   └── recurring.go     # Recurring transaction model
│   ├── handlers/
│   │   ├── dashboard.go     # Main dashboard handler
│   │   ├── expenses.go      # Expense CRUD handlers
│   │   ├── income.go        # Income CRUD handlers
│   │   ├── categories.go    # Category management handlers
│   │   └── recurring.go     # Recurring transaction handlers
│   └── scheduler/
│       └── scheduler.go     # Background job for recurring transactions
├── web/
│   ├── templates/
│   │   ├── layout.html      # Base layout
│   │   ├── dashboard.html   # Main dashboard
│   │   └── partials/        # HTMX partial templates
│   └── static/
│       ├── css/
│       └── js/
├── specs/
│   ├── bootstrap.md
│   └── DESIGN.md
└── go.mod
```

## Database Schema

### Tables

#### `categories`
```sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT,                    -- Hex color for UI display
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### `expenses`
```sql
CREATE TABLE expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    amount INTEGER NOT NULL,       -- Stored as cents (e.g., $12.34 = 1234)
    category_id INTEGER,
    expense_date DATE NOT NULL,    -- When the expense occurred
    notes TEXT,
    recurring_id INTEGER,          -- NULL for one-time, ID if auto-generated from recurring
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FOREIGN KEY (recurring_id) REFERENCES recurring_expenses(id) ON DELETE SET NULL
);

CREATE INDEX idx_expenses_date ON expenses(expense_date);
CREATE INDEX idx_expenses_category ON expenses(category_id);
```

#### `income`
```sql
CREATE TABLE income (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    amount INTEGER NOT NULL,       -- Stored as cents (e.g., $5000.00 = 500000)
    income_date DATE NOT NULL,
    notes TEXT,
    recurring_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (recurring_id) REFERENCES recurring_income(id) ON DELETE SET NULL
);

CREATE INDEX idx_income_date ON income(income_date);
```

#### `recurring_expenses`
```sql
CREATE TABLE recurring_expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    amount INTEGER NOT NULL,       -- Stored as cents
    category_id INTEGER,
    cadence TEXT NOT NULL,         -- 'monthly', 'semi-annual', 'annual'
    start_date DATE NOT NULL,      -- When to begin creating transactions
    next_date DATE NOT NULL,       -- Next scheduled transaction date
    end_date DATE,                 -- NULL for indefinite, or specific end date
    active BOOLEAN DEFAULT 1,      -- Can be paused without deletion
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE INDEX idx_recurring_expenses_next_date ON recurring_expenses(next_date);
```

#### `recurring_income`
```sql
CREATE TABLE recurring_income (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    amount INTEGER NOT NULL,       -- Stored as cents
    cadence TEXT NOT NULL,
    start_date DATE NOT NULL,
    next_date DATE NOT NULL,
    end_date DATE,
    active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recurring_income_next_date ON recurring_income(next_date);
```

## User Interface Design

### Dashboard Layout (Single Page)

```
┌─────────────────────────────────────────────────────────────┐
│  Budgeting App                                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │ Quick Add       │  │ Quick Add       │                 │
│  │ Expense         │  │ Income          │                 │
│  │                 │  │                 │                 │
│  │ [Name     ]     │  │ [Name     ]     │                 │
│  │ [Amount   ]     │  │ [Amount   ]     │                 │
│  │ [Category ▼]    │  │ [Date     ]     │                 │
│  │ [Date     ]     │  │                 │                 │
│  │     [Add]       │  │     [Add]       │                 │
│  └─────────────────┘  └─────────────────┘                 │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Overview - January 2026                            │  │
│  │  ───────────────────────────────────────────────────│  │
│  │  Total Income:    $5,000.00                         │  │
│  │  Total Expenses:  $3,245.67                         │  │
│  │  Net Savings:     $1,754.33                         │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Monthly View                 [Month ▼] [Year ▼]    │  │
│  │  ────────────────────────────────────────────────────│  │
│  │  [Bar chart showing income vs expenses by month]    │  │
│  │                                                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Yearly View                              [Year ▼]  │  │
│  │  ────────────────────────────────────────────────────│  │
│  │  [Line chart showing cumulative savings over year]  │  │
│  │                                                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Recent Transactions                    [View All]  │  │
│  │  ────────────────────────────────────────────────────│  │
│  │  2026-01-13  Groceries      $125.34  [Edit] [Del]   │  │
│  │  2026-01-12  Netflix         $15.99  [Edit] [Del]   │  │
│  │  2026-01-10  Electricity    $234.56  [Edit] [Del]   │  │
│  │  ...                                                 │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  [Manage Categories] [Manage Recurring Transactions]       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Additional Views (HTMX Modals/Overlays)

- **Category Management**: CRUD interface for categories
- **Recurring Transactions**: List and manage recurring expenses/income
- **Transaction Details**: Edit/view individual transactions
- **All Transactions**: Paginated list with filters

## Key Features & Behavior

### 1. Expense Tracking
- Quick-add form on dashboard for rapid entry
- Optional date defaults to current time if not specified
- Category selection from user-defined categories
- Edit/delete from recent transactions list

### 2. Recurring Transactions
- Background scheduler runs daily at midnight
- Automatically creates expense/income entries when `next_date == today` (exactly on due date)
- Updates `next_date` based on cadence:
  - Monthly: +1 month
  - Semi-annual: +6 months
  - Annual: +1 year
- Links auto-generated transactions to parent recurring record
- Can be paused (active=0) or deleted
- Respects `end_date` if set

### 3. Category Management
- User can create custom categories with name and optional color
- Categories can be edited or deleted
- Deleting a category sets category_id to NULL on existing expenses (not cascading delete)
- Default "Uncategorized" for expenses without category

### 4. Income Tracking
- Simple income entry with name, amount, and date
- Supports recurring income (e.g., monthly salary)
- Displayed separately from expenses in overview

### 5. Visualizations

#### Monthly View
- Bar chart showing income vs expenses per month
- Selectable month/year dropdown
- Shows net savings/deficit for selected month

#### Yearly View
- Line chart showing monthly totals across the year
- Year selector dropdown
- Displays trend over 12 months

### 6. Data Display
- Recent transactions list (last 10-20 entries)
- Combined expenses and income, sorted by date
- Quick edit/delete actions using HTMX for inline updates

## HTMX Interaction Patterns

### Quick Add Forms
- Form submission triggers HTMX POST
- Server responds with updated recent transactions partial
- Form clears on successful submission
- Error messages displayed inline

### Transaction Edit/Delete
- Edit button loads inline edit form via HTMX
- Delete button triggers confirmation, then HTMX DELETE request
- Updates affected UI sections (recent list, overview stats)

### Chart Updates
- Month/year selector triggers HTMX request
- Server re-renders chart with new data
- Swaps chart container content

### Modal Dialogs
- Category management and recurring transaction views
- Loaded via HTMX into overlay/modal container
- Forms within modals use HTMX for submission

## Recurring Transaction Scheduler

### Implementation
```go
// Runs as goroutine on app startup
func StartScheduler(db *sql.DB) {
    ticker := time.NewTicker(24 * time.Hour) // Daily check at midnight
    defer ticker.Stop()

    // Run immediately on startup
    processRecurringTransactions(db)

    for range ticker.C {
        processRecurringTransactions(db)
    }
}

func processRecurringTransactions(db *sql.DB) {
    today := time.Now().Format("2006-01-02")

    // 1. Find all active recurring expenses where next_date == today
    // 2. Create expense entry for each with expense_date = today
    // 3. Update next_date based on cadence (monthly +1 month, semi-annual +6 months, annual +1 year)
    // 4. Repeat for recurring income
    // 5. If end_date is set and new next_date > end_date, set active = 0
}
```

## API Endpoints (HTMX Targets)

### Dashboard
- `GET /` - Main dashboard page

### Expenses
- `POST /expenses` - Create new expense
- `GET /expenses/:id/edit` - Get edit form
- `PUT /expenses/:id` - Update expense
- `DELETE /expenses/:id` - Delete expense
- `GET /expenses/list` - Get paginated list (optional filter params)

### Income
- `POST /income` - Create new income
- `GET /income/:id/edit` - Get edit form
- `PUT /income/:id` - Update income
- `DELETE /income/:id` - Delete income

### Categories
- `GET /categories` - List all categories (modal view)
- `POST /categories` - Create category
- `PUT /categories/:id` - Update category
- `DELETE /categories/:id` - Delete category

### Recurring Transactions
- `GET /recurring` - List all recurring transactions (modal view)
- `POST /recurring/expenses` - Create recurring expense
- `POST /recurring/income` - Create recurring income
- `PUT /recurring/:type/:id` - Update recurring transaction
- `DELETE /recurring/:type/:id` - Delete recurring transaction
- `POST /recurring/:type/:id/pause` - Toggle active status

### Charts
- `GET /charts/monthly?month=X&year=Y` - Get monthly chart data/render
- `GET /charts/yearly?year=Y` - Get yearly chart data/render

### Partials (for HTMX swaps)
- `GET /partials/recent-transactions` - Recent transactions list
- `GET /partials/overview?month=X&year=Y` - Overview stats box

## Data Validation

### Input Validation Rules

#### Amounts (Expenses, Income, Recurring)
- **Must be positive**: All amounts must be > 0
- **Format**: Accept decimal input (e.g., "12.34") and convert to cents (1234) for storage
- **Display**: Format from cents to USD with two decimal places (e.g., 1234 → "$12.34")
- **Error message**: "Amount must be a positive number"

#### Names
- **Required**: Cannot be empty or whitespace-only
- **Max length**: 255 characters
- **Error message**: "Name is required"

#### Dates
- **Format**: ISO 8601 (YYYY-MM-DD)
- **Valid range**: Reasonable date range (e.g., 1900-01-01 to 2100-12-31)
- **Default**: Current date if not specified for expenses/income
- **Error message**: "Invalid date format"

#### Categories
- **Name required**: Cannot be empty
- **Unique names**: Category names must be unique (case-insensitive)
- **Color format**: Optional hex color (e.g., "#FF5733")
- **Error messages**: "Category name is required", "Category already exists"

#### Recurring Transactions
- **Cadence**: Must be one of: "monthly", "semi-annual", "annual"
- **Dates**: start_date <= next_date, end_date must be > start_date if set
- **All other validations**: Same as expenses/income (positive amount, required name, etc.)

### Server-Side Validation
- All validation performed server-side in Go handlers
- Return HTTP 400 with error messages for invalid input
- HTMX displays error messages inline in forms

### Currency Formatting Helpers

```go
// Convert dollars string to cents (integer)
// "12.34" → 1234, "5" → 500, "0.99" → 99
func DollarsToCents(dollars string) (int, error)

// Convert cents to formatted USD string
// 1234 → "$12.34", 500000 → "$5,000.00"
func CentsToUSD(cents int) string

// Parse and validate positive amount input
func ParseAmount(input string) (int, error)
```

## Development Phases

### Phase 1: Foundation
- Database schema and migrations
- Basic Go server setup with routing
- SQLite connection and basic CRUD operations

### Phase 2: Core Features
- Expense and income tracking (CRUD)
- Category management
- Dashboard layout and quick-add forms

### Phase 3: Recurring Transactions
- Recurring expense/income models
- Scheduler implementation
- Management UI

### Phase 4: Visualizations
- Monthly and yearly chart implementation
- Chart library integration
- Interactive month/year selection

### Phase 5: Polish
- CSS styling and responsive design
- Error handling and validation
- Data export functionality (optional future)

## Future Considerations

These are not in scope for initial version but could be added later:
- Budget limits and alerts
- Expense search and filtering
- Data export (CSV, PDF reports)
- Mobile-responsive improvements
- Dark mode
- Multi-currency support
