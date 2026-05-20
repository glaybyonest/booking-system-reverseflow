-- Make all events use internal ReserveFlow booking (no external redirects)
UPDATE events
SET booking_mode = 'reserveflow_managed', updated_at = NOW()
WHERE booking_mode = 'external_link_only';

-- Mark all scheduled sessions as bookable so they appear on booking pages
UPDATE sessions
SET is_bookable = TRUE, updated_at = NOW()
WHERE status = 'scheduled' AND is_bookable = FALSE;
