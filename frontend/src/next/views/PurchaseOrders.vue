<template>
  <div v-if="!canManage" class="p-4 sm:p-6 lg:p-8">
    <h1 class="text-2xl font-semibold tracking-tight">Purchase Orders</h1>
    <p class="mt-2 text-sm text-muted-foreground">
      You do not have permission to manage purchase orders. Ask an administrator for access.
    </p>
  </div>
  <div v-else class="space-y-5 p-4 sm:p-6 lg:p-8">
    <!-- Header -->
    <header class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Purchase Orders</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Plan what to buy, send it to your supplier, and track what arrives.
        </p>
      </div>
      <Button v-if="!editing" @click="startNew()">
        <Plus :size="16" class="mr-2" />
        New purchase order
      </Button>
    </header>

    <p
      v-if="notice"
      class="rounded-md border px-3 py-2 text-sm"
      :class="notice.error ? 'border-destructive/30 bg-destructive/5 text-destructive' : 'border-border bg-muted/40'"
      role="status"
    >
      {{ notice.text }}
    </p>

    <datalist id="po-suppliers">
      <option v-for="name in supplierSuggestions" :key="name" :value="name" />
    </datalist>

    <!-- ===== EDITOR ===== -->
    <section v-if="editing" class="space-y-4 rounded-lg border border-border bg-card p-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h2 class="flex items-center gap-2 text-lg font-semibold">
          {{ editId ? `Edit ${orderNumber(editId)}` : "New purchase order" }}
          <span v-if="editId" :class="[badgeBase, STATUS[editStatus].badge]">{{ editStatus }}</span>
        </h2>
        <Button variant="outline" size="sm" @click="cancelEdit">
          <X :size="14" class="mr-1" />
          Cancel
        </Button>
      </div>

      <div
        v-if="conflict"
        class="flex flex-wrap items-center justify-between gap-2 rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950/30 dark:text-amber-200"
        role="alert"
      >
        <span>This order was changed by someone else (or received against) since you opened it. Your edits were not saved.</span>
        <Button size="sm" @click="reloadLatest">Reload latest</Button>
      </div>

      <div class="grid gap-3 md:grid-cols-2">
        <label class="block text-sm">
          Supplier <span class="text-muted-foreground">(optional now, required when receiving)</span>
          <input
            v-model="supplier"
            list="po-suppliers"
            class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2"
            placeholder="Supplier name"
            @input="searchSuppliers($event.target.value)"
          />
        </label>
        <label class="block text-sm">
          Additional notes <span class="text-muted-foreground">(optional)</span>
          <textarea
            v-model="note"
            rows="2"
            maxlength="1000"
            class="mt-1 w-full resize-y rounded-md border border-border bg-background px-3 py-2"
            placeholder="Delivery instructions, payment terms, anything the supplier or receiver should know"
          ></textarea>
        </label>
      </div>

      <!-- Add items -->
      <div class="space-y-3 rounded-md border border-dashed border-border p-3">
        <div class="text-sm font-medium">Add items</div>

        <div class="relative">
          <input
            v-model="query"
            class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            placeholder="Search products you already stock"
            aria-label="Search products"
            @input="searchProducts"
          />
          <div
            v-if="results.length"
            class="absolute z-10 mt-1 max-h-64 w-full overflow-y-auto rounded-md border border-border bg-card shadow-lg"
          >
            <button
              v-for="p in results"
              :key="p.id"
              type="button"
              class="flex w-full items-center justify-between gap-3 border-b border-border p-3 text-left text-sm last:border-0 hover:bg-muted/40"
              @click="addProduct(p)"
            >
              <span>
                {{ p.name }}
                <small class="text-muted-foreground">{{ p.manufacturer }}</small>
              </span>
              <span class="text-xs text-muted-foreground">{{ onOrderLabel(p.id) }}</span>
            </button>
          </div>
        </div>

        <div class="flex flex-wrap gap-2">
          <Button variant="outline" size="sm" @click="toggleLowStock">
            <AlertTriangle :size="14" class="mr-2 text-amber-600" />
            {{ lowStockOpen ? "Hide low stock" : "Choose from low stock" }}
          </Button>
        </div>

        <!-- Low stock picker -->
        <div v-if="lowStockOpen" class="rounded-md border border-border">
          <div v-if="lowStockLoading" class="p-4 text-sm text-muted-foreground">Loading low stock items...</div>
          <div v-else-if="!lowStockRows.length" class="p-4 text-sm text-muted-foreground">
            Nothing is at or below its reorder level.
          </div>
          <template v-else>
            <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border bg-muted/30 px-3 py-2 text-xs">
              <span class="text-muted-foreground">
                {{ lowStockRows.length }} low stock · tick the ones you want to order
              </span>
              <span class="flex gap-3">
                <button type="button" class="text-primary hover:underline" @click="selectAllLowStock">Select all</button>
                <button type="button" class="text-primary hover:underline" @click="clearLowStock">Clear</button>
              </span>
            </div>
            <div class="max-h-72 divide-y divide-border overflow-y-auto">
              <div
                v-for="row in lowStockRows"
                :key="row.id"
                class="flex items-center gap-3 px-3 py-2 text-sm"
                :class="{ 'opacity-60': row.added }"
              >
                <input
                  type="checkbox"
                  class="h-4 w-4"
                  :checked="row.id in lowStockPick"
                  :disabled="row.added"
                  :aria-label="`Select ${row.name}`"
                  @change="toggleLowStockRow(row, $event.target.checked)"
                />
                <div class="min-w-0 flex-1">
                  <div class="truncate">{{ row.name }} <small class="text-muted-foreground">{{ row.manufacturer }}</small></div>
                  <div class="text-xs text-muted-foreground">
                    In stock {{ row.stock }} · reorder at {{ row.reorder_level }}
                    <span v-if="row.added" class="ml-1 text-primary">· already on this order</span>
                    <span v-else-if="row.onOrder" class="ml-1 text-amber-600 dark:text-amber-400">
                      · {{ row.onOrder }} already on order
                    </span>
                  </div>
                </div>
                <QuantityStepper
                  v-if="row.id in lowStockPick"
                  v-model="lowStockPick[row.id]"
                  :label="`quantity for ${row.name}`"
                />
              </div>
            </div>
            <div class="flex justify-end border-t border-border p-2">
              <Button size="sm" :disabled="!pickedCount" @click="addPickedLowStock">
                Add {{ pickedCount || "" }} selected
              </Button>
            </div>
          </template>
        </div>

        <!-- Item not in the product list -->
        <div class="space-y-1">
          <div class="text-xs text-muted-foreground">Not in your products yet? Add it by name. You will register it when it arrives.</div>
          <div class="flex flex-wrap items-center gap-2">
            <input
              v-model="newItem.name"
              class="min-w-0 flex-1 basis-48 rounded-md border border-border bg-background px-3 py-2 text-sm"
              placeholder="New item name"
              aria-label="New item name"
              @keydown.enter.prevent="addNewItem"
            />
            <input
              v-model="newItem.manufacturer"
              class="min-w-0 flex-1 basis-36 rounded-md border border-border bg-background px-3 py-2 text-sm"
              placeholder="Manufacturer (optional)"
              aria-label="New item manufacturer"
              @keydown.enter.prevent="addNewItem"
            />
            <QuantityStepper v-model="newItem.quantity" label="new item quantity" />
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <input
              v-model="newItem.note"
              maxlength="500"
              class="min-w-0 flex-1 basis-48 rounded-md border border-border bg-background px-3 py-2 text-sm"
              placeholder="Note for this item (optional)"
              aria-label="New item note"
              @keydown.enter.prevent="addNewItem"
            />
            <Button variant="outline" :disabled="addingNew" @click="addNewItem">Add item</Button>
          </div>
        </div>
      </div>

      <!-- Lines -->
      <div v-if="lines.length" class="space-y-2">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="text-sm font-medium">
            {{ lines.length }} item{{ lines.length === 1 ? "" : "s" }} · {{ lineUnits }} units
          </div>
          <input
            v-if="lines.length > LINES_PER_PAGE"
            v-model="lineFilter"
            class="w-56 rounded-md border border-border bg-background px-3 py-1.5 text-sm"
            placeholder="Filter items"
            aria-label="Filter items"
          />
        </div>
        <div class="overflow-x-auto rounded-md border border-border">
          <table class="w-full text-sm">
            <thead class="bg-muted/30 text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="p-2 text-left">Item</th>
                <th v-if="hasReceived" class="w-24 p-2 text-center">Received</th>
                <th class="w-40 p-2 text-center">Ordered</th>
                <th class="w-24 p-2"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="line in pagedLines" :key="line.key" class="border-t border-border">
                <td class="p-2">
                  {{ line.name }}
                  <span v-if="line.manufacturer" class="ml-2 text-xs text-muted-foreground">{{ line.manufacturer }}</span>
                  <span
                    v-if="!line.product_id"
                    class="ml-2 rounded-full border border-amber-300 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300"
                  >
                    New product
                  </span>
                  <input
                    v-if="line.note || line.showNote"
                    v-model="line.note"
                    maxlength="500"
                    class="mt-1 block w-full rounded border border-border bg-background px-2 py-1 text-xs"
                    placeholder="Note for this item (brand, strength, pack size...)"
                    :aria-label="`Note for ${line.name}`"
                  />
                  <button
                    v-else
                    type="button"
                    class="mt-0.5 block text-xs text-primary hover:underline"
                    @click="line.showNote = true"
                  >
                    + Add note
                  </button>
                </td>
                <td v-if="hasReceived" class="p-2 text-center tabular-nums text-muted-foreground">{{ line.received_quantity }}</td>
                <td class="p-2 text-center">
                  <QuantityStepper
                    v-model="line.quantity"
                    :min="Math.max(1, line.received_quantity)"
                    :label="`quantity for ${line.name}`"
                  />
                </td>
                <td class="p-2 text-right">
                  <button
                    type="button"
                    class="text-red-600 hover:underline disabled:cursor-not-allowed disabled:no-underline disabled:opacity-40 dark:text-red-400"
                    :disabled="line.received_quantity > 0"
                    :title="line.received_quantity > 0 ? 'Stock has already been received for this item' : ''"
                    @click="removeLine(line.key)"
                  >
                    Remove
                  </button>
                </td>
              </tr>
              <tr v-if="!pagedLines.length">
                <td colspan="4" class="p-4 text-center text-sm text-muted-foreground">No items match your filter.</td>
              </tr>
            </tbody>
          </table>
        </div>
        <PagerBar v-model:page="linePage" :total="filteredLines.length" :size="LINES_PER_PAGE" />
      </div>
      <p v-else class="text-sm text-muted-foreground">No items yet. Search, pick from low stock, or add a new item above.</p>

      <div class="flex flex-wrap justify-end gap-2">
        <Button variant="outline" :disabled="!lines.length" @click="exportDraft">
          <Download :size="16" class="mr-2" />
          Export PDF
        </Button>
        <Button :disabled="saving || !lines.length" @click="save">
          {{ saving ? "Saving..." : editId ? "Save changes" : "Save purchase order" }}
        </Button>
      </div>
    </section>

    <!-- ===== ORDERS ===== -->
    <section v-if="!editing" class="space-y-3">
      <div class="flex gap-1 border-b border-border">
        <button
          v-for="t in tabs"
          :key="t.key"
          type="button"
          class="-mb-px border-b-2 px-4 py-2 text-sm font-medium"
          :class="tab === t.key ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'"
          @click="switchTab(t.key)"
        >
          {{ t.label }}
        </button>
      </div>

      <div v-if="loading" class="py-12 text-center text-sm text-muted-foreground">Loading purchase orders...</div>
      <div v-else-if="loadError" class="py-12 text-center text-sm">
        <p class="mb-3 text-muted-foreground">{{ loadError }}</p>
        <Button variant="outline" size="sm" @click="loadOrders">Retry</Button>
      </div>
      <div
        v-else-if="!orders.length"
        class="rounded-lg border border-border bg-card p-8 text-center text-sm text-muted-foreground"
      >
        {{ tab === "active" ? "No purchase orders waiting for stock." : "No completed purchase orders yet." }}
      </div>

      <template v-else>
        <article
          v-for="o in pagedOrders"
          :key="o.id"
          class="rounded-lg border border-l-4 border-border bg-card p-4"
          :class="STATUS[o.status].accent"
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0 flex-1 basis-72">
              <div class="flex flex-wrap items-center gap-2 font-medium">
                {{ orderNumber(o.id) }}
                <span :class="[badgeBase, STATUS[o.status].badge]">{{ o.status }}</span>
                <span class="font-normal text-muted-foreground">{{ o.supplier || "No supplier" }}</span>
              </div>
              <div class="mt-1 text-sm text-muted-foreground">
                {{ o.items.length }} item{{ o.items.length === 1 ? "" : "s" }} · {{ formatDate(o.created_at) }}
              </div>
              <div class="mt-2 flex items-center gap-3">
                <div class="h-1.5 w-40 overflow-hidden rounded-full bg-muted">
                  <div class="h-full rounded-full" :class="STATUS[o.status].bar" :style="{ width: progress(o) + '%' }"></div>
                </div>
                <span class="text-xs tabular-nums text-muted-foreground">
                  {{ totals(o).received }} of {{ totals(o).ordered }} units received
                </span>
              </div>
              <p v-if="o.note" class="mt-2 line-clamp-2 text-sm text-muted-foreground">
                <span class="font-medium text-foreground">Note:</span> {{ o.note }}
              </p>
              <div v-if="o.source_purchase_order_id || o.moved_to_purchase_order_id" class="mt-1 text-xs text-muted-foreground">
                <span v-if="o.source_purchase_order_id">Outstanding items moved here from {{ orderNumber(o.source_purchase_order_id) }}. </span>
                <span v-if="o.moved_to_purchase_order_id">Outstanding items moved to {{ orderNumber(o.moved_to_purchase_order_id) }}.</span>
              </div>
            </div>

            <div v-if="closeTarget === o.id" class="w-full space-y-2 text-sm md:w-auto md:max-w-md">
              <span class="block text-muted-foreground">
                {{ totals(o).received ? `Stop waiting for the remaining ${totals(o).outstanding} units?` : "Cancel this order?" }}
                Say why, so it is on record.
              </span>
              <ReasonPicker v-model:preset="closeReason.preset" v-model:detail="closeReason.detail" />
              <div class="flex gap-2">
                <Button variant="destructive" size="sm" :disabled="closing || !composeReason(closeReason)" @click="closeOrder(o)">
                  {{ closing ? "Closing..." : totals(o).received ? "Yes, close" : "Yes, cancel" }}
                </Button>
                <Button variant="outline" size="sm" :disabled="closing" @click="closeTarget = null">Keep open</Button>
              </div>
            </div>
            <div v-else class="flex flex-wrap gap-2">
              <Button variant="outline" size="sm" @click="openDetails(o.id)">
                <Eye :size="14" class="mr-1" />
                View
              </Button>
              <Button v-if="isEditable(o)" variant="outline" size="sm" @click="startEdit(o)">Edit</Button>
              <Button v-if="isEditable(o)" size="sm" @click="receive(o)">Receive</Button>
              <Button variant="outline" size="sm" @click="exportOrder(o)">
                <Download :size="14" class="mr-1" />
                PDF
              </Button>
              <Button v-if="isEditable(o) && totals(o).outstanding" variant="outline" size="sm" @click="startMove(o)">
                Other supplier
              </Button>
              <Button
                v-if="isEditable(o)"
                variant="outline"
                size="sm"
                class="border-red-300 text-red-600 hover:bg-red-50 hover:text-red-700 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950/40"
                @click="startClose(o)"
              >
                {{ totals(o).received ? "Close order" : "Cancel order" }}
              </Button>
            </div>
          </div>

          <div v-if="moveTarget === o.id" class="mt-3 space-y-2 rounded-md border border-dashed border-border p-3 text-sm">
            <p class="text-muted-foreground">
              Send the {{ totals(o).outstanding }} outstanding units to a different supplier. This order is closed
              with what it already received, and a new order is created for them.
            </p>
            <input
              v-model="moveSupplier"
              list="po-suppliers"
              class="w-full rounded-md border border-border bg-background px-3 py-2 md:max-w-sm"
              placeholder="New supplier name"
              aria-label="New supplier"
              @input="searchSuppliers($event.target.value)"
            />
            <ReasonPicker v-model:preset="moveReason.preset" v-model:detail="moveReason.detail" />
            <div class="flex gap-2">
              <Button size="sm" :disabled="moving || !moveSupplier.trim() || !composeReason(moveReason)" @click="moveOutstanding(o)">
                {{ moving ? "Moving..." : "Move outstanding" }}
              </Button>
              <Button variant="outline" size="sm" :disabled="moving" @click="moveTarget = null">Cancel</Button>
            </div>
          </div>
        </article>
        <PagerBar v-model:page="ordersPage" :total="orders.length" :size="ORDERS_PER_PAGE" />
      </template>
    </section>

    <!-- ===== ORDER DETAILS ===== -->
    <div
      v-if="detailOrder"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="closeDetails"
    >
      <section
        class="flex max-h-[85vh] w-full max-w-2xl flex-col rounded-xl border border-border bg-card shadow-2xl"
        role="dialog"
        aria-modal="true"
        :aria-label="`${orderNumber(detailOrder.id)} items`"
      >
        <header class="space-y-2 border-b border-border p-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex flex-wrap items-center gap-2 text-lg font-semibold">
                {{ orderNumber(detailOrder.id) }}
                <span :class="[badgeBase, STATUS[detailOrder.status].badge]">{{ detailOrder.status }}</span>
              </div>
              <div class="text-sm text-muted-foreground">
                {{ detailOrder.supplier || "No supplier" }} · {{ formatDate(detailOrder.created_at) }}
              </div>
            </div>
            <button type="button" class="rounded-md p-1 text-muted-foreground hover:bg-accent" aria-label="Close" @click="closeDetails">
              <X :size="18" />
            </button>
          </div>
          <div class="flex items-center gap-3">
            <div class="h-2 flex-1 overflow-hidden rounded-full bg-muted">
              <div class="h-full rounded-full" :class="STATUS[detailOrder.status].bar" :style="{ width: progress(detailOrder) + '%' }"></div>
            </div>
            <span class="text-xs tabular-nums text-muted-foreground">
              {{ totals(detailOrder).received }} / {{ totals(detailOrder).ordered }} units
            </span>
          </div>
          <p v-if="detailOrder.note" class="whitespace-pre-line rounded-md bg-muted/40 p-2 text-sm">{{ detailOrder.note }}</p>
        </header>

        <div class="flex flex-wrap items-center justify-between gap-2 px-4 pt-3">
          <div class="flex gap-1">
            <button
              v-for="f in detailFilters"
              :key="f.key"
              type="button"
              class="rounded-full border px-3 py-1 text-xs font-medium"
              :class="detailFilter === f.key ? 'border-primary bg-primary text-primary-foreground' : 'border-border text-muted-foreground hover:bg-accent'"
              @click="detailFilter = f.key"
            >
              {{ f.label }} ({{ f.count(detailOrder) }})
            </button>
          </div>
          <input
            v-model="detailQuery"
            class="w-48 rounded-md border border-border bg-background px-3 py-1.5 text-sm"
            placeholder="Search items"
            aria-label="Search items"
          />
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto px-4 py-3">
          <table class="w-full text-sm">
            <thead class="sticky top-0 bg-card text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="py-1 text-left">Item</th>
                <th class="w-20 py-1 text-center">Ordered</th>
                <th class="w-20 py-1 text-center">Received</th>
                <th class="w-24 py-1 text-center">Outstanding</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in pagedDetailItems" :key="item.id" class="border-t border-border">
                <td class="py-1.5 pr-2">
                  {{ item.name }}
                  <span v-if="item.manufacturer" class="ml-2 text-xs text-muted-foreground">{{ item.manufacturer }}</span>
                  <span
                    v-if="!item.product_id"
                    class="ml-2 rounded-full border border-amber-300 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300"
                  >
                    New product
                  </span>
                  <div v-if="item.note" class="text-xs italic text-muted-foreground">{{ item.note }}</div>
                </td>
                <td class="py-1.5 text-center tabular-nums">{{ item.quantity }}</td>
                <td class="py-1.5 text-center tabular-nums">{{ item.received_quantity }}</td>
                <td
                  class="py-1.5 text-center tabular-nums"
                  :class="item.quantity === item.received_quantity ? 'text-emerald-600 dark:text-emerald-400' : detailOrder.status === 'closed' ? 'text-muted-foreground' : 'font-medium text-amber-600 dark:text-amber-400'"
                >
                  {{ item.quantity === item.received_quantity ? "Done" : item.quantity - item.received_quantity }}
                </td>
              </tr>
              <tr v-if="!pagedDetailItems.length">
                <td colspan="4" class="py-6 text-center text-muted-foreground">No items match.</td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer class="space-y-2 border-t border-border p-3">
          <PagerBar v-model:page="detailPage" :total="filteredDetailItems.length" :size="DETAIL_PER_PAGE" />
          <div v-if="showActivity" class="max-h-48 space-y-2 overflow-y-auto rounded-md border border-border p-2 text-xs">
            <div v-for="e in history" :key="e.id">
              <div class="font-medium capitalize">
                {{ e.action }}
                <span class="font-normal text-muted-foreground">· {{ e.actor || "unknown" }} · {{ new Date(e.occurred_at).toLocaleString() }}</span>
              </div>
              <div v-for="(line, i) in eventLines(e)" :key="i" class="text-muted-foreground">{{ line }}</div>
              <div v-if="e.reason" class="italic">Reason: {{ e.reason }}</div>
            </div>
            <div v-if="!history.length" class="text-muted-foreground">No activity recorded.</div>
          </div>
          <div class="flex justify-end gap-2">
            <Button variant="outline" size="sm" @click="showActivity = !showActivity">
              Activity ({{ history.length }})
            </Button>
            <Button variant="outline" size="sm" @click="exportOrder(detailOrder)">
              <Download :size="14" class="mr-1" />
              PDF
            </Button>
            <Button variant="outline" size="sm" @click="closeDetails">Close</Button>
          </div>
        </footer>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { AlertTriangle, Download, Eye, Plus, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import QuantityStepper from "../components/QuantityStepper.vue";
import PagerBar from "../components/PagerBar.vue";
import ReasonPicker from "../components/ReasonPicker.vue";
import { csrfFetch } from "../lib/csrf.js";
import { usePermissions } from "../composables/usePermissions.js";
import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";

const API = "";
const { hasPermission } = usePermissions();
const canManage = computed(() => hasPermission("purchaseorder:create"));
const route = useRoute();
const router = useRouter();

const LINES_PER_PAGE = 8;
const DETAIL_PER_PAGE = 10;
const ORDERS_PER_PAGE = 10;

const tabs = [
  { key: "active", label: "Waiting for stock" },
  { key: "history", label: "Completed" },
];

// Each status keeps one colour everywhere it appears: badge, card accent and progress bar.
const badgeBase = "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold capitalize";
const STATUS = {
  open: {
    badge: "border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950/40 dark:text-blue-300",
    accent: "border-l-blue-500",
    bar: "bg-blue-500",
  },
  partial: {
    badge: "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300",
    accent: "border-l-amber-500",
    bar: "bg-amber-500",
  },
  received: {
    badge: "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300",
    accent: "border-l-emerald-500",
    bar: "bg-emerald-500",
  },
  closed: {
    badge: "border-slate-300 bg-slate-100 text-slate-600 dark:border-slate-700 dark:bg-slate-800/60 dark:text-slate-300",
    accent: "border-l-slate-400",
    bar: "bg-slate-400",
  },
};

// --- list state ---
const tab = ref("active");
const orders = ref([]);
const ordersPage = ref(1);
const loading = ref(false);
const loadError = ref("");
const closeTarget = ref(null);
const closing = ref(false);
const moveTarget = ref(null);
const moveSupplier = ref("");
const moving = ref(false);
const closeReason = reactive({ preset: "", detail: "" });
const moveReason = reactive({ preset: "", detail: "" });
const notice = ref(null);
let noticeTimer = null;

// --- details dialog state ---
const detailId = ref(null);
const detailQuery = ref("");
const detailFilter = ref("all");
const detailPage = ref(1);
const history = ref([]);
const showActivity = ref(false);

// --- editor state ---
const editing = ref(false);
const editId = ref(0);
const editStatus = ref("open");
const editVersion = ref(0);
const conflict = ref(false);
const supplier = ref("");
const note = ref("");
const lines = ref([]);
const lineFilter = ref("");
const linePage = ref(1);
const saving = ref(false);
const savedSnapshot = ref("");
let lineKey = 0;

const query = ref("");
const results = ref([]);
const supplierSuggestions = ref([]);
const newItem = reactive({ name: "", manufacturer: "", note: "", quantity: 1 });
const addingNew = ref(false);
let searchTimer = null;
let supplierTimer = null;

// --- low stock picker ---
const lowStockOpen = ref(false);
const lowStockLoading = ref(false);
const lowStockItems = ref([]);
const outstandingByProduct = ref({});
const lowStockPick = reactive({});

const hasReceived = computed(() => lines.value.some((l) => l.received_quantity > 0));
const lineUnits = computed(() => lines.value.reduce((sum, l) => sum + (Number(l.quantity) || 0), 0));
const pickedCount = computed(() => Object.keys(lowStockPick).length);
const isDirty = computed(() => editing.value && snapshot() !== savedSnapshot.value);

const lowStockRows = computed(() =>
  lowStockItems.value.map((p) => ({
    ...p,
    onOrder: outstandingByProduct.value[p.id] || 0,
    added: lines.value.some((l) => l.product_id === p.id),
  })),
);

const filteredLines = computed(() => {
  const q = lineFilter.value.trim().toLowerCase();
  if (!q) return lines.value;
  return lines.value.filter((l) => l.name.toLowerCase().includes(q) || l.manufacturer.toLowerCase().includes(q));
});
const pagedLines = computed(() => page(filteredLines.value, linePage.value, LINES_PER_PAGE));
const pagedOrders = computed(() => page(orders.value, ordersPage.value, ORDERS_PER_PAGE));

const detailOrder = computed(() => orders.value.find((o) => o.id === detailId.value) || null);
const detailFilters = [
  { key: "all", label: "All", count: (o) => o.items.length },
  { key: "outstanding", label: "Outstanding", count: (o) => o.items.filter((i) => i.received_quantity < i.quantity).length },
  { key: "done", label: "Complete", count: (o) => o.items.filter((i) => i.received_quantity >= i.quantity).length },
];
const filteredDetailItems = computed(() => {
  if (!detailOrder.value) return [];
  const q = detailQuery.value.trim().toLowerCase();
  return detailOrder.value.items.filter((i) => {
    if (detailFilter.value === "outstanding" && i.received_quantity >= i.quantity) return false;
    if (detailFilter.value === "done" && i.received_quantity < i.quantity) return false;
    return !q || i.name.toLowerCase().includes(q) || (i.manufacturer || "").toLowerCase().includes(q);
  });
});
const pagedDetailItems = computed(() => page(filteredDetailItems.value, detailPage.value, DETAIL_PER_PAGE));

// A filter, tab or edit can leave the current page past the end; step back onto the last one.
function clampPage(pageRef, total, size) {
  const last = Math.max(1, Math.ceil(total / size));
  if (pageRef.value > last) pageRef.value = last;
}
watch(() => filteredLines.value.length, (n) => clampPage(linePage, n, LINES_PER_PAGE));
watch(() => orders.value.length, (n) => clampPage(ordersPage, n, ORDERS_PER_PAGE));
watch(lineFilter, () => { linePage.value = 1; });
watch([detailQuery, detailFilter], () => { detailPage.value = 1; });

function page(list, current, size) {
  return list.slice((current - 1) * size, current * size);
}

function orderNumber(id) {
  return `PO-${String(id).padStart(5, "0")}`;
}

function formatDate(d) {
  try {
    return new Date(d).toLocaleDateString();
  } catch {
    return "";
  }
}

function isEditable(o) {
  return o.status === "open" || o.status === "partial";
}

function totals(o) {
  const ordered = o.items.reduce((sum, i) => sum + i.quantity, 0);
  const received = o.items.reduce((sum, i) => sum + i.received_quantity, 0);
  return { ordered, received, outstanding: ordered - received };
}

function progress(o) {
  const { ordered, received } = totals(o);
  return ordered ? Math.round((received / ordered) * 100) : 0;
}

function showNotice(text, error = false) {
  notice.value = { text, error };
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => { notice.value = null; }, error ? 8000 : 4000);
}

