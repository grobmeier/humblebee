# Flatpickr Local Patch

`flatpickr-4.5.2.min.js` has a guard in the calendar sizing animation-frame
callback: return when `calendarContainer` no longer exists. Flatpickr destroys
that property during cleanup, but a previously queued callback can still run.
React Strict Mode and rapid modal teardown expose this race.

Keep this guard when replacing the vendored script unless the replacement fixes
the race. `frontend/e2e/date-time-picker.spec.ts` verifies repeated modal opening,
date/time popup usability, picker cleanup, and absence of browser exceptions.
