// Retakes the What's new pictures against the demo database. See docs/whats-new/README.md.
//
//   node capture.mjs <output folder> [release]
//
// Takes the pictures of one release (default: the newest below), named <release>-<slide>.png.
// Needs `npm i playwright-core` next to this file, a Chromium (set CHROMIUM to its path), and the
// demo server running (BASE, default http://localhost:8123) on a database filled by
// `go run ./cmd/demoseed`. Light mode, 1100x700, one picture per slide.
import { chromium } from "playwright-core";
import { mkdirSync } from "node:fs";
import { join } from "node:path";

const BASE = process.env.BASE || "http://localhost:8123";
const EXE = process.env.CHROMIUM; // path to a Chromium or Chrome binary
const PASSWORD = "Demo-Pharmacy-2026"; // internal/demoseed.Password
const [out, release = "v0.6.0"] = process.argv.slice(2);
if (!out) throw new Error("usage: node capture.mjs <output folder> [release]");
mkdirSync(out, { recursive: true });

const browser = await chromium.launch({ executablePath: EXE, headless: true });

async function as(user) {
  const ctx = await browser.newContext({
    viewport: { width: 1100, height: 700 }, deviceScaleFactor: 1, colorScheme: "light",
  });
  const page = await ctx.newPage();
  await page.goto(BASE + "/");
  await page.locator('input[type="text"], input[name="username"]').first().fill(user);
  await page.locator('input[type="password"]').first().fill(PASSWORD);
  await Promise.all([page.waitForURL(/\/app\//), page.keyboard.press("Enter")]);
  await page.waitForLoadState("networkidle");
  return { ctx, page };
}

async function open(page, path) {
  await page.goto(BASE + path);
  await page.waitForLoadState("networkidle");
  await page.waitForTimeout(600);
}

async function shot(page, slug) {
  await page.waitForTimeout(300);
  await page.screenshot({ path: join(out, `${release}-${slug}.png`) });
  console.log("saved", `${release}-${slug}.png`);
}

// The slip as it prints: the slip's own page, shown at its 80mm width on a grey desk.
async function showSlip(page, html) {
  await page.setContent(`<body style="margin:0;background:#e5e7eb;display:flex;justify-content:center;
    align-items:flex-start;padding:24px 0;height:100vh;box-sizing:border-box">
    <iframe style="width:80mm;height:84mm;zoom:1.5;border:0;background:#fff;box-shadow:0 4px 18px #0002"></iframe>`);
  await page.locator("iframe").evaluate((f, h) => { f.srcdoc = h; }, html);
  await page.waitForTimeout(400);
}

const releases = {
  "v0.6.0": [
    async () => { // Health checks: the cashier fills in a reading for a regular
      const { ctx, page } = await as("tunde");
      await open(page, "/app/health-checks");
      await page.getByRole("button", { name: "Add customer" }).click();
      await page.getByPlaceholder("Search by phone or name").fill("Adaeze");
      await page.waitForTimeout(700);
      await page.getByRole("list", { name: "Matching customers" }).getByRole("button").first().click();
      const nums = page.locator("form input[type='text']");
      await nums.nth(0).fill("126");
      await nums.nth(1).fill("82");
      await nums.nth(2).fill("71");
      await page.getByText("Blood glucose", { exact: true }).click();
      await nums.nth(3).fill("5.6");
      await page.getByText("Fasting").click();
      await shot(page, "health-check");
      await ctx.close();
    },
    async () => { // The slip: a saved reading, printed (the print is caught and shown as a page)
      const { ctx, page } = await as("tunde");
      await page.addInitScript(() => {
        window.print = () => { window.top.__slip = document.documentElement.outerHTML; };
      });
      await open(page, "/app/health-checks");
      const nums = page.locator("form input[type='text']");
      await page.locator("input[maxlength='60']").fill("Mr Okon");
      await nums.nth(0).fill("126");
      await nums.nth(1).fill("82");
      await nums.nth(2).fill("71");
      await page.getByText("Blood glucose", { exact: true }).click();
      await nums.nth(3).fill("5.6");
      await page.getByText("Fasting").click();
      await page.locator("form").getByRole("button", { name: "Print" }).click();
      await page.waitForFunction(() => window.__slip);
      const html = await page.evaluate(() => window.__slip);
      await showSlip(page, html);
      await shot(page, "health-slip");
      await ctx.close();
    },
    async () => { // History: the manager looks back on a regular's readings
      const { ctx, page } = await as("amaka");
      await open(page, "/app/customers");
      await page.locator("tbody tr", { hasText: "Adaeze Obi" }).click();
      await page.getByRole("tab", { name: "Health readings" }).click();
      await page.waitForTimeout(800);
      await shot(page, "health-history");
      await ctx.close();
    },
    async () => { // Customers: the list with one open
      const { ctx, page } = await as("amaka");
      await open(page, "/app/customers");
      await page.locator("tbody tr", { hasText: "Chinedu Okafor" }).click();
      await page.waitForTimeout(500);
      await shot(page, "customers");
      await ctx.close();
    },
    async () => { // The till: a customer attached to the sale
      const { ctx, page } = await as("tunde");
      await open(page, "/app/pos");
      const search = page.getByPlaceholder("Search or scan barcode…");
      await search.fill("Paracetamol");
      await page.waitForTimeout(700);
      await page.locator("tbody tr", { hasText: "Paracetamol" }).getByRole("button").click();
      await page.waitForTimeout(400);
      // At 1100px the till shows one panel at a time: open the cart.
      await page.getByRole("button", { name: /^Cart/ }).click();
      await page.waitForTimeout(400);
      await page.getByRole("button", { name: "Add customer" }).click();
      await page.getByPlaceholder("Search by phone or name").fill("0803");
      await page.waitForTimeout(700);
      await page.getByRole("list", { name: "Matching customers" }).getByRole("button").first().click();
      await page.waitForTimeout(4000); // the "added to cart" toast goes
      await shot(page, "customer-at-till");
      await ctx.close();
    },
  ],
  "v0.3.0": [
  async () => { // Stock Adjustments: the manager adjusts from a flag
    const { ctx, page } = await as("amaka");
    await open(page, "/app/stock-adjustments");
    await page.getByRole("button", { name: "Adjust", exact: true }).click();
    await page.waitForTimeout(600);
    await page.locator('input[type="number"]').first().fill("9");
    await page.getByPlaceholder(/What happened/).fill("Recounted the shelf and the back room");
    await shot(page, "stock-adjustments");
    await ctx.close();
  },
  async () => { // Flag for check: the cashier raises it from a product
    const { ctx, page } = await as("tunde");
    await open(page, "/app/products");
    await page.getByText("Diclofenac Gel 30g").first().click();
    await page.getByRole("button", { name: "Flag for check" }).click();
    await page.getByPlaceholder(/two packs crushed/).fill("Shelf looks short");
    await page.locator("#flag-count").fill("9");
    await shot(page, "flag-product");
    await ctx.close();
  },
  async () => { // Purchase orders: waiting for the manager's approval
    const { ctx, page } = await as("amaka");
    await open(page, "/app/purchase-orders");
    await page.getByRole("button", { name: "Needs approval" }).click();
    await page.waitForTimeout(500);
    await shot(page, "purchase-orders");
    await ctx.close();
  },
  async () => { // Reprint receipt: the preview of a sale that had an item returned
    const { ctx, page } = await as("owner");
    await open(page, "/app/sales-history");
    await page.locator("tbody tr").last().click();
    await page.waitForTimeout(500);
    await page.getByRole("button", { name: "Reprint receipt" }).click();
    await page.waitForTimeout(600);
    await shot(page, "reprint-receipt");
    await ctx.close();
  },
  async () => { // Add user: an admin adds someone in Administration
    const { ctx, page } = await as("owner");
    await page.getByRole("button", { name: /Administration/ }).click();
    await page.getByText("Manage users, change roles").click();
    await page.getByRole("button", { name: "Add user" }).click();
    await page.waitForTimeout(600);
    await page.locator("#new-username").fill("chioma");
    await page.locator("select").filter({ hasText: "Choose a role" }).selectOption({ label: "Cashier" });
    await shot(page, "add-user");
    await ctx.close();
  },
  ],
};

const steps = releases[release];
if (!steps) throw new Error(`no pictures listed for ${release}`);
for (const step of steps) await step();
await browser.close();