async function apiError(response, fallback) {
  const body = await response.json().catch(() => ({}));
  return body.error || fallback;
}

// ===== list =====
async function loadOrders() {
  loading.value = true;
  loadError.value = "";
  try {
    const r = await fetch(`${API}/inventory/purchase-orders?view=${tab.value}`);
    if (!r.ok) throw new Error(await apiError(r, "Could not load purchase orders."));
    orders.value = (await r.json()) || [];
  } catch (e) {
    loadError.value = e.message || "Could not load purchase orders.";
  } finally {
    loading.value = false;
  }
}

function switchTab(key) {
  tab.value = key;
  ordersPage.value = 1;
  closeTarget.value = null;
  moveTarget.value = null;
  loadOrders();
}

function openDetails(id) {
  detailQuery.value = "";
  detailFilter.value = "all";
  detailPage.value = 1;
  history.value = [];
  showActivity.value = false;
  detailId.value = id;
  loadHistory(id);
}

async function loadHistory(id) {
  try {
    const r = await fetch(`${API}/inventory/purchase-orders/${id}/history`);
    if (r.ok && detailId.value === id) history.value = (await r.json()) || [];
  } catch {
    history.value = [];
  }
}

// Turn an audit record into short, readable lines.
function eventLines(e) {
  const d = e.details || {};
  const lines = [];
  const arrow = (c) => `${c.from} → ${c.to}`;
  if (e.action === "created") {
    lines.push(d.moved_from_purchase_order_id
      ? `Created from the outstanding items of ${orderNumber(d.moved_from_purchase_order_id)}`
      : `Created with ${(d.items || []).length} item${(d.items || []).length === 1 ? "" : "s"}`);
  } else if (e.action === "edited") {
    if (d.supplier) lines.push(`Supplier: ${d.supplier.from || "none"} → ${d.supplier.to || "none"}`);
    if (d.note) lines.push("Order note changed");
    for (const i of d.added || []) lines.push(`Added ${i.name} × ${i.quantity}`);
    for (const i of d.removed || []) lines.push(`Removed ${i.name} (${i.quantity})`);
    for (const u of d.updated || []) {
      for (const [field, c] of Object.entries(u.changes)) {
        lines.push(field === "note" ? `${u.item}: note changed` : `${u.item}: ${field} ${arrow(c)}`);
      }
    }
  } else if (e.action === "received") {
    for (const l of d.lines || []) {
      lines.push(`Received ${l.item} × ${l.quantity}${l.product_linked ? " (linked to a product)" : ""} · ${l.outstanding_after} still outstanding`);
    }
    if (d.supplier) lines.push(`Delivered by ${d.supplier}`);
  } else if (e.action === "closed") {
    const units = (d.abandoned || []).reduce((sum, i) => sum + i.quantity, 0);
    lines.push(`Closed with ${units} unit${units === 1 ? "" : "s"} outstanding`);
  } else if (e.action === "moved") {
    lines.push(`Outstanding items moved to ${orderNumber(d.to_purchase_order_id)} (${d.supplier.from || "none"} → ${d.supplier.to})`);
  }
  return lines;
}

