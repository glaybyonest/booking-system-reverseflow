-- Note: cannot selectively revert — sets all external-source events back
UPDATE events
SET booking_mode = 'external_link_only', updated_at = NOW()
WHERE source IN ('yandex_afisha', 'kudago', 'timepad');

UPDATE sessions
SET is_bookable = FALSE, updated_at = NOW()
WHERE external_source IN ('yandex_afisha', 'kudago', 'timepad');
