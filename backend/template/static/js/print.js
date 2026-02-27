function printReceipt(sale) {
  const receiptWindow = window.open("", "PRINT", "height=700,width=400");

  const {
    items = [],
    subtotal = 0,
    total = 0,
    payments = [],
    date = new Date().toLocaleString(),
    discount = 0,
  } = sale;

  const itemRows = items
    .map(
      (item) => `
    <tr>
      <td>${item.name ?? item.product_name}</td>
      <td style="text-align:center;">${item.quantity}</td>
      <td style="text-align:right;">${(item.unit_price * item.quantity).toFixed(2)}</td>
    </tr>
  `,
    )
    .join("");

  const paymentRows = payments
    .map(
      (p) => `
    <tr>
      <td>${p.payment_method ?? p.method_name}</td>
      <td style="text-align:right;">${p.amount.toFixed(2)}</td>
    </tr>
  `,
    )
    .join("");

  const totalPaid = payments.reduce((sum, p) => sum + p.amount, 0);
  const change = totalPaid - total;

  const html = `
    <html>
    <head>
      <title>Receipt</title>
      <style>
        body {
          font-family: 'Courier New', monospace;
          width: 280px; /* use 280px for 58mm, ~550px for 80mm */
          margin: auto;
        }
        h2, h4, p {
          text-align: center;
          margin: 4px 0;
        }
        table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 8px;
        }
        td {
          padding: 2px 0;
          font-size: 13px;
        }
        .totals td {
          font-weight: bold;
          border-top: 1px dashed #000;
        }
        .heading td {
          font-weight: bold;
          border-bottom: 1px dashed #000;
        }
      </style>
    </head>
    <body>
      <h2>PrimoCrest Pharmacy</h2>
      <p>7, Chima-Abel Street, Ayobo, Lagos</p>
      <p>${date}</p>

      <table>
        <thead>
          <tr class="heading">
            <td>Item</td>
            <td style="text-align:center;">Qty</td>
            <td style="text-align:right;">Price</td>
          </tr>
        </thead>
        <tbody>
          ${itemRows}
        </tbody>
        <tfoot>
          <tr class="totals">
            <td colspan="2">Subtotal</td>
            <td style="text-align:right;">${subtotal.toFixed(2)}</td>
          </tr>
          
          ${discount > 0 ? `
            <tr class="totals">
              <td colspan="2">Discount</td>
              <td style="text-align:right;">${discount.toFixed(2)}</td>
            </tr>
          ` : ""}

          <tr class="totals">
            <td colspan="2">Total</td>
            <td style="text-align:right;">${total.toFixed(2)}</td>
          </tr>
        </tfoot>
      </table>

      <h4>Payments</h4>
      <table>
        <thead>
          <tr class="heading">
            <td>Method</td>
            <td style="text-align:right;">Amount</td>
          </tr>
        </thead>
        <tbody>
          ${paymentRows}
        </tbody>
        <tfoot>
          <tr class="totals">
            <td>Total Paid</td>
            <td style="text-align:right;">${totalPaid.toFixed(2)}</td>
          </tr>
          <tr class="totals">
            <td>Change</td>
            <td style="text-align:right;">${change.toFixed(2)}</td>
          </tr>
        </tfoot>
      </table>

      <h4>Thank you for your purchase!</h4>
    </body>
    </html>
  `;

  receiptWindow.document.write(html);
  receiptWindow.document.close();
  receiptWindow.focus();
  receiptWindow.print();
  receiptWindow.close();
}
