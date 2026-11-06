// Retakes the What's new pictures against the demo database. See docs/whats-new/README.md.
//
//   node capture.mjs <output folder> [prefix]
//
// Needs `npm i playwright-core` next to this file, a Chromium (set CHROMIUM to its path), and the
// demo server running (BASE, default http://localhost:8123) on a database filled by
// `go run ./cmd/demoseed`. Light mode, 1100x700, one picture per slide.
import { chromium } from "playwright-core";
import { mkdirSync } from "node:fs";
import { join } from "node:path";

const BASE = process.env.BASE || "http://localhost:8123";
const EXE = process.env.CHROMIUM; // path to a Chromium or Chrome binary
const PASSWORD = "Demo-Pharmacy-2026"; // internal/demoseed.Password
const [out, prefix = "v0.3.0"] = process.argv.slice(2);
if (!out) throw new Error("usage: node capture.mjs <output folder> [prefix]");
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
  await page.screenshot({ path: join(out, `${prefix}-${slug}.png`) });
  console.log("saved", `${prefix}-${slug}.png`);
}

{ // Stock Adjustments: the manager adjusts from a flag
  const { ctx, page } = await as("amaka");
  await open(page, "/app/stock-adjustments");
  await page.getByRole("button", { name: "Adjust", exact: true }).click();
  await page.waitForTimeout(600);
  await page.locator('input[type="number"]').first().fill("9");
  await page.getByPlaceholder(/What happened/).fill("Recounted the shelf and the back room");
  await shot(page, "stock-adjustments");
  await ctx.close();
}
{ // Flag for check: the cashier raises it from a product
  const { ctx, page } = await as("tunde");
  await open(page, "/app/products");
  await page.getByText("Diclofenac Gel 30g").first().click();
  await page.getByRole("button", { name: "Flag for check" }).click();
  await page.getByPlaceholder(/two packs crushed/).fill("Shelf looks short");
  await page.locator("#flag-count").fill("9");
  await shot(page, "flag-product");
  await ctx.close();
}
{ // Purchase orders: waiting for the manager's approval
  const { ctx, page } = await as("amaka");
  await open(page, "/app/purchase-orders");
  await page.getByRole("button", { name: "Needs approval" }).click();
  await page.waitForTimeout(500);
  await shot(page, "purchase-orders");
  await ctx.close();
}
{ // Reprint receipt: the preview of a sale that had an item returned
  const { ctx, page } = await as("owner");
  await open(page, "/app/sales-history");
  await page.locator("tbody tr").last().click();
  await page.waitForTimeout(500);
  await page.getByRole("button", { name: "Reprint receipt" }).click();
  await page.waitForTimeout(600);
  await shot(page, "reprint-receipt");
  await ctx.close();
}
{ // Add user: an admin adds someone in Administration
  const { ctx, page } = await as("owner");
  await page.getByRole("button", { name: /Administration/ }).click();
  await page.getByText("Manage users, change roles").click();
  await page.getByRole("button", { name: "Add user" }).click();
  await page.waitForTimeout(600);
  await page.locator("#new-username").fill("chioma");
  await page.locator("select").filter({ hasText: "Choose a role" }).selectOption({ label: "Cashier" });
  await shot(page, "add-user");
  await ctx.close();
}
await browser.close();
