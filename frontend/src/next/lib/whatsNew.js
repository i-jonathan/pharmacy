// What's new: one entry per release, newest first. To announce a release, add an entry on top with
// its pictures in backend/template/static/img/whats-new/ (see docs/whats-new/README.md).
//
// A slide with `permissions` is shown only to people who hold at least one of them; a slide with
// none is shown to everyone.

const IMAGES = "/static/img/whats-new/";

export const releases = [
  {
    id: "v0.6.0",
    title: "What's new in v0.6.0",
    slides: [
      {
        id: "health-check",
        title: "Record blood pressure and glucose",
        text: "Open Health checks, choose the customer or a walk-in, enter the reading and print a slip for them. A customer's readings are kept only if they agree; a walk-in's are printed and not kept.",
        image: IMAGES + "v0.6.0-health-check.png",
        alt: "The Health checks screen with a customer chosen, a blood pressure and a fasting glucose entered",
        permissions: ["healthcheck:record"],
      },
      {
        id: "health-slip",
        title: "A printed slip instead of a handwritten note",
        text: "The slip comes out of the receipt printer with the readings, the date and who took them. It never says whether a number is high or low; that is the pharmacist's call.",
        image: IMAGES + "v0.6.0-health-slip.png",
        alt: "A printed health check slip with a blood pressure, pulse and fasting glucose",
        permissions: ["healthcheck:record"],
      },
      {
        id: "health-history",
        title: "Look back on a customer's readings",
        text: "Managers open a customer and choose Health readings to see every saved reading, newest first. Reprint one as a marked copy, or correct a wrong number with a reason.",
        image: IMAGES + "v0.6.0-health-history.png",
        alt: "A customer's health readings, newest first, each with Reprint and Correct buttons",
        permissions: ["healthcheck:view"],
      },
      {
        id: "customers",
        title: "Keep a list of customers",
        text: "Customers holds the people you serve often: name, phone, email, notes and whether they agreed to be contacted. A phone number belongs to one customer.",
        image: IMAGES + "v0.6.0-customers.png",
        alt: "The Customers list with one customer's details open",
        permissions: ["customer:view"],
      },
      {
        id: "customer-at-till",
        title: "Put the customer on the sale",
        text: "At the till, choose Add customer and search by phone or name. The receipt shows their name, and Sales History shows who bought what.",
        image: IMAGES + "v0.6.0-customer-at-till.png",
        alt: "The current sale at the till with a customer attached",
        permissions: ["customer:view"],
      },
    ],
  },
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
