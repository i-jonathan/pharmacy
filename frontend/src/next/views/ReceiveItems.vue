<template>
  <div class="p-6 lg:p-8">
    <!-- ===== DASHBOARD ===== -->
    <template v-if="view === 'dashboard'">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-2xl font-bold text-foreground">Receive Items</h1>
          <p class="text-sm text-muted-foreground mt-1">Record incoming inventory from suppliers</p>
        </div>
        <Button size="lg" @click="startNewReceipt"><Plus :size="16" class="mr-2" />New Receipt</Button>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div class="rounded-lg border border-border bg-card p-5"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-sky-100 dark:bg-sky-900/20 flex items-center justify-center"><CalendarCheck :size="20" class="text-sky-600" /></div><div><div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Received Today</div><div class="text-2xl font-bold text-foreground">{{ todayCount }}</div></div></div></div>
        <div class="rounded-lg border border-border bg-card p-5 cursor-pointer hover:bg-muted/20 transition-colors" @click="$router.push('/held-receive-items')"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/20 flex items-center justify-center"><PauseCircle :size="20" class="text-amber-600" /></div><div><div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Held Drafts</div><div class="text-2xl font-bold text-foreground">{{ heldCount > 0 ? heldCount : '—' }}</div></div></div></div>
        <div class="rounded-lg border border-border bg-card p-5"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-900/20 flex items-center justify-center"><Package :size="20" class="text-emerald-600" /></div><div><div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">This Month</div><div class="text-2xl font-bold text-foreground">{{ monthCount }}</div></div></div></div>
      </div>
      <!-- Filters -->
      <div class="flex flex-col sm:flex-row gap-3 mb-4">
        <div class="relative flex-1">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="historyQuery"
            type="text"
            placeholder="Search by supplier..."
            class="w-full pl-9 pr-4 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          />
        </div>
        <div class="flex gap-2 flex-wrap">
          <select
            v-model="rangePreset"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring"
            @change="onRangePreset"
          >
            <option value="today">Today</option>
            <option value="yesterday">Yesterday</option>
            <option value="this-week">This Week</option>
            <option value="last-week">Last Week</option>
            <option value="this-month">This Month</option>
            <option value="last-month">Last Month</option>
            <option value="all">All Time</option>
          </select>
          <input
            v-model="dateStart"
            type="date"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-36"
          />
          <input
            v-model="dateEnd"
            type="date"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-36"
          />
          <Button variant="outline" size="icon" :disabled="historyLoading" @click="fetchHistory">
            <RotateCw :size="16" :class="{ 'animate-spin': historyLoading }" />
          </Button>
        </div>
      </div>

      <!-- Summary bar -->
      <div class="flex items-center justify-between px-4 py-2 bg-muted/30 rounded-lg border border-border mb-4">
        <span class="text-xs text-muted-foreground">
          <template v-if="!historyLoading && !historyError">{{ historyCount }} receipt{{ historyCount !== 1 ? 's' : '' }}</template>
        </span>
        <router-link to="/received-items-history" class="text-xs text-primary hover:underline">Full history →</router-link>
      </div>

      <!-- History Table -->
      <div v-if="historyLoading" class="flex items-center justify-center py-12 text-muted-foreground">
        <RotateCw :size="20" class="animate-spin mr-3" />
        <span class="text-sm">Loading...</span>
      </div>

      <div v-else-if="historyError" class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card">
        <AlertCircle :size="32" class="text-destructive/40 mb-2" />
        <p class="text-sm text-muted-foreground">{{ historyError }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetchHistory">Retry</Button>
      </div>

      <div v-else-if="filteredHistory.length === 0" class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card">
        <ClipboardList :size="32" class="text-muted-foreground/40 mb-2" />
        <h3 class="text-sm font-semibold text-foreground mb-1">No receipts found</h3>
        <p class="text-xs text-muted-foreground">Start a new receipt or adjust your filters.</p>
      </div>

      <div v-else class="border border-border rounded-lg overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30 text-[11px]">
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2"></th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2">Supplier</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2">Date</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2">Received By</th>
              <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-16">Items</th>
              <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-28">Total Cost</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            <tr
              v-for="(b, i) in filteredHistory"
              :key="b.id"
              class="cursor-pointer hover:bg-muted/20 transition-colors"
              @click="receiptDetail = b"
            >
              <td class="px-3 py-2.5 text-center"><Truck :size="16" class="text-muted-foreground/60" /></td>
              <td class="px-3 py-2.5"><span class="text-sm font-medium text-foreground">{{ b.supplier_name }}</span></td>
              <td class="px-3 py-2.5 text-xs text-muted-foreground whitespace-nowrap">{{ formatDate(b.created_at) }}</td>
              <td class="px-3 py-2.5 text-xs text-muted-foreground">{{ b.received_by }}</td>
              <td class="px-3 py-2.5 text-center text-xs text-muted-foreground">{{ b.items?.length || 0 }}</td>
              <td class="px-3 py-2.5 text-right font-semibold text-sm text-foreground">&#8358;{{ historyTotal(b).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Receipt Detail Modal -->
      <Transition name="fade">
        <div v-if="receiptDetail" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="receiptDetail = null">
          <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-lg mx-4 p-6">
            <div class="flex items-center justify-between mb-4">
              <div>
                <h3 class="text-sm font-semibold">Receipt Details</h3>
                <p class="text-xs text-muted-foreground">{{ receiptDetail.supplier_name }} · {{ formatDate(receiptDetail.created_at) }} · {{ receiptDetail.received_by }}</p>
              </div>
              <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="receiptDetail = null"><X :size="15" /></Button>
            </div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead class="text-[10px]">Item</TableHead>
                  <TableHead class="w-10 text-[10px] text-center">Qty</TableHead>
                  <TableHead class="w-16 text-[10px] text-right">Cost</TableHead>
                  <TableHead class="w-16 text-[10px] text-right">Total</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in (receiptDetail.items || [])" :key="item.product_id">
                  <TableCell class="py-1.5">
                    <div class="text-xs font-medium">{{ item.product_name }}</div>
                    <div v-if="item.manufacturer" class="text-[10px] text-muted-foreground">{{ item.manufacturer }}</div>
                  </TableCell>
                  <TableCell class="py-1.5 text-center text-xs text-muted-foreground">{{ item.quantity }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs text-muted-foreground">&#8358;{{ Number(item.cost_price || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs font-medium">&#8358;{{ (Number(item.cost_price || 0) * Number(item.quantity || 0)).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </div>
      </Transition>
    </template>

    <!-- ===== NEW RECEIPT ===== -->
    <template v-if="view === 'receipt'">
      <div class="flex items-center gap-3 mb-6">
        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" @click="view = 'dashboard'"><ChevronLeft :size="16" /></Button>
        <div><h1 class="text-xl font-bold text-foreground">New Receipt</h1><p class="text-sm text-muted-foreground">Record incoming items from a supplier</p></div>
      </div>

      <!-- Step 1: Supplier -->
      <div class="rounded-lg border border-border bg-card p-5 mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">1</div>
          <span class="text-sm font-semibold text-foreground">Supplier</span>
          <span v-if="supplier" class="text-xs text-emerald-600 ml-2">✓ {{ supplier }}</span>
        </div>
        <div class="relative">
          <Building :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input v-model="supplier" type="text" placeholder="Enter supplier name..." ref="supplierInputRef" class="no-spinners w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring" :class="{ 'border-red-500': validationErrors.supplier }" @input="onSupplierInput" />
        </div>
        <ul v-if="supplierSuggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-hidden">
          <li v-for="s in supplierSuggestions" :key="s" class="flex items-center gap-2 px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" @click="selectSupplier(s)"><Building :size="14" class="text-muted-foreground" /><span>{{ s }}</span></li>
        </ul>
      </div>

      <!-- Step 2: Add Products -->
      <div class="rounded-lg border border-border bg-card p-5 mb-4">
        <div class="flex items-center justify-between gap-2 mb-3">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">2</div>
            <span class="text-sm font-semibold text-foreground">Add Products</span>
            <span v-if="items.length" class="text-xs text-muted-foreground ml-2">· {{ items.length }} item{{ items.length !== 1 ? 's' : '' }}</span>
          </div>
          <Button variant="outline" size="sm" @click="openNewProductModal"><Plus :size="13" class="mr-1" />New Product</Button>
        </div>
        <div class="relative">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input v-model="productQuery" type="text" placeholder="Search products by name or barcode..." ref="searchInputRef" class="no-spinners w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring" @input="onProductSearch" @keydown.escape="productSuggestions = []" />
        </div>
        <ul v-if="productSuggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-y-auto max-h-56">
          <li v-for="(p,i) in productSuggestions" :key="p.id" class="flex items-center justify-between px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" :class="i < productSuggestions.length - 1 ? 'border-b border-border/50' : ''" @click="addProduct(p)">
            <div class="flex items-center gap-2.5"><PillBottle :size="16" class="text-muted-foreground/60 shrink-0" /><div><div class="font-medium text-foreground">{{ p.name }}</div><div class="text-xs text-muted-foreground">{{ p.manufacturer || '—' }} · {{ p.barcode || 'no barcode' }}</div></div></div>
            <div class="text-xs text-muted-foreground font-mono">&#8358;{{ (p.default_price?.selling_price || 0).toLocaleString() }}</div>
          </li>
        </ul>
        <div v-if="!items.length && !productSuggestions.length" class="flex items-center justify-center py-8 text-center text-sm text-muted-foreground mt-2"><span class="text-xs">Search above or use <span class="font-medium text-primary">New Product</span> to add items.</span></div>

        <!-- New Product Modal -->
        <div v-if="showNewProductModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm overflow-y-auto" @click.self="showNewProductModal = false">
          <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-lg mx-4 p-6 my-8">
            <h3 class="text-base font-semibold mb-4">New Product</h3>
            <div class="space-y-3">
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Product Name <span class="text-destructive">*</span></label>
                <input v-model="newProduct.name" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" @input="checkDuplicateProduct" />
                <div v-if="newProduct.duplicateMsg" class="text-xs text-amber-600 mt-1">{{ newProduct.duplicateMsg }}</div>
              </div>
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Manufacturer <span class="text-destructive">*</span></label>
                <div class="relative">
                  <input v-model="newProduct.manufacturer" type="text" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" @input="onManufacturerInput" />
                  <ul v-if="manufacturerSuggestions.length" class="absolute z-10 w-full mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-hidden">
                    <li v-for="m in manufacturerSuggestions" :key="m" class="px-4 py-2 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" @click="selectManufacturer(m)">{{ m }}</li>
                  </ul>
                </div>
              </div>
              <div><label class="text-xs text-muted-foreground mb-1 block">Barcode</label><input v-model="newProduct.barcode" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              <div class="grid grid-cols-2 gap-3">
                <div><label class="text-xs text-muted-foreground mb-1 block">Cost Price (&#8358;) <span class="text-destructive">*</span></label><input v-model.number="newProduct.cost_price" type="number" step="0.01" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
                <div><label class="text-xs text-muted-foreground mb-1 block">Selling Price (&#8358;) <span class="text-destructive">*</span></label><input v-model.number="newProduct.selling_price" type="number" step="0.01" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div><label class="text-xs text-muted-foreground mb-1 block">Category</label>
                  <select v-model.number="newProduct.category_id" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring">
                    <option :value="0">Select...</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
                  </select>
                </div>
                <div><label class="text-xs text-muted-foreground mb-1 block">Reorder Level</label><input v-model.number="newProduct.reorder_level" type="number" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              </div>
            </div>
            <div class="flex items-center gap-2 mt-5 justify-end">
              <Button variant="outline" size="sm" @click="showNewProductModal = false">Cancel</Button>
              <Button size="sm" :disabled="newProductSaving" @click="saveNewProduct"><RotateCw v-if="newProductSaving" :size="14" class="animate-spin mr-1.5" /><Plus v-else :size="14" class="mr-1.5" />Create & Add</Button>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 3: Set Details -->
      <div v-if="items.length" class="rounded-lg border border-border bg-card p-5 mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">3</div>
          <span class="text-sm font-semibold text-foreground">Set Details</span>
          <span class="text-xs text-muted-foreground ml-2">Cost, price, quantity, and expiry</span>
        </div>
        <div class="rounded-lg border border-border overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30 text-[11px]">
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2">Item</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-44">Cost (&#8358;)</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-44">Sell (&#8358;)</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-24">Prices</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-24">Qty</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-36">Expiry</th>
                <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-28">Total</th>
                <th class="w-10"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <template v-for="(item, idx) in items" :key="item._key">
                <tr class="hover:bg-muted/20 transition-colors">
                  <td class="px-3 py-2.5">
                    <div class="text-sm font-medium text-foreground">{{ item.name }}</div>
                    <div class="text-[11px] text-muted-foreground">{{ item.manufacturer || '' }}</div>
                  </td>
                  <td class="px-3 py-2.5">
                      <input :value="item.cost_price" @input="item.cost_price = num($event.target.value); suggestPrice(item)" type="number" step="0.01" min="0" class="no-spinners w-full px-2 py-1.5 text-sm border border-border rounded-md bg-background outline-none font-medium" :class="{ 'border-red-500': item._errors?.cost }" />
                  </td>
                  <td class="px-3 py-2.5">
                      <input :value="item.selling_price" @input="item.selling_price = num($event.target.value); suggestPrice(item)" type="number" step="0.01" min="0" class="no-spinners w-full px-2 py-1.5 text-sm border border-border rounded-md bg-background outline-none font-medium" :class="{ 'border-red-500': item._errors?.sell || (num(item.selling_price) > 0 && num(item.selling_price) <= num(item.cost_price)) }" :title="(num(item.selling_price) > 0 && num(item.selling_price) <= num(item.cost_price)) ? 'Selling price should be higher than cost price' : ''" />
                      <button v-if="item._suggestedPrice !== undefined" class="text-[10px] text-primary/70 hover:text-primary cursor-pointer" @click="item.selling_price = item._suggestedPrice; item._suggestedPrice = undefined">Suggested: &#8358;{{ item._suggestedPrice }}</button>
                  </td>
                  <td class="px-2 py-2 text-center">
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:text-primary rounded-lg" title="Price options" @click="openPriceOptions(item)"><Settings2 :size="15" /></Button>
                  </td>
                  <td class="px-3 py-2.5"><div class="inline-flex items-center border border-border rounded-md overflow-hidden"><button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-sm transition-colors" @click="item.quantity = Math.max(0, (item.quantity || 0) - 1)">−</button><input :value="item.quantity" @input="item.quantity = Math.max(0, num($event.target.value))" class="no-spinners h-7 w-9 text-center text-sm bg-transparent border-x border-border outline-none" /><button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-sm transition-colors" @click="item.quantity = (item.quantity || 0) + 1">+</button></div></td>
                  <td class="px-3 py-2.5"><input :value="item.expiry" @input="item.expiry = $event.target.value" type="date" class="no-spinners w-full px-2 py-1.5 text-sm text-center border border-border rounded-md bg-background outline-none" :class="{ 'border-red-500': item._errors?.expiry }" /></td>
                  <td class="px-2 py-2 text-right font-semibold text-sm text-foreground">&#8358;{{ (num(item.cost_price||0) * num(item.quantity||0)).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</td>
                  <td class="px-3 py-2.5"><Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-destructive" @click="removeItem(idx)"><X :size="13" /></Button></td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Step 4: Review Totals -->
      <div v-if="items.length" class="border border-border bg-card rounded-lg p-5 mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">4</div>
          <span class="text-sm font-semibold text-foreground">Review Totals</span>
        </div>
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="text-sm space-y-1.5"><div class="flex items-center gap-2"><Package :size="16" class="text-muted-foreground shrink-0" /><span>{{ items.length }} item{{ items.length !== 1 ? 's' : '' }}</span></div><div class="flex items-center gap-2"><span class="font-semibold">&#8358;{{ items.reduce((s,i) => s + num(i.cost_price||0) * num(i.quantity||0), 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</span><span class="text-muted-foreground">total cost</span></div></div>
          <div class="flex items-center gap-2"><Button variant="outline" size="sm" :disabled="submitting" @click="holdReceipt"><PauseCircle :size="14" class="mr-1.5" />Hold Draft</Button><Button size="lg" class="px-6 gap-2" :disabled="submitting" @click="receiveItems"><CircleCheck :size="16" />Receive Items</Button></div>
        </div>
        <div v-if="validationErrors.global" class="text-xs text-destructive mt-1.5 flex items-center gap-1"><AlertTriangle :size="12" />{{ validationErrors.global }}</div>
      </div>

      <!-- Price Options Modal -->
      <div v-if="priceOptionsTarget" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="closePriceOptions">
        <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold">Price Options</h3>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closePriceOptions"><X :size="15" /></Button>
          </div>
          <p class="text-xs text-muted-foreground mb-4">Set alternative pricing for different pack sizes or formulations.</p>
          <div v-for="(po, pi) in priceOptionsTarget._priceOptions" :key="pi" class="flex items-center gap-2 text-xs border border-border/50 rounded-md px-3 py-2 mb-2">
            <input :value="po.name" @input="po.name = $event.target.value" class="no-spinners w-24 px-2 py-1.5 border border-border rounded-md bg-background outline-none" placeholder="Name" />
            <input :value="po.price" @input="po.price = num($event.target.value)" type="number" step="0.01" class="no-spinners w-22 px-2 py-1.5 text-right border border-border rounded-md bg-background outline-none" placeholder="Price" />
            <input :value="po.qty" @input="po.qty = num($event.target.value)" type="number" class="no-spinners w-16 px-2 py-1.5 text-right border border-border rounded-md bg-background outline-none" placeholder="Qty" />
            <Button variant="ghost" size="icon" class="h-6 w-6 text-muted-foreground hover:text-destructive" @click="priceOptionsTarget._priceOptions.splice(pi, 1)"><X :size="11" /></Button>
          </div>
          <div class="flex items-center justify-between mt-2">
            <Button variant="outline" size="sm" @click="priceOptionsTarget._priceOptions.push({ id: null, name: '', price: 0, qty: 1 })"><Plus :size="12" class="mr-1" />Add Option</Button>
            <Button variant="outline" size="sm" @click="closePriceOptions">Done</Button>
          </div>
        </div>
      </div>
    </template>

    <!-- Nav Guard Modal -->
    <Transition name="fade">
      <div v-if="showNavGuardModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="cancelDiscard">
        <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-sm mx-4 p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/20 flex items-center justify-center shrink-0">
              <AlertTriangle :size="20" class="text-amber-600" />
            </div>
            <div>
              <h3 class="text-sm font-semibold">Unsaved Receipt</h3>
              <p class="text-xs text-muted-foreground">You have items that haven't been received yet.</p>
            </div>
          </div>
          <p class="text-sm text-foreground mb-5">Navigate away and discard these items, or cancel and hold them first.</p>
          <div class="flex items-center gap-2 justify-end">
            <Button variant="outline" size="sm" @click="cancelDiscard">Cancel</Button>
            <Button size="sm" class="gap-2" @click="confirmDiscard">
              <X :size="13" class="shrink-0" />
              Discard
            </Button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 flex items-center gap-2.5 bg-card border border-border px-4 py-3 rounded-xl shadow-2xl text-sm font-medium"><CircleCheck v-if="toastType === 'success'" :size="16" class="text-emerald-600" /><AlertCircle v-else :size="16" class="text-destructive" />{{ toast }}</div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { Search, X, Building, Package, PillBottle, PauseCircle, CircleCheck, Plus, AlertTriangle, AlertCircle, ChevronLeft, CalendarCheck, ClipboardList, Truck, RotateCw, Settings2 } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const API = "";

const VIEW_DASHBOARD = "dashboard";
const VIEW_RECEIPT = "receipt";

const view = ref(VIEW_DASHBOARD);
const supplier = ref("");
const supplierSuggestions = ref([]);
const productQuery = ref("");
const productSuggestions = ref([]);
const showNewProductModal = ref(false);
const newProductSaving = ref(false);
const newProduct = ref({ name: "", manufacturer: "", barcode: "", selling_price: 0, cost_price: 0, category_id: 0, reorder_level: 1, duplicateMsg: "" });
const categories = ref([]);
const manufacturerSuggestions = ref([]);
const items = ref([]);
const heldReference = ref("");
const recentReceipts = ref([]);
const todayCount = ref(0);
const monthCount = ref(0);
const heldCount = ref(0);
const receiptDetail = ref(null);
const historyQuery = ref("");
const historyBatches = ref([]);
const historyLoading = ref(false);
const historyError = ref(null);
const rangePreset = ref("today");
const dateStart = ref("");
const dateEnd = ref("");
const submitting = ref(false);
const validationErrors = ref({ supplier: false, global: "" });
const toast = ref(null);
const toastType = ref("success");
const priceOptionsTarget = ref(null);
const searchInputRef = ref(null);
const showNavGuardModal = ref(false);
const pendingNav = ref(null);

let tTimer = null; let keyCounter = 0; let msTimer = null;

function num(v) { return Number(v || 0); }
function showToast(msg, type = "success") { toast.value = msg; toastType.value = type; clearTimeout(tTimer); tTimer = setTimeout(() => { toast.value = null; }, 3000); }

function suggestPrice(item) {
  if (num(item.cost_price) > 0) {
    // Match old UI: cost × 1.3, round UP to nearest 50
    const suggested = Math.ceil(num(item.cost_price) * 1.3 / 50) * 50;
    // If no selling price yet, auto-fill with suggested
    if (!num(item.selling_price)) {
      item.selling_price = suggested;
      item._suggestedPrice = undefined;
    } else {
      // Show clickable suggestion when it differs from current selling price
      item._suggestedPrice = (num(item.selling_price) !== suggested) ? suggested : undefined;
    }
  }
}

// === Dashboard ===
async function fetchDashboard() {
  let allBatches = [];
  try {
    const r = await fetch(`${API}/inventory/received-items-history/api`);
    if (r.ok) { const d = await r.json(); allBatches = d.batches || []; }
  } catch {}
  recentReceipts.value = allBatches.slice(0, 5);
  // Count today's receipts and this month's receipts from the full list
  const now = new Date();
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).toISOString();
  const monthStart = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
  todayCount.value = allBatches.filter(b => b.created_at >= todayStart).length;
  monthCount.value = allBatches.filter(b => b.created_at >= monthStart).length;
  // Fetch held drafts count
  try {
    const h = await fetch(`${API}/inventory/receive-items/held/api`);
    if (h.ok) { const d = await h.json(); heldCount.value = d.length || 0; }
  } catch {}
}

function formatDate(d) { if (!d) return ""; try { return new Date(d).toLocaleDateString(); } catch { return ""; } }

// === History table helpers ===
const filteredHistory = computed(() => {
  if (!historyQuery.value) return historyBatches.value;
  const q = historyQuery.value.toLowerCase();
  return historyBatches.value.filter(
    (b) => b.supplier_name?.toLowerCase().includes(q) || b.received_by?.toLowerCase().includes(q)
  );
});

const historyCount = computed(() => filteredHistory.value.length);

function historyTotal(batch) {
  if (!batch?.items) return 0;
  return batch.items.reduce((sum, item) => sum + Number(item.cost_price || 0) * Number(item.quantity || 0), 0);
}

function getDateRange(preset) {
  const now = new Date();
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, "0");
  const d = String(now.getDate()).padStart(2, "0");
  const today = `${y}-${m}-${d}`;
  switch (preset) {
    case "today": return { start: today, end: today };
    case "yesterday": { const yest = new Date(now); yest.setDate(yest.getDate() - 1); return { start: yest.toISOString().slice(0, 10), end: yest.toISOString().slice(0, 10) }; }
    case "this-week": { const sun = new Date(now); sun.setDate(now.getDate() - now.getDay()); return { start: sun.toISOString().slice(0, 10), end: today }; }
    case "last-week": { const s = new Date(now); s.setDate(now.getDate() - now.getDay() - 7); const e = new Date(s); e.setDate(s.getDate() + 6); return { start: s.toISOString().slice(0, 10), end: e.toISOString().slice(0, 10) }; }
    case "this-month": return { start: `${y}-${m}-01`, end: today };
    case "last-month": { const lm = new Date(now.getFullYear(), now.getMonth() - 1, 1); const lme = new Date(now.getFullYear(), now.getMonth(), 0); return { start: lm.toISOString().slice(0, 10), end: lme.toISOString().slice(0, 10) }; }
    default: return { start: "", end: "" };
  }
}

function onRangePreset() {
  const { start, end } = getDateRange(rangePreset.value);
  dateStart.value = start;
  dateEnd.value = end;
  fetchHistory();
}

async function fetchHistory() {
  historyLoading.value = true;
  historyError.value = null;
  try {
    const params = new URLSearchParams();
    if (dateStart.value) params.set("start", dateStart.value);
    if (dateEnd.value) params.set("end", dateEnd.value);
    const res = await fetch(`${API}/inventory/received-items-history/api?${params}`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    historyBatches.value = data.batches || [];
  } catch (e) {
    historyError.value = e.message || "Failed to load";
  } finally {
    historyLoading.value = false;
  }
}

// === Navigation ===
function startNewReceipt() {
  view.value = VIEW_RECEIPT; supplier.value = ""; items.value.splice(0);
  productQuery.value = ""; validationErrors.value = { supplier: false, global: "" }; fetchCategories();
  heldReference.value = "";
}

// === Categories ===
async function fetchCategories() {
  try {
    const r = await fetch(`${API}/inventory/item-list`);
    if (r.ok) { const d = await r.json(); categories.value = d.categories || []; }
  } catch {}
}

// === Supplier search (starts from 1 char) ===
function onSupplierInput() { validationErrors.value.supplier = false; debouncedSupplierSearch(); }
function selectSupplier(s) { supplier.value = s; supplierSuggestions.value = []; validationErrors.value.supplier = false; }
let ssTimer = null;
function debouncedSupplierSearch() {
  const q = supplier.value.trim();
  if (q.length < 1) { supplierSuggestions.value = []; return; }
  clearTimeout(ssTimer);
  ssTimer = setTimeout(async () => {
    try { const r = await fetch(`${API}/inventory/suppliers/search?query=${encodeURIComponent(q)}`); if (r.ok) { const d = await r.json(); supplierSuggestions.value = d.data || []; } } catch {}
  }, 200);
}

// === Product Search ===
function onProductSearch() { debouncedProductSearch(); }
let psTimer = null;
function debouncedProductSearch() {
  const q = productQuery.value.trim();
  if (q.length < 2) { productSuggestions.value = []; return; }
  clearTimeout(psTimer);
  psTimer = setTimeout(async () => {
    try { const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`); if (r.ok) productSuggestions.value = await r.json(); } catch {}
  }, 250);
}

// === Add Product ===
function addProduct(p) {
  productQuery.value = ""; productSuggestions.value = []; keyCounter++;
  const cost = p.cost_price || (p.default_price?.selling_price ? p.default_price.selling_price * 0.7 : 0);
  const sell = p.default_price?.selling_price || 0;
  const costBasedSuggested = cost > 0 ? Math.ceil(cost * 1.3 / 50) * 50 : 0;
  const suggestedSell = sell > 0 ? sell : costBasedSuggested;
  const priceOpts = (p.price_options || []).map((po) => ({ id: po.id, name: po.name, price: po.selling_price, qty: po.quantity_per_unit || 1 }));
  items.value.push({
    _key: p.id + "-" + keyCounter, id: p.id, name: p.name, manufacturer: p.manufacturer || "",
    barcode: p.barcode || "", cost_price: cost, selling_price: suggestedSell, quantity: 1, expiry: "",
    _errors: {}, _priceOptions: priceOpts,
    // Show suggestion if using default sell price and cost-based differs (matching old UI)
    _suggestedPrice: (sell > 0 && costBasedSuggested > 0 && costBasedSuggested !== suggestedSell) ? costBasedSuggested : undefined,
  });
  refocusSearch();
}
function removeItem(idx) { items.value.splice(idx, 1); }
function refocusSearch() { setTimeout(() => { const el = searchInputRef.value; if (el) el.focus(); }, 50); }

// === Price Options Modal ===
function openPriceOptions(item) { priceOptionsTarget.value = item; }
function closePriceOptions() { priceOptionsTarget.value = null; }

// === New Product Modal ===
function openNewProductModal() {
  productSuggestions.value = []; showNewProductModal.value = true;
  newProduct.value = { name: "", manufacturer: "", barcode: "", selling_price: 0, cost_price: 0, category_id: 0, reorder_level: 1, duplicateMsg: "" };
  manufacturerSuggestions.value = [];
}

// === Manufacturer autocomplete ===
function onManufacturerInput() {
  const q = newProduct.value.manufacturer.trim();
  if (q.length < 1) { manufacturerSuggestions.value = []; return; }
  clearTimeout(msTimer);
  msTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`);
      if (r.ok) {
        const products = await r.json();
        const seen = new Set();
        manufacturerSuggestions.value = products.filter(p => p.manufacturer).map(p => p.manufacturer).filter(m => { if (seen.has(m)) return false; seen.add(m); return true; }).slice(0, 8);
      }
    } catch {}
    checkDuplicateProduct();
  }, 200);
}
function selectManufacturer(m) {
  newProduct.value.manufacturer = m; manufacturerSuggestions.value = [];
  checkDuplicateProduct();
}

function checkDuplicateProduct() {
  const name = newProduct.value.name.trim();
  const mfr = newProduct.value.manufacturer.trim();
  if (name.length < 2 || mfr.length < 2) { newProduct.value.duplicateMsg = ""; return; }
  fetch(`${API}/inventory/search?query=${encodeURIComponent(name)}`)
    .then(r => r.ok ? r.json() : [])
    .then(products => {
      const dup = products.find(p => p.name?.toLowerCase() === name.toLowerCase() && p.manufacturer?.toLowerCase() === mfr.toLowerCase());
      newProduct.value.duplicateMsg = dup ? '⚠ "' + dup.name + '" by ' + dup.manufacturer + ' already exists' : "";
    }).catch(() => {});
}

async function saveNewProduct() {
  if (!newProduct.value.name.trim()) { showToast("Product name is required", "error"); return; }
  if (!newProduct.value.manufacturer.trim()) { showToast("Manufacturer is required", "error"); return; }
  if (num(newProduct.value.category_id) <= 0) { showToast("Category is required", "error"); return; }
  if (num(newProduct.value.cost_price) <= 0) { showToast("Cost price is required", "error"); return; }
  if (num(newProduct.value.selling_price) <= 0) { showToast("Selling price is required", "error"); return; }
  if (newProduct.value.duplicateMsg) { showToast("A product with this name and manufacturer already exists", "error"); return; }
  newProductSaving.value = true;
  try {
    const r = await fetch(`${API}/inventory/add-item`, {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: newProduct.value.name.trim(), manufacturer: newProduct.value.manufacturer.trim(),
        barcode: newProduct.value.barcode.trim(), category_id: num(newProduct.value.category_id),
        reorder_level: num(newProduct.value.reorder_level) || 1, cost_price: num(newProduct.value.cost_price),
        selling_price: num(newProduct.value.selling_price),
      }),
    });
    if (!r.ok) throw new Error("Failed to create product");
    const created = await r.json();
    showNewProductModal.value = false;
    const suggestedSell = num(newProduct.value.selling_price) || Math.ceil(num(newProduct.value.cost_price) * 1.3 / 50) * 50;
    items.value.push({
      _key: "new-" + keyCounter++, id: created.id, name: newProduct.value.name.trim(),
      manufacturer: newProduct.value.manufacturer.trim(), barcode: newProduct.value.barcode.trim(),
      cost_price: num(newProduct.value.cost_price), selling_price: suggestedSell, quantity: 1, expiry: "",
      _errors: {}, _priceOptions: [], _suggestedPrice: undefined,
    });
    showToast("Product created and added"); refocusSearch();
  } catch (e) { showToast(e.message || "Failed to create product", "error"); }
  finally { newProductSaving.value = false; }
}

// === Hold (no validation) ===
async function holdReceipt() {
  submitting.value = true;
  try {
    const p = { reference: heldReference.value, payload: { supplier: supplier.value.trim(), products: items.value.map(i => ({ id: i.id, name: i.name, manufacturer: i.manufacturer, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: (i._priceOptions || []).map(po => ({ id: po.id, name: po.name, selling_price: po.price, quantity_per_unit: po.qty || 1 })) })) } };
    const r = await fetch(`${API}/inventory/receive-items/hold`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) throw new Error("Failed to hold");
    view.value = VIEW_DASHBOARD; showToast("Receipt saved as draft"); fetchDashboard();
  } catch (e) { showToast(e.message || "Failed to hold", "error"); }
  finally { submitting.value = false; }
}

// === Validate + Receive ===
function validate() {
  let valid = true; validationErrors.value = { supplier: false, global: "" };
  if (!supplier.value.trim()) { validationErrors.value.supplier = true; valid = false; }
  if (!items.value.length) { validationErrors.value.global = "Add at least one product."; valid = false; }
  for (const item of items.value) {
    item._errors = {};
    if (num(item.cost_price) <= 0) { item._errors.cost = true; valid = false; }
    if (num(item.selling_price) <= 0) { item._errors.sell = true; valid = false; }
    if (num(item.quantity) <= 0) { item._errors.qty = true; valid = false; }
    if (!item.expiry) { item._errors.expiry = true; valid = false; }
  }
  if (!valid && !validationErrors.value.global) validationErrors.value.global = "Fill in all required fields marked in red.";
  return valid;
}

async function receiveItems() {
  if (!validate()) return;
  submitting.value = true;
  try {
    const p = { supplier: supplier.value.trim(), held_receiving_reference: heldReference.value, idempotency_key: (() => { try { return crypto.randomUUID(); } catch { return Date.now() + "-" + Math.random().toString(36).slice(2); } })(), products: items.value.map(i => ({ id: i.id, barcode: i.barcode || "", cost_price: num(i.cost_price), selling_price: num(i.selling_price), quantity: Math.max(1, num(i.quantity)), expiry: i.expiry && i.expiry.trim() ? i.expiry.split("T")[0] + "T00:00:00Z" : new Date().toISOString().split("T")[0] + "T00:00:00Z", price_options_changes: (i._priceOptions || []).map(po => ({ id: po.id, name: po.name || "", selling_price: num(po.price), quantity_per_unit: Math.max(1, po.qty || 1) })) })) };
    const r = await fetch(`${API}/inventory/receive-items`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) { const e = await r.json().catch(() => ({})); throw new Error(e.error || `HTTP ${r.status}`); }
    view.value = VIEW_DASHBOARD; showToast("Items received successfully"); fetchDashboard();
  } catch (e) { showToast(e.message || "Failed to receive", "error"); }
  finally { submitting.value = false; }
}

function onBeforeUnload(e) {
  if (view.value === VIEW_RECEIPT && items.value.length > 0) {
    e.preventDefault();
  }
}

const router = useRouter();
let navConfirmed = false;

const navGuard = router.beforeEach((to) => {
  if (navConfirmed) { navConfirmed = false; return true; }
  if (view.value === VIEW_RECEIPT && items.value.length > 0 && to.name !== undefined) {
    pendingNav.value = to;
    showNavGuardModal.value = true;
    return false;
  }
  return true;
});

function confirmDiscard() {
  showNavGuardModal.value = false;
  const target = pendingNav.value;
  pendingNav.value = null;
  if (target) { navConfirmed = true; router.push(target); }
}

function cancelDiscard() {
  showNavGuardModal.value = false;
  pendingNav.value = null;
}

function restoreHeldData() {
  const data = localStorage.getItem("heldReceiveItems");
  if (!data) return;
  localStorage.removeItem("heldReceiveItems");
  try {
    const parsed = JSON.parse(data);
    heldReference.value = parsed.reference || "";
    const payload = parsed.payload || {};
    if (payload.supplier) supplier.value = payload.supplier;
    (payload.products || []).forEach((p) => {
      keyCounter++;
      const priceOpts = (p.price_options_changes || []).map((po) => ({
        id: po.id, name: po.name, price: po.selling_price, qty: po.quantity_per_unit || 1,
      }));
      items.value.push({
        _key: (p.id || "restored") + "-" + keyCounter,
        id: p.id, name: p.name, manufacturer: p.manufacturer || "",
        barcode: p.barcode || "", cost_price: Number(p.cost_price || 0),
        selling_price: Number(p.selling_price || 0), quantity: Number(p.quantity || 1),
        expiry: (p.expiry || "").split("T")[0] || "",
        _errors: {}, _priceOptions: priceOpts, _suggestedPrice: undefined,
      });
    });
    if (items.value.length) {
      view.value = VIEW_RECEIPT;
      showToast("Held receipt restored");
    }
  } catch (e) {
    console.error("Failed to restore held data:", e);
  }
}

onMounted(() => {
  fetchDashboard(); fetchCategories();
  window.addEventListener("beforeunload", onBeforeUnload);
  restoreHeldData();
  onRangePreset();
});

onUnmounted(() => {
  window.removeEventListener("beforeunload", onBeforeUnload);
  navGuard(); // remove the global navigation guard so other pages aren't blocked
});
</script>

<style scoped>
input.no-spinners::-webkit-outer-spin-button,
input.no-spinners::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
input.no-spinners[type="number"] { -moz-appearance: textfield; }
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>