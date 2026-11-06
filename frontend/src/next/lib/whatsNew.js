// What's new: one entry per release, newest first. To announce a release, add an entry on top with
// its pictures in backend/template/static/img/whats-new/ (see docs/whats-new/README.md).
//
// A slide with `permissions` is shown only to people who hold at least one of them; a slide with
// none is shown to everyone.

const IMAGES = "/static/img/whats-new/";

export const releases = [
  {
    id: "v0.3.0",
    title: "What's new in v0.3.0",
    slides: [
      {
        id: "stock-adjustments",
        title: "Correct stock without a full stock take",
        text: "Open Inventory → Stock Adjustments, pick the products and set, add or remove stock. Every change needs a reason and is recorded.",
        image: IMAGES + "v0.3.0-stock-adjustments.png",
        alt: "The New adjustment form: one product set from 36 to 9 in stock, answering a flag",
        permissions: ["stock:adjust"],
      },
      {
        id: "flag-product",
        title: "Flag a product that looks wrong",
        text: "Open a product and choose Flag for check, with the count you saw. A manager is told and can adjust the stock from your flag.",
        image: IMAGES + "v0.3.0-flag-product.png",
        alt: "A product's details with a note and the count seen, ready to send to a manager",
        permissions: ["stock:flag", "stock:adjust"],
      },
      {
        id: "purchase-orders",
        title: "Purchase orders with approval",
        text: "Write a draft and submit it. A manager approves or rejects it, and receiving against an order fills in what is still outstanding.",
        image: IMAGES + "v0.3.0-purchase-orders.png",
        alt: "A purchase order waiting for approval, with View, Edit, Reject and Approve buttons",
        permissions: ["purchaseorder:draft", "purchaseorder:approve"],
      },
      {
        id: "reprint-receipt",
        title: "Reprint a receipt",
        text: "In Sales History, open a sale and choose Reprint receipt. Preview it, then print a copy marked REPRINT. Returned items are listed.",
        image: IMAGES + "v0.3.0-reprint-receipt.png",
        alt: "A receipt preview marked REPRINT, with a returned item listed and a Print button",
      },
      {
        id: "add-user",
        title: "Add people in Administration",
        text: "Public sign-up is gone. Admins add users in Administration → Users → Add user, with a role and a starting password.",
        image: IMAGES + "v0.3.0-add-user.png",
        alt: "The Add user form with a username, a role and a starting password",
        permissions: ["admin:access"],
      },
    ],
  },
];

// visibleSlides returns the slides of a release that someone with hasPermission may see.
export function visibleSlides(release, hasPermission) {
  return release.slides.filter(
    (slide) => !slide.permissions?.length || slide.permissions.some((key) => hasPermission(key)),
  );
}
