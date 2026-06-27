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

import { useEffect, useRef, useState, type FormEventHandler, type KeyboardEvent } from "react";
import { FormRow, Modal } from "../components/Modal";
import { flatpickrDateFormat, formatDisplayDate, parseDisplayDate, type DateLanguage } from "./dateFormat";
import type { TimeEntryFormState } from "./timeEntryTypes";
import { labelWorkItemName } from "./workItemUtils";
import { decodeWindowsAltCode, numpadDigitFromKeyboardEvent } from "./windowsAltCodeInput";

type WorkItem = { id: number; name: string; parentId?: number | null; depth: number; status?: string };

type TimeEntryModalProps = {
  error: string | null;
  form: TimeEntryFormState;
  isSaving: boolean;
  language: DateLanguage;
  t: {
    end: string;
    note: string;
    project: string;
    save: string;
    saving: string;
    start: string;
    task: string;
    title: string;
    untilMidnight: string;
  };
  onChange: (form: TimeEntryFormState) => void;
  onClose: () => void;
  onSubmit: FormEventHandler<HTMLFormElement>;
  workItems: WorkItem[];
};

type FlatpickrInstance = {
  destroy: () => void;
  setDate: (value: string | Date, triggerChange?: boolean, format?: string) => void;
};

type FlatpickrFactory = (element: HTMLInputElement, options: Record<string, unknown>) => FlatpickrInstance;

declare global {
  interface Window {
    flatpickr?: FlatpickrFactory;
  }
}

export function TimeEntryModal({ error, form, isSaving, language, t, onChange, onClose, onSubmit, workItems }: TimeEntryModalProps) {
  const projects = workItems.filter((workItem) => workItem.parentId == null);
  const tasks = workItems.filter((workItem) => workItem.parentId === form.projectId);
  const noteRef = useRef<HTMLTextAreaElement | null>(null);
  const pendingWindowsAltCodeRef = useRef("");

  function insertNoteCharacter(character: string, textarea: HTMLTextAreaElement) {
    const selectionStart = textarea.selectionStart ?? textarea.value.length;
    const selectionEnd = textarea.selectionEnd ?? selectionStart;
    const nextDescription = textarea.value.slice(0, selectionStart) + character + textarea.value.slice(selectionEnd);
    const nextCursor = selectionStart + character.length;
    onChange({ ...form, description: nextDescription });
    window.requestAnimationFrame(() => {
      noteRef.current?.setSelectionRange(nextCursor, nextCursor);
    });
  }

  function handleNoteKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (!event.altKey) {
      return;
    }
    const digit = numpadDigitFromKeyboardEvent(event);
    if (!digit) {
      return;
    }
    pendingWindowsAltCodeRef.current += digit;
    event.preventDefault();
  }

  function handleNoteKeyUp(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key !== "Alt" && event.code !== "AltLeft" && event.code !== "AltRight") {
      return;
    }
    const digits = pendingWindowsAltCodeRef.current;
    pendingWindowsAltCodeRef.current = "";
    if (!digits) {
      return;
    }
    event.preventDefault();
    const character = decodeWindowsAltCode(digits);
    if (character) {
      insertNoteCharacter(character, event.currentTarget);
    }
  }

  return (
    <Modal
      title={t.title}
      onClose={onClose}
      onSubmit={onSubmit}
      footer={
        <button className="primary-button modal-submit-button" type="submit" disabled={isSaving || !form.projectId || !form.taskId}>
          {isSaving ? t.saving : t.save}
        </button>
      }
    >
      {error ? <div className="errors alert alert-error">{error}</div> : null}
      <FormRow label={t.start} controlsClassName="tab-form-controls tab-form-controls--inline">
        <DashboardDateInput
          className="tab-form-control tab-form-control--compact tab-form-control--small"
          language={language}
          value={form.startDate}
          onChange={(value) => onChange({ ...form, startDate: value })}
        />
        <DashboardTimeInput
          className="tab-form-control tab-form-control--compact tab-form-control--small"
          value={form.startTime}
          onChange={(value) => onChange({ ...form, startTime: value })}
        />
      </FormRow>

      {!form.untilMidnight ? (
        <FormRow label={t.end} controlsClassName="tab-form-controls tab-form-controls--inline">
          <DashboardDateInput
            className="tab-form-control tab-form-control--compact tab-form-control--small"
            language={language}
            value={form.endDate}
            onChange={(value) => onChange({ ...form, endDate: value })}
          />
          <DashboardTimeInput
            className="tab-form-control tab-form-control--compact tab-form-control--small"
            value={form.endTime}
            onChange={(value) => onChange({ ...form, endTime: value })}
          />
        </FormRow>
      ) : null}

      <FormRow label="" labelHidden>
        <label className="tab-form-checkbox tab-form-checkbox--row">
          <input
            type="checkbox"
            checked={form.untilMidnight}
            onChange={(event) =>
              onChange({
                ...form,
                untilMidnight: event.target.checked,
                endDate: event.target.checked ? form.startDate : form.endDate,
                endTime: event.target.checked ? "00:00" : form.endTime
              })
            }
          />
          {t.untilMidnight}
        </label>
      </FormRow>

      <FormRow label={t.project}>
        <select
          className="tab-form-control"
          value={form.projectId}
          onChange={(event) => {
            const projectId = Number(event.target.value);
            const firstTask = workItems.find((workItem) => workItem.parentId === projectId);
            onChange({ ...form, projectId, taskId: firstTask?.id ?? 0 });
          }}
        >
          <option value={0}></option>
          {projects.map((workItem) => (
            <option key={workItem.id} value={workItem.id}>
              {labelWorkItemName(workItem.name, language)}
            </option>
          ))}
        </select>
      </FormRow>

      <FormRow label={t.task}>
        <select
          className="tab-form-control"
          value={form.taskId}
          disabled={!form.projectId || tasks.length === 0}
          onChange={(event) => onChange({ ...form, taskId: Number(event.target.value) })}
        >
          <option value={0}></option>
          {tasks.map((task) => (
            <option key={task.id} value={task.id}>
              {labelWorkItemName(task.name, language)}
            </option>
          ))}
        </select>
      </FormRow>

      <FormRow label={t.note}>
        <textarea
          className="tab-form-control"
          ref={noteRef}
          rows={4}
          value={form.description}
          onBlur={() => {
            pendingWindowsAltCodeRef.current = "";
          }}
          onChange={(event) => onChange({ ...form, description: event.target.value })}
          onKeyDown={handleNoteKeyDown}
          onKeyUp={handleNoteKeyUp}
        />
      </FormRow>
    </Modal>
  );
}

