## Goal

- Taxes lets merchants manage tax definitions used in transactions (e.g., VAT/PPN) in one place

- Each tax is created per merchant with a unique code for easy identification

- Users can create, view, update, delete (soft delete), list, and toggle active status for taxes

- Listing supports search by name/code, filter by type and active status, with pagination

- Each tax stores key details: name, type, percentage, optional description, status, and audit info (created/updated by & time)

- Tax percentage must be between 0 and 100

- Deleting a tax does not remove it permanently; it is marked inactive and hidden from normal lists

- Error states are clear to users:
  - Invalid input or out-of-range percentage returns a validation error
  - Duplicate tax code returns a conflict error
  - Tax not found returns a not found error
  - Server/database errors return a server error

#### Definition Of Done

[x] Merchant can create a tax with a unique code

[x] Tax percentage is validated between 0 and 100

[x] Merchant can view, update, and soft delete a tax

[x] Merchant can list taxes with search, filter, and pagination

[x] Merchant can toggle tax active/inactive status

[x] Error responses cover invalid input, duplicates, not found, and server errors
