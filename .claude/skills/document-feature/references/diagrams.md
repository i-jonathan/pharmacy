# Diagrams in feature docs

Load before drawing any diagram.

## Which diagrams

- **Guide**: at most one, a simple `flowchart LR` of the main journey in the shop's words ("Draft order → Submit → Manager approves → Receive delivery"). Only when the feature has steps across people or days; a single-screen task needs none.
- **Technical**: exactly one overview `flowchart LR` right after the purpose (screen → route → service → tables and side effects). Add a `sequenceDiagram` for a flow only when it crosses three or more layers with a non-obvious order (locks, transactions, broadcasts), placed at the end of the section it explains.

## Rules

- No subgraphs, styles or colours.
- At most 12 nodes and 15 edges per diagram; split or drop it if it doesn't fit.
- Real names in the technical doc (`CreateSale`, `stock_movement`, `POST /sales/`); plain words in the guide.
- Path parameters as `:id`, never `{id}` (Mermaid reads braces as a node shape).
- Dotted links carry labels in pipes: `A -.->|label| B`. The `A -.label.-> B` form breaks on any `.` in the label.
- No `\n` inside labels.

## Check every diagram parses

Don't eyeball it. In a scratch directory:

```bash
npm i --silent mermaid@11 jsdom
cat > check.mjs <<'JS'
import fs from 'node:fs';
import { JSDOM } from 'jsdom';
const dom = new JSDOM('<!DOCTYPE html><body></body>', { pretendToBeVisual: true });
global.window = dom.window; global.document = dom.window.document;
Object.defineProperty(global, 'navigator', { value: dom.window.navigator, configurable: true });
const mermaid = (await import('mermaid')).default;
mermaid.initialize({ startOnLoad: false });
let bad = 0;
for (const file of process.argv.slice(2)) {
  const blocks = [...fs.readFileSync(file, 'utf8').matchAll(/```mermaid\n([\s\S]*?)```/g)].map(m => m[1]);
  for (const [i, b] of blocks.entries()) {
    try { await mermaid.parse(b); console.log(`${file} diagram ${i + 1}: OK`); }
    catch (e) { bad++; console.log(`${file} diagram ${i + 1}: FAIL — ${e.message}`); }
  }
}
process.exit(bad ? 1 : 0);
JS
node check.mjs /path/to/docs/features/<slug>/guide.md /path/to/docs/features/<slug>/technical.md
```
