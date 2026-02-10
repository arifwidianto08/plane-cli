## Goal

- Bank Account lets merchants manage their payment accounts (bank or cash) in one place for finance operations

- Accounts are created and managed per merchant, with a unique code so each account is easy to identify

- Two types are supported: `bank` and `cash`. For `bank`, account number and account name are required

- Users can create, view, update, delete (soft delete), list, and toggle active status for bank accounts

- Listing supports search by name/code, filter by type (bank/cash) and active status, with pagination

- Each record stores key details: bank/cash name, account number, account name, branch, notes, status, and audit info (created/updated by & time)

- Deleting an account does not remove it permanently; it is marked inactive and hidden from normal lists

- Error states are clear to users:
  - Invalid input or missing required fields returns a validation error
  - Duplicate account code returns a conflict error
  - Account not found returns a not found error
  - Server/database errors return a server error

#### Definition Of Done

[x] Merchant can create a bank/cash account with a unique code

[x] Bank type requires account number and account name

[x] Merchant can view, update, and soft delete an account

[x] Merchant can list accounts with search, filter, and pagination

[x] Merchant can toggle account active/inactive status

[x] Error responses cover invalid input, duplicates, not found, and server errors
