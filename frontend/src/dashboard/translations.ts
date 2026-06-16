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

export type Language = "de" | "en";

export const translations = {
  de: {
    dashboardCalendar: {
      addTime: "Zeit erfassen",
      currentWeek: "Aktuelle Woche",
      navigation: "Kalendernavigation",
      nextDay: "Naechster Tag",
      nextWeek: "Naechste Woche",
      previousDay: "Vorheriger Tag",
      previousWeek: "Vorherige Woche",
      today: "Heute"
    },
    nav: {
      dashboard: "Arbeitsplatz",
      projects: "Projekte",
      reports: "Berichte"
    },
    databasePage: {
      createNew: "Neue Datenbank erstellen",
      createNewHint: "Lege eine neue lokale HumbleBee-Datenbankdatei an.",
      current: "Aktuelle Datenbank",
      defaultPath: "Standarddatenbank",
      openExisting: "Andere Datenbank oeffnen",
      openExistingHint: "Waehle eine vorhandene HumbleBee-Datenbankdatei.",
      switchButton: "Datenbank wechseln",
      switchWarning: "",
      title: "Datenbank wechseln",
      useDefault: "Standard verwenden"
    },
    importPage: {
      alreadyImported: "Dieser Time & Bill-Export wurde bereits importiert.",
      chooseFile: "Datei waehlen",
      completed: "Import abgeschlossen.",
      conflictDetails: "{count} Konflikt(e) anzeigen",
      conflicts: "Konflikte",
      created: "Angelegt",
      existingTimeWarning: "Diese Datenbank enthaelt bereits {count} gebuchte Zeiteintraege. HumbleBee ueberschreibt keine vorhandenen Zeiten; ueberschneidende importierte Zeiten werden uebersprungen.",
      exportedAt: "Exportiert am",
      exportUuid: "Export UUID",
      file: "Exportdatei",
      importAction: "Importieren",
      importing: "Importiert...",
      importButton: "Time & Bill importieren",
      mapped: "Bestehend",
      noFileSelected: "Keine Datei ausgewaehlt",
      preview: "Vorschau",
      previewing: "Prueft...",
      projects: "Projekte",
      skipped: "Uebersprungen",
      sourceUser: "Benutzer",
      tasks: "Aufgaben",
      timeEntries: "Zeiten",
      title: "Time & Bill importieren",
      wouldConflict: "Konflikt",
      wouldCreate: "Anlegen",
      wouldMap: "Bestehend",
      wouldSkip: "Ueberspringen"
    },
    placeholders: {
      reports: {
        eyebrow: "Noch nicht fertig",
        title: "Berichte",
        body: "Zeitauswertungen und Exporte werden hier spaeter verfuegbar sein."
      }
    },
    projectsPage: {
      addProject: "Projekt hinzufuegen",
      addTask: "Aufgabe hinzufuegen",
      archiveProject: "Projekt archivieren",
      cancel: "Abbrechen",
      completedTask: "Aufgabe erledigt",
      copyTasksFrom: "Aufgaben kopieren von",
      createProject: "Projekt anlegen",
      createTask: "Aufgabe anlegen",
      deleteProject: "Projekt loeschen",
      deleteProjectConfirm: "Projekt loeschen",
      deleteProjectTitle: "Projekt loeschen",
      deleteProjectWarning: "Das Projekt \"{name}\" wird geloescht. Alle Aufgaben und gebuchten Zeiten fuer dieses Projekt werden ebenfalls geloescht.",
      editProject: "Projekt bearbeiten",
      emptyProjects: "Noch keine Projekte vorhanden.",
      emptyTasks: "Noch keine Aufgaben fuer dieses Projekt.",
      name: "Name",
      nameRequired: "Bitte gib einen Namen ein.",
      noTaskTemplate: "Keine Aufgaben kopieren",
      projectList: "Projekte",
      reactivateProject: "Projekt wieder aktivieren",
      saveProject: "Projekt speichern",
      selectProject: "Waehle links ein Projekt aus.",
      showArchivedProjects: "Archivierte Projekte anzeigen",
      showHiddenTasks: "Ausgeblendete Aufgaben anzeigen",
      tasks: "Aufgaben"
    },
    reportsPage: {
      columns: {
        date: "Datum",
        description: "Beschreibung",
        duration: "Dauer",
        end: "Ende",
        project: "Projekt",
        projectTime: "Projektzeit",
        start: "Start",
        task: "Aufgabe",
        total: "Gesamt"
      },
      dateRange: "Datumsbereich",
      emptyReport: "Keine Berichtsdaten fuer diesen Zeitraum.",
      exportExcel: "Excel exportieren",
      filterMode: "Berichtsfilter",
      firstReportableProject: "Erstes auswertbares Projekt",
      fromMonth: "Von Monat",
      loadingReport: "Bericht wird geladen...",
      monthly: "Monatlich",
      months: ["Januar", "Februar", "Maerz", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"],
      print: "Drucken",
      reportList: "Berichte",
      savedTo: "Gespeichert unter",
      selectProject: "Projekt auswaehlen",
      toMonth: "Bis Monat",
      titles: {
        "worktime-by-month": "Arbeitszeit pro Monat",
        "worktime-grouped-by-project": "Arbeitszeit nach Projekt",
        "worktime-project-details": "Projektdetails",
        "worktime-task-details": "Arbeitszeitdetails nach Aufgabe",
        timesheet: "Stundenzettel"
      }
    },
    stopwatch: {
      book: "Buchen",
      createStopwatch: "Stoppuhr anlegen",
      discardRunning: "Stoppuhr verwerfen",
      discardRunningConfirm: "Die Stoppuhr wird geloescht und nicht gebucht. Fortfahren?",
      selectWorkItem: "Stoppuhr-Aufgabe",
      start: "Starten",
      stopStopwatch: "Stoppen"
    },
    timeEntryModal: {
      conflictMessage: "Die Stoppuhr ueberschneidet sich mit bereits gebuchter Zeit. Passe den Zeitraum an und speichere den Eintrag.",
      end: "Ende",
      note: "Notiz",
      project: "Projekt",
      save: "Speichern",
      saving: "Speichern...",
      selectProjectRequired: "Bitte waehle ein Projekt aus.",
      selectTaskRequired: "Bitte waehle eine Taetigkeit aus.",
      start: "Start",
      task: "Taetigkeit",
      title: "Zeiteintrag erfassen",
      untilMidnight: "Bis Mitternacht?"
    }
  },
  en: {
    dashboardCalendar: {
      addTime: "Add time",
      currentWeek: "Current week",
      navigation: "Calendar navigation",
      nextDay: "Next day",
      nextWeek: "Next week",
      previousDay: "Previous day",
      previousWeek: "Previous week",
      today: "Today"
    },
    nav: {
      dashboard: "Dashboard",
      projects: "Projects",
      reports: "Reports"
    },
    databasePage: {
      createNew: "Create New Database",
      createNewHint: "Create a new local HumbleBee database file.",
      current: "Current database",
      defaultPath: "Default database",
      openExisting: "Open Another Database",
      openExistingHint: "Choose an existing HumbleBee database file.",
      switchButton: "Switch database",
      switchWarning: "",
      title: "Switch Database",
      useDefault: "Use default"
    },
    importPage: {
      alreadyImported: "This Time & Bill export has already been imported.",
      chooseFile: "Choose file",
      completed: "Import completed.",
      conflictDetails: "Show {count} conflict(s)",
      conflicts: "Conflicts",
      created: "Created",
      existingTimeWarning: "This database already contains {count} booked time entries. HumbleBee will not overwrite existing time; overlapping imported time will be skipped.",
      exportedAt: "Exported at",
      exportUuid: "Export UUID",
      file: "Export file",
      importAction: "Import",
      importing: "Importing...",
      importButton: "Import Time & Bill",
      mapped: "Existing",
      noFileSelected: "No file selected",
      preview: "Preview",
      previewing: "Previewing...",
      projects: "Projects",
      skipped: "Skipped",
      sourceUser: "User",
      tasks: "Tasks",
      timeEntries: "Time entries",
      title: "Import Time & Bill",
      wouldConflict: "Conflict",
      wouldCreate: "Create",
      wouldMap: "Existing",
      wouldSkip: "Skip"
    },
    placeholders: {
      reports: {
        eyebrow: "Not ready yet",
        title: "Reports",
        body: "Time reports and exports will be available here later."
      }
    },
    projectsPage: {
      addProject: "Add Project",
      addTask: "Add Task",
      archiveProject: "Archive Project",
      cancel: "Cancel",
      completedTask: "Task completed",
      copyTasksFrom: "Copy tasks from",
      createProject: "Create Project",
      createTask: "Create Task",
      deleteProject: "Delete Project",
      deleteProjectConfirm: "Delete Project",
      deleteProjectTitle: "Delete Project",
      deleteProjectWarning: "The project \"{name}\" will be deleted. All tasks and booked time for this project will be deleted as well.",
      editProject: "Edit Project",
      emptyProjects: "No projects yet.",
      emptyTasks: "No tasks for this project yet.",
      name: "Name",
      nameRequired: "Enter a name.",
      noTaskTemplate: "Do not copy tasks",
      projectList: "Projects",
      reactivateProject: "Reactivate Project",
      saveProject: "Save Project",
      selectProject: "Select a project on the left.",
      showArchivedProjects: "Show archived projects",
      showHiddenTasks: "Show hidden tasks",
      tasks: "Tasks"
    },
    reportsPage: {
      columns: {
        date: "Date",
        description: "Description",
        duration: "Duration",
        end: "End",
        project: "Project",
        projectTime: "Project time",
        start: "Start",
        task: "Task",
        total: "Total"
      },
      dateRange: "Date range",
      emptyReport: "No report data for this period.",
      exportExcel: "Export Excel",
      filterMode: "Report filter mode",
      firstReportableProject: "First reportable project",
      fromMonth: "From month",
      loadingReport: "Loading report...",
      monthly: "Monthly",
      months: ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
      print: "Print",
      reportList: "Reports",
      savedTo: "Saved to",
      selectProject: "Select project",
      toMonth: "To month",
      titles: {
        "worktime-by-month": "Worktime by month",
        "worktime-grouped-by-project": "Worktime grouped by project",
        "worktime-project-details": "Project details",
        "worktime-task-details": "Worktime task details",
        timesheet: "Timesheet"
      }
    },
    stopwatch: {
      book: "Book",
      createStopwatch: "Create stopwatch",
      discardRunning: "Discard stopwatch",
      discardRunningConfirm: "The stopwatch will be deleted and not booked. Continue?",
      selectWorkItem: "Stopwatch task",
      start: "Start",
      stopStopwatch: "Stop"
    },
    timeEntryModal: {
      conflictMessage: "The stopwatch overlaps with booked time. Adjust the time range and save the entry.",
      end: "End",
      note: "Note",
      project: "Project",
      save: "Save",
      saving: "Saving...",
      selectProjectRequired: "Select a project.",
      selectTaskRequired: "Select a task.",
      start: "Start",
      task: "Task",
      title: "Record time entry",
      untilMidnight: "Until midnight?"
    }
  }
} satisfies Record<Language, {
  dashboardCalendar: {
    addTime: string;
    currentWeek: string;
    navigation: string;
    nextDay: string;
    nextWeek: string;
    previousDay: string;
    previousWeek: string;
    today: string;
  };
  nav: {
    dashboard: string;
    projects: string;
    reports: string;
  };
  databasePage: {
    createNew: string;
    createNewHint: string;
    current: string;
    defaultPath: string;
    openExisting: string;
    openExistingHint: string;
    switchButton: string;
    switchWarning: string;
    title: string;
    useDefault: string;
  };
  importPage: {
    alreadyImported: string;
    chooseFile: string;
    completed: string;
    conflictDetails: string;
    conflicts: string;
    created: string;
    existingTimeWarning: string;
    exportedAt: string;
    exportUuid: string;
    file: string;
    importAction: string;
    importing: string;
    importButton: string;
    mapped: string;
    noFileSelected: string;
    preview: string;
    previewing: string;
    projects: string;
    skipped: string;
    sourceUser: string;
    tasks: string;
    timeEntries: string;
    title: string;
    wouldConflict: string;
    wouldCreate: string;
    wouldMap: string;
    wouldSkip: string;
  };
  placeholders: {
    reports: {
      eyebrow: string;
      title: string;
      body: string;
    };
  };
  projectsPage: {
    addProject: string;
    addTask: string;
    archiveProject: string;
    cancel: string;
    completedTask: string;
    copyTasksFrom: string;
    createProject: string;
    createTask: string;
    deleteProject: string;
    deleteProjectConfirm: string;
    deleteProjectTitle: string;
    deleteProjectWarning: string;
    editProject: string;
    emptyProjects: string;
    emptyTasks: string;
    name: string;
    nameRequired: string;
    noTaskTemplate: string;
    projectList: string;
    reactivateProject: string;
    saveProject: string;
    selectProject: string;
    showArchivedProjects: string;
    showHiddenTasks: string;
    tasks: string;
  };
  reportsPage: {
    columns: {
      date: string;
      description: string;
      duration: string;
      end: string;
      project: string;
      projectTime: string;
      start: string;
      task: string;
      total: string;
    };
    dateRange: string;
    emptyReport: string;
    exportExcel: string;
    filterMode: string;
    firstReportableProject: string;
    fromMonth: string;
    loadingReport: string;
    monthly: string;
    months: string[];
    print: string;
    reportList: string;
    savedTo: string;
    selectProject: string;
    toMonth: string;
    titles: {
      "worktime-by-month": string;
      "worktime-grouped-by-project": string;
      "worktime-project-details": string;
      "worktime-task-details": string;
      timesheet: string;
    };
  };
  stopwatch: {
    createStopwatch: string;
    book: string;
    discardRunning: string;
    discardRunningConfirm: string;
    selectWorkItem: string;
    start: string;
    stopStopwatch: string;
  };
  timeEntryModal: {
    conflictMessage: string;
    end: string;
    note: string;
    project: string;
    save: string;
    saving: string;
    selectProjectRequired: string;
    selectTaskRequired: string;
    start: string;
    task: string;
    title: string;
    untilMidnight: string;
  };
}>;

export type DatabasePageText = typeof translations.en.databasePage;
export type ImportPageText = typeof translations.en.importPage;
