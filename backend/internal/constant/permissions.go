package constant

const ViewSalesTotalPermissionKey = "salestotal:view"
const CompleteStockTakingPermissionKey = "stock:complete"
const EditInventoryPermissionKey = "inventory:edit"
const AdminAccessPermissionKey = "admin:access"

// DraftPurchaseOrderPermissionKey lets a user write purchase order drafts and submit them for
// approval. ApprovePurchaseOrderPermissionKey is for managers: create orders directly, edit any
// draft, approve or reject, and close or move orders that are already open.
const DraftPurchaseOrderPermissionKey = "purchaseorder:draft"
const ApprovePurchaseOrderPermissionKey = "purchaseorder:approve"

// AdjustStockPermissionKey lets a user correct stock directly (managers by default).
// FlagStockPermissionKey lets a user flag a product for a manager to check.
const AdjustStockPermissionKey = "stock:adjust"
const FlagStockPermissionKey = "stock:flag"

// ViewCustomersPermissionKey lets a user see the customer list. ManageCustomersPermissionKey lets
// a user add, edit, deactivate and reactivate customers. Cashiers, managers and admins have both
// by default.
const ViewCustomersPermissionKey = "customer:view"
const ManageCustomersPermissionKey = "customer:manage"

// ChangeOldSaleCustomerPermissionKey lets a user change or remove the customer on a sale more
// than 24 hours old (managers and admins). Anyone with customer:manage may do it for newer sales.
const ChangeOldSaleCustomerPermissionKey = "sale:changeoldcustomer"

// OversellPermissionKey lets a user complete a sale that takes a product's stock below zero
// (managers by default). Every oversell is audited and opens a stock flag.
const OversellPermissionKey = "sale:oversell"

// RecordHealthCheckPermissionKey lets a user record a blood pressure or glucose reading and print
// its slip (cashiers, managers, admins). ViewHealthChecksPermissionKey lets a user see a
// customer's saved readings and reprint one, and CorrectHealthCheckPermissionKey lets them
// correct a saved reading (managers and admins).
const RecordHealthCheckPermissionKey = "healthcheck:record"
const ViewHealthChecksPermissionKey = "healthcheck:view"
const CorrectHealthCheckPermissionKey = "healthcheck:correct"
