import {
  addDays,
  buildWeekDays,
  formatCalendarHeadline,
  sameCalendarDay
} from "./calendarUtils";
import type { Language } from "./translations";

type DashboardCalendarProps = {
  language: Language;
  selectedDate: Date;
  onAddEntry: (date: Date) => void;
  onSelectDate: (date: Date) => void;
};

export function DashboardCalendar({
  language,
  selectedDate,
  onAddEntry,
  onSelectDate
}: DashboardCalendarProps) {
  const weekDays = buildWeekDays(selectedDate, language);

  function selectCalendarDate(date: Date) {
    if (sameCalendarDay(date, selectedDate)) {
      onAddEntry(date);
      return;
    }
    onSelectDate(date);
  }

  return (
    <section className="dashboard-calendar">
      <div className="calendar-headline">
        <h2>{formatCalendarHeadline(selectedDate, language)}</h2>
        <div className="calendar-navigation" aria-label="Kalendernavigation">
          <button className="secondary-button" type="button" aria-label="Vorherige Woche" onClick={() => onSelectDate(addDays(selectedDate, -7))}>
            «
          </button>
          <button className="secondary-button" type="button" aria-label="Vorheriger Tag" onClick={() => onSelectDate(addDays(selectedDate, -1))}>
            ‹
          </button>
          <button className="secondary-button today-button" type="button" onClick={() => onSelectDate(new Date())}>
            Heute
          </button>
          <button className="secondary-button" type="button" aria-label="Naechster Tag" onClick={() => onSelectDate(addDays(selectedDate, 1))}>
            ›
          </button>
          <button className="secondary-button" type="button" aria-label="Naechste Woche" onClick={() => onSelectDate(addDays(selectedDate, 7))}>
            »
          </button>
        </div>
      </div>

      <div className="week-calendar" aria-label="Aktuelle Woche">
        {weekDays.map((day) => (
          <button className={`week-day ${day.isSelected ? "selected" : ""}`} key={day.isoDate} type="button" onClick={() => selectCalendarDate(day.date)}>
            <span>{day.dayNumber}</span>
            <small>{day.dayName}</small>
          </button>
        ))}
      </div>

      <div className="calendar-footer">
        <button className="primary-button" type="button" onClick={() => onAddEntry(selectedDate)}>
          Zeit erfassen
        </button>
      </div>
    </section>
  );
}
