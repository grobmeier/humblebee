export type Language = "de" | "en";

export const translations = {
  de: {
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
  stopwatch: {
    createStopwatch: string;
    book: string;
    discardRunning: string;
    discardRunningConfirm: string;
    start: string;
    stopStopwatch: string;
  };
}>;