function closeDetails() {
  detailId.value = null;
}

async function closeOrder(o) {
  closing.value = true;
  try {
    const r = await csrfFetch(`${API}/inventory/purchase-orders/${o.id}/close`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ version: o.version, reason: composeReason(closeReason) }),
    });
    if (!r.ok) {
      const message = await apiError(r, "Could not close the purchase order.");
      if (r.status === 409) { closeTarget.value = null; await loadOrders(); }
      throw new Error(message);
    }
    closeTarget.value = null;
    showNotice(`${orderNumber(o.id)} closed.`);
    await loadOrders();
  } catch (e) {
    showNotice(e.message, true);
  } finally {
    closing.value = false;
  }
}

// A close or a move abandons or reassigns what was ordered, so both need a recorded reason.
function composeReason({ preset, detail }) {
  const text = detail.trim();
  if (!preset) return "";
  if (preset === "Other") return text;
  return text ? `${preset}: ${text}` : preset;
}

function startClose(o) {
  moveTarget.value = null;
  Object.assign(closeReason, { preset: "", detail: "" });
  closeTarget.value = o.id;
}

function startMove(o) {
  closeTarget.value = null;
  moveSupplier.value = "";
  Object.assign(moveReason, { preset: "", detail: "" });
  moveTarget.value = o.id;
}

