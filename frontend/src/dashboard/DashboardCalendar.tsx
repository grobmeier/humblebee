/*
 * Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
  t: {
    addTime: string;
    currentWeek: string;
    navigation: string;
    nextDay: string;
    nextWeek: string;
    previousDay: string;
    previousWeek: string;
    today: string;
  };
  onAddEntry: (date: Date) => void;
  onSelectDate: (date: Date) => void;
};

export function DashboardCalendar({
  language,
  selectedDate,
  t,
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
        <div className="calendar-navigation" aria-label={t.navigation}>
          <button className="secondary-button" type="button" aria-label={t.previousWeek} onClick={() => onSelectDate(addDays(selectedDate, -7))}>
            «
          </button>
          <button className="secondary-button" type="button" aria-label={t.previousDay} onClick={() => onSelectDate(addDays(selectedDate, -1))}>
            ‹
          </button>
          <button className="secondary-button today-button" type="button" onClick={() => onSelectDate(new Date())}>
            {t.today}
          </button>
          <button className="secondary-button" type="button" aria-label={t.nextDay} onClick={() => onSelectDate(addDays(selectedDate, 1))}>
            ›
          </button>
          <button className="secondary-button" type="button" aria-label={t.nextWeek} onClick={() => onSelectDate(addDays(selectedDate, 7))}>
            »
          </button>
        </div>
      </div>

      <div className="week-calendar" aria-label={t.currentWeek}>
        {weekDays.map((day) => (
          <button className={`week-day ${day.isSelected ? "selected" : ""}`} key={day.isoDate} type="button" onClick={() => selectCalendarDate(day.date)}>
            <span>{day.dayNumber}</span>
            <small>{day.dayName}</small>
          </button>
        ))}
      </div>

      <div className="calendar-footer">
        <button className="primary-button" type="button" onClick={() => onAddEntry(selectedDate)}>
          {t.addTime}
        </button>
      </div>
    </section>
  );
}
