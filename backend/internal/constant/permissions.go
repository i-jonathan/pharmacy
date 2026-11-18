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

// ChangeOldSaleCustomerPermissionKey lets a user change or remove the customer on a sale made on
// an earlier day (managers and admins). Anyone with customer:manage may do it for today's sales.
const ChangeOldSaleCustomerPermissionKey = "sale:changeoldcustomer"