async function moveOutstanding(o) {
  const name = moveSupplier.value.trim();
  if (!name) return;
  moving.value = true;
  try {
    const r = await csrfFetch(`${API}/inventory/purchase-orders/${o.id}/move-outstanding`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ version: o.version, supplier: name, reason: composeReason(moveReason) }),
    });
    if (!r.ok) {
      const message = await apiError(r, "Could not move the outstanding items.");
      if (r.status === 409) { moveTarget.value = null; await loadOrders(); }
      throw new Error(message);
    }
    const { id } = await r.json();
    moveTarget.value = null;
    showNotice(`Outstanding items moved to ${orderNumber(id)} for ${name}.`);
    await loadOrders();
  } catch (e) {
    showNotice(e.message, true);
  } finally {
    moving.value = false;
  }
}

function receive(o) {
  router.push({ path: "/receive-items", query: { purchase_order: String(o.id) } });
}

// ===== editor =====
function snapshot() {
  return JSON.stringify([
    supplier.value,
    note.value,
    lines.value.map((l) => [l.id, l.product_id, l.name, l.manufacturer, l.note, l.quantity]),
  ]);
}

function resetPicker() {
  query.value = "";
  results.value = [];
  lowStockOpen.value = false;
  lineFilter.value = "";
  linePage.value = 1;
  Object.keys(lowStockPick).forEach((k) => delete lowStockPick[k]);
  Object.assign(newItem, { name: "", manufacturer: "", note: "", quantity: 1 });
}

