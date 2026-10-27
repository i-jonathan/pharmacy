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
