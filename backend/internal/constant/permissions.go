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
