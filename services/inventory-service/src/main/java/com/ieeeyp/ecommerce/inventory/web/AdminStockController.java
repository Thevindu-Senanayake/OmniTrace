package com.ieeeyp.ecommerce.inventory.web;

import com.ieeeyp.ecommerce.inventory.domain.Inventory;
import com.ieeeyp.ecommerce.inventory.service.InventoryService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

/**
 * Admin endpoints for inspecting and resetting stock during demos. Not part of
 * the saga flow - purely operational tooling.
 */
@RestController
@RequestMapping("/admin/stock")
@RequiredArgsConstructor
public class AdminStockController {

    private final InventoryService inventoryService;

    /** Current stock levels for every item. */
    @GetMapping
    public List<StockView> listStock() {
        return inventoryService.listStock().stream()
                .map(StockView::from)
                .toList();
    }

    /** Add {@code quantity} units back to an item so a demo can be re-run. */
    @PostMapping("/{itemId}/replenish")
    public StockView replenish(@PathVariable String itemId, @RequestParam int quantity) {
        return StockView.from(inventoryService.replenish(itemId, quantity));
    }

    /** Read model - keeps the JPA entity out of the HTTP contract. */
    public record StockView(String itemId, int stock, int reserved) {
        static StockView from(Inventory i) {
            return new StockView(i.getItemId(), i.getStock(), i.getReserved());
        }
    }
}