type DashboardDateInputProps = {
  className?: string;
  language: DateLanguage;
  value: string;
  onChange: (value: string) => void;
};

export function DashboardDateInput({ className, language, value, onChange }: DashboardDateInputProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const instanceRef = useRef<FlatpickrInstance | null>(null);
  const onChangeRef = useRef(onChange);
  const languageRef = useRef(language);
  const valueRef = useRef(value);
  const [flatpickrReady, setFlatpickrReady] = useState(() => typeof window !== "undefined" && typeof window.flatpickr === "function");
  const [displayValue, setDisplayValue] = useState(() => formatDisplayDate(value, language));

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  useEffect(() => {
    languageRef.current = language;
  }, [language]);

  useEffect(() => {
    valueRef.current = value;
  }, [value]);

  useEffect(() => {
    if (flatpickrReady) {
      return;
    }

    let cancelled = false;
    let attempts = 0;

    const pollUntilReady = () => {
      if (cancelled) {
        return;
      }
      if (typeof window.flatpickr === "function") {
        setFlatpickrReady(true);
        return;
      }
      attempts += 1;
      if (attempts < 20) {
        window.setTimeout(pollUntilReady, 50);
      }
    };

    pollUntilReady();

    return () => {
      cancelled = true;
    };
  }, [flatpickrReady]);

  useEffect(() => {
    const element = inputRef.current;
    const flatpickr = window.flatpickr;
    if (!element || typeof flatpickr !== "function") {
      return;
    }

    const dateFormat = flatpickrDateFormat(language);
    instanceRef.current = flatpickr(element, {
      allowInput: true,
      dateFormat,
      defaultDate: value ? formatDisplayDate(value, language) : undefined,
      disableMobile: true,
      weekNumbers: true,
      onOpen: (_selectedDates: unknown, _dateStr: string, instance: { setDate: (value: string | Date, triggerChange?: boolean, format?: string) => void }) => {
        instance.setDate(valueRef.current ? formatDisplayDate(valueRef.current, languageRef.current) : new Date(), false, flatpickrDateFormat(languageRef.current));
      },
      onChange: (_selectedDates: unknown, dateStr: string) => {
        updateDate(dateStr);
      },
      onClose: (_selectedDates: unknown, dateStr: string) => {
        const parsed = parseDisplayDate(dateStr, languageRef.current);
        if (parsed) {
          onChangeRef.current(parsed);
          return;
        }
        setDisplayValue(formatDisplayDate(valueRef.current, languageRef.current));
      }
    });

    return () => {
      instanceRef.current?.destroy();
      instanceRef.current = null;
    };
  }, [flatpickrReady, language]);

  useEffect(() => {
    setDisplayValue(formatDisplayDate(value, language));
    instanceRef.current?.setDate(formatDisplayDate(value, language), false, flatpickrDateFormat(language));
  }, [language, value]);

  function updateDate(nextValue: string) {
    setDisplayValue(nextValue);
    const parsed = parseDisplayDate(nextValue, languageRef.current);
    if (parsed) {
      onChangeRef.current(parsed);
    }
  }

  return (
    <input
      ref={inputRef}
      className={className}
      type="text"
      inputMode="numeric"
      pattern={language === "en" ? "[0-9]{1,2}/[0-9]{1,2}/[0-9]{4}" : "[0-9]{1,2}[.][0-9]{1,2}[.][0-9]{4}"}
      placeholder={language === "en" ? "MM/DD/YYYY" : "TT.MM.JJJJ"}
      value={displayValue}
      onBlur={() => {
        if (!parseDisplayDate(displayValue, language)) {
          setDisplayValue(formatDisplayDate(value, language));
        }
      }}
      onChange={(event) => updateDate(event.target.value)}
    />
  );
}

