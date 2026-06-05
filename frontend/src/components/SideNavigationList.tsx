import type { ReactNode } from "react";

export type SideNavigationItem = {
  href?: string;
  id: number | string;
  label: string;
};

type SideNavigationListProps = {
  action?: ReactNode;
  ariaLabel: string;
  emptyText?: string;
  items: SideNavigationItem[];
  selectedId?: number | string;
  title: string;
  onSelect?: (id: number | string) => void;
};

export function SideNavigationList({ action, ariaLabel, emptyText, items, selectedId, title, onSelect }: SideNavigationListProps) {
  return (
    <aside className="side-navigation-panel" aria-label={ariaLabel}>
      <div className="side-navigation-header">
        <h2>{title}</h2>
        {action}
      </div>

      {items.length ? (
        <div className="side-navigation-list">
          {items.map((item) =>
            item.href ? (
              <a className={`side-navigation-item ${item.id === selectedId ? "selected" : ""}`} href={item.href} key={item.id}>
                {item.label}
              </a>
            ) : (
              <button
                className={`side-navigation-item ${item.id === selectedId ? "selected" : ""}`}
                key={item.id}
                type="button"
                onClick={() => onSelect?.(item.id)}
              >
                {item.label}
              </button>
            )
          )}
        </div>
      ) : (
        <p className="projects-empty">{emptyText}</p>
      )}
    </aside>
  );
}