function openEditor() {
  conflict.value = false;
  resetPicker();
  closeTarget.value = null;
  moveTarget.value = null;
  detailId.value = null;
  editing.value = true;
  savedSnapshot.value = snapshot();
}

function startNew({ withLowStock = false } = {}) {
  editId.value = 0;
  editStatus.value = "open";
  supplier.value = "";
  note.value = "";
  lines.value = [];
  openEditor();
  if (withLowStock) toggleLowStock();
}

function startEdit(o) {
  editId.value = o.id;
  editStatus.value = o.status;
  editVersion.value = o.version;
  supplier.value = o.supplier || "";
  note.value = o.note || "";
  lines.value = o.items.map((i) => ({
    key: ++lineKey,
    id: i.id,
    product_id: i.product_id,
    name: i.name,
    manufacturer: i.manufacturer || "",
    note: i.note || "",
    quantity: i.quantity,
    received_quantity: i.received_quantity,
  }));
  openEditor();
}

// Someone else changed the order while it was open here. Throw away this copy and load theirs.
async function reloadLatest() {
  try {
    const r = await fetch(`${API}/inventory/purchase-orders/${editId.value}`);
    if (!r.ok) throw new Error(await apiError(r, "Could not reload the purchase order."));
    const latest = await r.json();
    if (!isEditable(latest)) {
      leaveEditor();
      showNotice(`${orderNumber(latest.id)} is now ${latest.status} and can no longer be edited.`, true);
      return;
    }
    startEdit(latest);
    showNotice(`Reloaded ${orderNumber(latest.id)} with the latest changes.`);
  } catch (e) {
    showNotice(e.message, true);
  }
}