type DashboardTimeInputProps = {
  className?: string;
  value: string;
  onChange: (value: string) => void;
};

function DashboardTimeInput({ className, value, onChange }: DashboardTimeInputProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const instanceRef = useRef<FlatpickrInstance | null>(null);
  const onChangeRef = useRef(onChange);
  const valueRef = useRef(value);
  const [flatpickrReady, setFlatpickrReady] = useState(() => typeof window !== "undefined" && typeof window.flatpickr === "function");
  const [displayValue, setDisplayValue] = useState(value);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  useEffect(() => {
    valueRef.current = value;
  }, [value]);

  useEffect(() => {
    if (flatpickrReady) {
      return;
    }

    let cancelled = false;
    let attempts = 0;

    const pollUntilReady = () => {
      if (cancelled) {
        return;
      }
      if (typeof window.flatpickr === "function") {
        setFlatpickrReady(true);
        return;
      }
      attempts += 1;
      if (attempts < 20) {
        window.setTimeout(pollUntilReady, 50);
      }
    };

    pollUntilReady();

    return () => {
      cancelled = true;
    };
  }, [flatpickrReady]);

  useEffect(() => {
    const element = inputRef.current;
    const flatpickr = window.flatpickr;
    if (!element || typeof flatpickr !== "function") {
      return;
    }

    instanceRef.current = flatpickr(element, {
      allowInput: true,
      dateFormat: "H:i",
      defaultDate: value || undefined,
      disableMobile: true,
      enableTime: true,
      minuteIncrement: 1,
      noCalendar: true,
      time_24hr: true,
      onOpen: (_selectedDates: unknown, _dateStr: string, instance: { setDate: (value: string | Date, triggerChange?: boolean, format?: string) => void }) => {
        instance.setDate(valueRef.current || "00:00", false, "H:i");
      },
      onChange: (_selectedDates: unknown, dateStr: string) => {
        setDisplayValue(dateStr);
        if (isTimeValue(dateStr)) {
          onChangeRef.current(dateStr);
        }
      },
      onClose: (_selectedDates: unknown, dateStr: string) => {
        if (isTimeValue(dateStr)) {
          setDisplayValue(dateStr);
          onChangeRef.current(dateStr);
          return;
        }
        setDisplayValue(valueRef.current);
        instanceRef.current?.setDate(valueRef.current, false, "H:i");
      }
    });

    return () => {
      instanceRef.current?.destroy();
      instanceRef.current = null;
    };
  }, [flatpickrReady]);

  useEffect(() => {
    instanceRef.current?.setDate(value, false, "H:i");
    setDisplayValue(value);
  }, [value]);

  function updateTime(nextValue: string) {
    setDisplayValue(nextValue);
    if (isTimeValue(nextValue)) {
      onChangeRef.current(nextValue);
    }
  }

  return (
    <input
      ref={inputRef}
      className={className}
      inputMode="numeric"
      pattern="[0-2][0-9]:[0-5][0-9]"
      placeholder="HH:MM"
      type="text"
      value={displayValue}
      onBlur={() => {
        if (!isTimeValue(displayValue)) {
          setDisplayValue(valueRef.current);
        }
      }}
      onChange={(event) => updateTime(event.target.value)}
    />
  );
}

function isTimeValue(value: string): boolean {
  return /^([01][0-9]|2[0-3]):[0-5][0-9]$/.test(value);
}
