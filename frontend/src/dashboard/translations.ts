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
      cancel: "Abbrechen",
      completedTask: "Aufgabe erledigt",
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
      projectList: "Projekte",
      saveProject: "Projekt speichern",
      selectProject: "Waehle links ein Projekt aus.",
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
      loadingReport: "Bericht wird geladen...",
      monthly: "Monatlich",
      months: ["Januar", "Februar", "Maerz", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"],
      print: "Drucken",
      reportList: "Berichte",
      savedTo: "Gespeichert unter",
      titles: {
        "worktime-by-month": "Arbeitszeit pro Monat",
        "worktime-grouped-by-project": "Arbeitszeit nach Projekt",
        "worktime-task-details": "Arbeitszeitdetails nach Aufgabe",
        timesheet: "Stundenzettel"
      }
    },
    stopwatch: {
      book: "Buchen",
      createStopwatch: "Stoppuhr anlegen",
      discardRunning: "Stoppuhr verwerfen",
      discardRunningConfirm: "Die Stoppuhr wird geloescht und nicht gebucht. Fortfahren?",
      start: "Starten",
      stopStopwatch: "Stoppen"
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
      cancel: "Cancel",
      completedTask: "Task completed",
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
      projectList: "Projects",
      saveProject: "Save Project",
      selectProject: "Select a project on the left.",
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
      loadingReport: "Loading report...",
      monthly: "Monthly",
      months: ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
      print: "Print",
      reportList: "Reports",
      savedTo: "Saved to",
      titles: {
        "worktime-by-month": "Worktime by month",
        "worktime-grouped-by-project": "Worktime grouped by project",
        "worktime-task-details": "Worktime task details",
        timesheet: "Timesheet"
      }
    },
    stopwatch: {
      book: "Book",
      createStopwatch: "Create stopwatch",
      discardRunning: "Discard stopwatch",
      discardRunningConfirm: "The stopwatch will be deleted and not booked. Continue?",
      start: "Start",
      stopStopwatch: "Stop"
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
    cancel: string;
    completedTask: string;
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
    projectList: string;
    saveProject: string;
    selectProject: string;
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
    loadingReport: string;
    monthly: string;
    months: string[];
    print: string;
    reportList: string;
    savedTo: string;
    titles: {
      "worktime-by-month": string;
      "worktime-grouped-by-project": string;
      "worktime-task-details": string;
      timesheet: string;
    };
  };
  stopwatch: {
    createStopwatch: string;
    book: string;
    discardRunning: string;
    discardRunningConfirm: string;
    start: string;
    stopStopwatch: string;
  };
}>;