function cancelEdit() {
  if (isDirty.value && !window.confirm("Discard your unsaved changes?")) return;
  leaveEditor();
}

function leaveEditor() {
  editing.value = false;
  savedSnapshot.value = "";
  loadOrders();
}

// New lines go to the top of the list so what was just added is always visible.
function pushLine(fields) {
  lines.value.unshift({ key: ++lineKey, id: 0, received_quantity: 0, manufacturer: "", note: "", ...fields });
  lineFilter.value = "";
  linePage.value = 1;
}

function removeLine(key) {
  lines.value = lines.value.filter((l) => l.key !== key);
}

function addProduct(p, quantity = 1, itemNote = "") {
  query.value = "";
  results.value = [];
  if (lines.value.some((l) => l.product_id === p.id)) {
    showNotice(`${p.name} is already on this order.`, true);
    return false;
  }
  pushLine({ product_id: p.id, name: p.name, manufacturer: p.manufacturer || "", note: itemNote, quantity });
  return true;
}

function searchProducts() {
  clearTimeout(searchTimer);
  const q = query.value.trim();
  if (q.length < 2) {
    results.value = [];
    return;
  }
  searchTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`);
      if (r.ok) results.value = (await r.json()) || [];
    } catch {
      results.value = [];
    }
  }, 200);
}

function searchSuppliers(text) {
  clearTimeout(supplierTimer);
  const q = (text || "").trim();
  if (!q) {
    supplierSuggestions.value = [];
    return;
  }
  supplierTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/suppliers/search?query=${encodeURIComponent(q)}`);
      if (r.ok) supplierSuggestions.value = ((await r.json()).data) || [];
    } catch {
      supplierSuggestions.value = [];
    }
  }, 200);
}

