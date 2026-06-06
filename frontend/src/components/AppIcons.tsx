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

type IconProps = {
  className?: string;
};

export function ImportIcon({ className }: IconProps) {
  return (
    <svg className={className} aria-hidden="true" focusable="false" viewBox="0 0 24 24">
      <path d="M12 3v11" />
      <path d="m8 10 4 4 4-4" />
      <path d="M5 15v4h14v-4" />
    </svg>
  );
}

export function DatabaseSwitchIcon({ className }: IconProps) {
  return (
    <svg className={className} aria-hidden="true" focusable="false" viewBox="0 0 24 24">
      <ellipse cx="8" cy="5" rx="5" ry="2.5" />
      <path d="M3 5v7c0 1.4 2.2 2.5 5 2.5 1.2 0 2.3-.2 3.2-.6" />
      <path d="M3 9c0 1.4 2.2 2.5 5 2.5 1.2 0 2.3-.2 3.2-.6" />
      <path d="M15 12h5" />
      <path d="m18 9 3 3-3 3" />
      <path d="M20 17h-5" />
      <path d="m17 20-3-3 3-3" />
    </svg>
  );
}
