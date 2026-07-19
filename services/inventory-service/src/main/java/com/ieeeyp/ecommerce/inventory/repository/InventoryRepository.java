package com.ieeeyp.ecommerce.inventory.repository;

import com.ieeeyp.ecommerce.inventory.domain.Inventory;
import jakarta.persistence.LockModeType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Lock;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.Optional;

public interface InventoryRepository extends JpaRepository<Inventory, String> {

    /**
     * Loads an inventory row with a pessimistic write lock. Translates to
     * {@code SELECT ... FOR UPDATE}, so concurrent reservations for the same
     * item block here until the holder's transaction commits - the genuine
     * row-lock contention this project is built to observe.
     */
    @Lock(LockModeType.PESSIMISTIC_WRITE)
    @Query("SELECT i FROM Inventory i WHERE i.itemId = :itemId")
    Optional<Inventory> findByIdForUpdate(@Param("itemId") String itemId);
}