function onOrderLabel(productId) {
  const n = outstandingByProduct.value[productId];
  return n ? `${n} on order` : "";
}

// An item typed by name may already exist as a product. Use that product when the name matches
// exactly, so the same drug is not ordered under two identities.
async function addNewItem() {
  const name = newItem.name.trim();
  const manufacturer = newItem.manufacturer.trim();
  const itemNote = newItem.note.trim();
  const quantity = Number(newItem.quantity);
  if (!name) return;
  if (!Number.isInteger(quantity) || quantity < 1) {
    showNotice("Enter a quantity of at least 1.", true);
    return;
  }
  addingNew.value = true;
  try {
    let match = null;
    try {
      const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(name)}`);
      if (r.ok) {
        const found = (await r.json()) || [];
        match = found.find((p) => p.name.trim().toLowerCase() === name.toLowerCase()
          && (!manufacturer || (p.manufacturer || "").toLowerCase() === manufacturer.toLowerCase()));
      }
    } catch {
      match = null;
    }
    if (match) {
      if (addProduct(match, quantity, itemNote)) showNotice(`${match.name} is already in your products, so it was added as that product.`);
    } else if (lines.value.some((l) => !l.product_id && l.name.toLowerCase() === name.toLowerCase()
      && l.manufacturer.toLowerCase() === manufacturer.toLowerCase())) {
      showNotice(`${name} is already on this order.`, true);
      return;
    } else {
      pushLine({ product_id: null, name, manufacturer, note: itemNote, quantity });
    }
    Object.assign(newItem, { name: "", manufacturer: "", note: "", quantity: 1 });
  } finally {
    addingNew.value = false;
  }
}

// --- low stock ---
async function toggleLowStock() {
  lowStockOpen.value = !lowStockOpen.value;
  if (!lowStockOpen.value) return;
  lowStockLoading.value = true;
  try {
    const [inventory, active] = await Promise.all([
      fetch(`${API}/inventory/item-list`).then((r) => (r.ok ? r.json() : { items: [] })),
      fetch(`${API}/inventory/purchase-orders?view=active`).then((r) => (r.ok ? r.json() : [])),
    ]);
    lowStockItems.value = (inventory.items || [])
      .filter((p) => p.stock <= p.reorder_level)
      .sort((a, b) => a.stock - b.stock || a.name.localeCompare(b.name));
    // Units already ordered but not delivered. The order being edited is excluded because its
    // lines are already in the editor.
    const outstanding = {};
    for (const o of active || []) {
      if (o.id === editId.value) continue;
      for (const i of o.items) {
        if (i.product_id) outstanding[i.product_id] = (outstanding[i.product_id] || 0) + i.quantity - i.received_quantity;
      }
    }
    outstandingByProduct.value = outstanding;
  } catch {
    lowStockItems.value = [];
    showNotice("Could not load low stock items.", true);
  } finally {
    lowStockLoading.value = false;
  }
}

function suggestedQuantity(row) {
  return Math.max(1, row.reorder_level - row.stock - row.onOrder);
}

function toggleLowStockRow(row, checked) {
  if (checked) lowStockPick[row.id] = suggestedQuantity(row);
  else delete lowStockPick[row.id];
}

function selectAllLowStock() {
  for (const row of lowStockRows.value) {
    if (!row.added && !(row.id in lowStockPick)) lowStockPick[row.id] = suggestedQuantity(row);
  }
}

function clearLowStock() {
  Object.keys(lowStockPick).forEach((k) => delete lowStockPick[k]);
}

function addPickedLowStock() {
  for (const row of lowStockRows.value) {
    if (!(row.id in lowStockPick) || row.added) continue;
    addProduct(row, Math.max(1, Math.floor(Number(lowStockPick[row.id])) || 1));
  }
  clearLowStock();
}

// --- save ---
function validateLines() {
  for (const l of lines.value) {
    if (!Number.isInteger(l.quantity) || l.quantity < 1) return `${l.name}: enter a whole quantity of at least 1.`;
    if (l.quantity < l.received_quantity) return `${l.name}: ${l.received_quantity} already received, so the quantity cannot be lower.`;
  }
  return "";
}

async function save() {
  const problem = validateLines();
  if (problem) {
    showNotice(problem, true);
    return;
  }
  saving.value = true;
  try {
    const body = {
      id: editId.value,
      version: editVersion.value,
      supplier: supplier.value.trim(),
      note: note.value.trim(),
      items: lines.value.map((l) => ({
        id: l.id,
        product_id: l.product_id,
        name: l.name,
        manufacturer: l.manufacturer,
        note: (l.note || "").trim(),
        quantity: l.quantity,
      })),
    };
    const r = await csrfFetch(`${API}/inventory/purchase-orders`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (r.status === 409 && editId.value) {
      const message = await apiError(r, "");
      if (message.includes("changed by someone else")) {
        conflict.value = true;
        return;
      }
      throw new Error(message || "Could not save the purchase order.");
    }
    if (!r.ok) throw new Error(await apiError(r, "Could not save the purchase order."));
    const { id } = await r.json();
    showNotice(`${orderNumber(id)} saved.`);
    tab.value = "active";
    ordersPage.value = 1;
    leaveEditor();
  } catch (e) {
    showNotice(e.message, true);
  } finally {
    saving.value = false;
  }
}

// ===== PDF =====
function exportDraft() {
  exportOrder({
    id: editId.value,
    supplier: supplier.value.trim(),
    note: note.value.trim(),
    created_at: new Date().toISOString(),
    items: lines.value,
  });
}

function exportOrder(o) {
  const doc = new jsPDF();
  const hasProgress = o.items.some((i) => i.received_quantity > 0);
  doc.setFontSize(18);
  doc.text("Purchase Order", 14, 20);
  doc.setFontSize(10);
  doc.text(o.id ? orderNumber(o.id) : "DRAFT", 14, 28);
  doc.text(`Supplier: ${o.supplier || "____________________"}`, 14, 35);
  doc.text(`Date: ${formatDate(o.created_at)}`, 14, 42);

  let startY = 48;
  if (o.note) {
    const wrapped = doc.splitTextToSize(`Notes: ${o.note}`, 180);
    doc.text(wrapped, 14, 49);
    startY = 49 + wrapped.length * 5 + 4;
  }

  const hasNotes = o.items.some((i) => i.note);
  const head = [["Item", "Manufacturer", ...(hasProgress ? ["Ordered", "Received", "Outstanding"] : ["Quantity"]), ...(hasNotes ? ["Notes"] : [])]];
  const body = o.items.map((i) => [
    i.name,
    i.manufacturer || "-",
    ...(hasProgress ? [i.quantity, i.received_quantity, i.quantity - i.received_quantity] : [i.quantity]),
    ...(hasNotes ? [i.note || ""] : []),
  ]);
  const numeric = hasProgress ? [2, 3, 4] : [2];
  autoTable(doc, {
    startY,
    head,
    body,
    styles: { fontSize: 10 },
    headStyles: { fillColor: [22, 101, 82] },
    columnStyles: Object.fromEntries(numeric.map((c) => [c, { halign: "center" }])),
  });
  doc.save(`purchase-order-${o.id ? String(o.id).padStart(5, "0") : "draft"}.pdf`);
}

// ===== lifecycle =====
function onKeydown(e) {
  if (e.key === "Escape" && detailId.value) closeDetails();
}

function warnBeforeUnload(e) {
  if (isDirty.value) {
    e.preventDefault();
    e.returnValue = "";
  }
}

onBeforeRouteLeave(() => {
  if (isDirty.value) return window.confirm("Discard your unsaved purchase order?");
  return true;
});

onMounted(async () => {
  window.addEventListener("beforeunload", warnBeforeUnload);
  window.addEventListener("keydown", onKeydown);
  if (!canManage.value) return;
  if (route.query.lowStock) {
    startNew({ withLowStock: true });
    router.replace({ path: route.path });
    return;
  }
  loadOrders();
});

onBeforeUnmount(() => {
  window.removeEventListener("beforeunload", warnBeforeUnload);
  window.removeEventListener("keydown", onKeydown);
  clearTimeout(noticeTimer);
  clearTimeout(searchTimer);
  clearTimeout(supplierTimer);
});
</script>
