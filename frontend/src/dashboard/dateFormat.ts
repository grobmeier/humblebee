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

export type DateLanguage = "de" | "en";

export function formatInputDate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatDisplayDate(isoDate: string, language: DateLanguage): string {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(isoDate);
  if (!match) {
    return isoDate;
  }
  if (language === "en") {
    return `${match[2]}/${match[3]}/${match[1]}`;
  }
  return `${match[3]}.${match[2]}.${match[1]}`;
}

export function parseDisplayDate(displayDate: string, language: DateLanguage): string | null {
  if (language === "en") {
    const match = /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/.exec(displayDate.trim());
    if (!match) {
      return null;
    }
    return formatValidatedDate(Number(match[3]), Number(match[1]), Number(match[2]));
  }

  const match = /^(\d{1,2})[.](\d{1,2})[.](\d{4})$/.exec(displayDate.trim());
  if (!match) {
    return null;
  }

  const day = Number(match[1]);
  const month = Number(match[2]);
  const year = Number(match[3]);
  return formatValidatedDate(year, month, day);
}

export function flatpickrDateFormat(language: DateLanguage): string {
  return language === "en" ? "m/d/Y" : "d.m.Y";
}

function formatValidatedDate(year: number, month: number, day: number): string | null {
  const date = new Date(year, month - 1, day);
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) {
    return null;
  }

  return `${String(year).padStart(4, "0")}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
}

export function formatTime(date: Date): string {
  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}
