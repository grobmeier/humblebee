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

import type { ReactNode } from "react";

export type SideNavigationItem = {
  className?: string;
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
              <a className={`side-navigation-item ${item.className ?? ""} ${item.id === selectedId ? "selected" : ""}`} href={item.href} key={item.id}>
                {item.label}
              </a>
            ) : (
              <button
                className={`side-navigation-item ${item.className ?? ""} ${item.id === selectedId ? "selected" : ""}`}
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
