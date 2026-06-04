import { useEffect, useRef, useState, type FormEventHandler } from "react";
import { DashboardFormRow, DashboardModal } from "./DashboardModal";
import { flatpickrDateFormat, formatDisplayDate, parseDisplayDate, type DateLanguage } from "./dateFormat";
import type { TimeEntryFormState } from "./timeEntryTypes";
import { displayWorkItem } from "./workItemUtils";

type WorkItem = { id: number; name: string; depth: number };

type TimeEntryModalProps = {
  error: string | null;
  form: TimeEntryFormState;
  isSaving: boolean;
  language: DateLanguage;
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

export function TimeEntryModal({ error, form, isSaving, language, onChange, onClose, onSubmit, workItems }: TimeEntryModalProps) {
  const selectedWorkItem = displayWorkItem(form.projectId, workItems);

  return (
    <DashboardModal
      title="Zeiteintrag erfassen"
      onClose={onClose}
      onSubmit={onSubmit}
      footer={
        <button className="primary-button modal-submit-button" type="submit" disabled={isSaving || !form.projectId}>
          {isSaving ? "Speichern..." : "Speichern"}
        </button>
      }
    >
      {error ? <div className="errors alert alert-error">{error}</div> : null}
      <DashboardFormRow label="Start" controlsClassName="tab-form-controls tab-form-controls--inline">
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
      </DashboardFormRow>

      {!form.untilMidnight ? (
        <DashboardFormRow label="Ende" controlsClassName="tab-form-controls tab-form-controls--inline">
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
        </DashboardFormRow>
      ) : null}

      <DashboardFormRow label="" labelHidden>
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
          Bis Mitternacht?
        </label>
      </DashboardFormRow>

      <DashboardFormRow label="Projekt">
        <select
          className="tab-form-control"
          value={form.projectId}
          onChange={(event) => onChange({ ...form, projectId: Number(event.target.value), taskId: Number(event.target.value) })}
        >
          <option value={0}></option>
          {workItems.map((workItem) => (
            <option key={workItem.id} value={workItem.id}>
              {"- ".repeat(Math.max(0, workItem.depth))}
              {workItem.name}
            </option>
          ))}
        </select>
      </DashboardFormRow>

      <DashboardFormRow label="Taetigkeit">
        <select className="tab-form-control" value={form.taskId} onChange={(event) => onChange({ ...form, taskId: Number(event.target.value) })}>
          <option value={form.projectId}>{selectedWorkItem.taskName || selectedWorkItem.projectName}</option>
        </select>
      </DashboardFormRow>

      <DashboardFormRow label="Notiz">
        <textarea
          className="tab-form-control"
          rows={4}
          value={form.description}
          onChange={(event) => onChange({ ...form, description: event.target.value })}
        />
      </DashboardFormRow>
    </DashboardModal>
  );
}

type DashboardDateInputProps = {
  className?: string;
  language: DateLanguage;
  value: string;
  onChange: (value: string) => void;
};

function DashboardDateInput({ className, language, value, onChange }: DashboardDateInputProps) {
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
