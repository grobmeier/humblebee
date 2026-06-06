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

export type CalendarDay = {
  date: Date;
  dayNumber: number;
  dayName: string;
  isoDate: string;
  isSelected: boolean;
};

export type CalendarLanguage = "de" | "en";

const germanDayNames = ["So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"];
const englishDayNames = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const germanMonthNames = [
  "Januar",
  "Februar",
  "Maerz",
  "April",
  "Mai",
  "Juni",
  "Juli",
  "August",
  "September",
  "Oktober",
  "November",
  "Dezember"
];
const englishMonthNames = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];

export function addDays(date: Date, days: number): Date {
  const next = new Date(date);
  next.setDate(date.getDate() + days);
  return atLocalNoon(next);
}

export function sameCalendarDay(left: Date, right: Date): boolean {
  return left.getFullYear() === right.getFullYear() && left.getMonth() === right.getMonth() && left.getDate() === right.getDate();
}

export function atLocalNoon(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 12, 0, 0, 0);
}

export function buildWeekDays(selectedDate: Date, language: CalendarLanguage): CalendarDay[] {
  const monday = startOfIsoWeek(selectedDate);
  const dayNames = language === "de" ? germanDayNames : englishDayNames;

  return Array.from({ length: 7 }, (_, index) => {
    const date = addDays(monday, index);
    return {
      date,
      dayNumber: date.getDate(),
      dayName: dayNames[date.getDay()],
      isoDate: toIsoDate(date),
      isSelected: sameCalendarDay(date, selectedDate)
    };
  });
}

export function formatCalendarHeadline(selectedDate: Date, language: CalendarLanguage): string {
  const monthNames = language === "de" ? germanMonthNames : englishMonthNames;
  const weekLabel = language === "de" ? "Woche" : "Week";
  return `${monthNames[selectedDate.getMonth()]} ${selectedDate.getFullYear()}  -  ${weekLabel} ${isoWeekNumber(selectedDate)}`;
}

function startOfIsoWeek(date: Date): Date {
  const normalized = atLocalNoon(date);
  const weekday = (normalized.getDay() + 6) % 7;
  return addDays(normalized, -weekday);
}

function isoWeekNumber(date: Date): number {
  const utcDate = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const day = utcDate.getUTCDay() || 7;
  utcDate.setUTCDate(utcDate.getUTCDate() + 4 - day);
  const yearStart = new Date(Date.UTC(utcDate.getUTCFullYear(), 0, 1));
  return Math.ceil(((utcDate.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
}

function toIsoDate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}
