# Budgeting

This is a simple budgeting app. The goal is basic expense tracking over time. It is a single application using Go and HTMX.
The features I want to support are:
- Expense tracking in categories: simply add a name, amount, and optional date (time.Now if unset), and a category for the expense
- Subscriptions/recurring expenses: in monthly, 6-monthly, and annual cadences
- One "bucket" for everything with no balance: no multiple wallets/bank accounts/etc., just a bucket that tracks how much was spent
- Another "bucket" for all income: a simple way to add monthly income to compare to spending
- Over time, graphs to show the difference between how much was spent and how much was saved

Important:
- No authentication: this is designed to be run on a local computer or network only
- Data should go into a sqlite database
