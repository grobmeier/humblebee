import { useEffect, useRef, useState } from "react";
import { GetNews } from "../../wailsjs/go/guiapp/App";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import { Modal } from "../components/Modal";
import { newsCopy } from "./copy";
import "./news.css";

type Language = "de" | "en";
type NewsResult = {
  items: { guid: string; title: string; url: string; publishedAt: string; summary: string }[];
  fetchedAt: string;
  cached: boolean;
  unavailable: boolean;
  cacheWriteFailed: boolean;
};

function openArticle(raw: string) {
  try {
    const url = new URL(raw);
    if (url.origin === "https://www.timeandbill.de" && !url.username && !url.password && !url.search && !url.hash) {
      BrowserOpenURL(raw);
    }
  } catch { /* Untrusted links must never reach the browser. */ }
}

export function NewsModal({ language, onClose }: { language: Language; onClose: () => void }) {
  const t = newsCopy[language];
  const [result, setResult] = useState<NewsResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [revision, setRevision] = useState(0);
  const content = useRef<HTMLDivElement>(null);
  const closeCallback = useRef(onClose);
  const request = useRef<{ language: Language; revision: number; promise: Promise<NewsResult> } | null>(null);

  useEffect(() => {
    closeCallback.current = onClose;
  }, [onClose]);

  useEffect(() => {
    let active = true;
    setResult(null);
    setLoading(true);
    setFailed(false);
    // StrictMode replays effects while retaining refs. Reuse this opening's request.
    if (!request.current || request.current.language !== language || request.current.revision !== revision) {
      request.current = { language, revision, promise: GetNews(language) };
    }
    request.current.promise.then((news: NewsResult) => {
      if (active) setResult(news);
    }).catch(() => {
      if (active) setFailed(true);
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [language, revision]);

  useEffect(() => {
    content.current?.closest("form")?.setAttribute("aria-label", t.title);
  }, [t.title]);

  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null;
    const dialog = content.current?.closest("form");
    if (!dialog) return;
    dialog.setAttribute("role", "dialog");
    dialog.setAttribute("aria-modal", "true");
    const close = dialog.querySelector<HTMLButtonElement>(".close");
    close?.focus();
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") { event.preventDefault(); closeCallback.current(); }
      if (event.key !== "Tab") return;
      const buttons = Array.from(dialog!.querySelectorAll<HTMLElement>("button:not(:disabled), a[href], [tabindex='0']"));
      const first = buttons[0], last = buttons[buttons.length - 1];
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
      if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
    }
    dialog.addEventListener("keydown", onKey);
    return () => { dialog.removeEventListener("keydown", onKey); previous?.focus(); };
  }, []);

  const unavailable = failed || result?.unavailable;
  const formatDate = (date: string) => new Date(date).toLocaleString(language === "de" ? "de-DE" : "en-GB");
  return <Modal title={t.title} onClose={onClose} onSubmit={(event) => event.preventDefault()} footer={<>
    <button type="button" className="secondary-button" disabled={loading} onClick={() => setRevision((value) => value + 1)}>{unavailable ? t.retry : t.refresh}</button>
    <button type="button" className="secondary-button" onClick={() => openArticle(`https://www.timeandbill.de/${language}/humblebee/news/`)}>{t.index}</button>
  </>}>
    <div className="news-content" ref={content}>
      <p className="news-privacy">{t.privacy}</p>
      <div role="status" aria-live="polite">
        {loading && <p>{t.loading}</p>}
        {unavailable && <p className="alert alert-warning">{result?.cached ? t.cached : t.unavailable}</p>}
        {result?.cacheWriteFailed && <p className="alert alert-warning">{t.cacheWarning}</p>}
        {result?.fetchedAt && <p className="news-updated">{t.updated}: {formatDate(result.fetchedAt)}</p>}
        {!loading && !unavailable && result?.items.length === 0 && <p>{t.empty}</p>}
      </div>
      <ul className="news-items">{result?.items.map((item) => <li key={item.guid}>
        <a href={item.url} onClick={(event) => { event.preventDefault(); openArticle(item.url); }}>{item.title}</a>
        <time dateTime={item.publishedAt}>{new Date(item.publishedAt).toLocaleDateString(language === "de" ? "de-DE" : "en-GB")}</time>
        <p>{item.summary}</p>
      </li>)}</ul>
    </div>
  </Modal>;
}
