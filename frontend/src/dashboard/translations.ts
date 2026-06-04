export type Language = "de" | "en";

export const translations = {
  de: {
    nav: {
      dashboard: "Arbeitsplatz",
      projects: "Projekte",
      reports: "Berichte"
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
  stopwatch: {
    createStopwatch: string;
    book: string;
    discardRunning: string;
    discardRunningConfirm: string;
    start: string;
    stopStopwatch: string;
  };
}>;
