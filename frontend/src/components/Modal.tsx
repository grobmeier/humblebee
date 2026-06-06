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

import type { FormEventHandler, ReactNode } from "react";

type ModalProps = {
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
  onSubmit: FormEventHandler<HTMLFormElement>;
  title: string;
};

export function Modal({ children, footer, onClose, onSubmit, title }: ModalProps) {
  return (
    <div
      className="tab-modal"
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <form className="modal-form modal-form--compact tab-form tab-form--horizontal" onSubmit={onSubmit}>
        <div className="modal-form-header">
          <h3>{title}</h3>
          <button type="button" className="close" onClick={onClose}>
            x
          </button>
        </div>
        <div className="modal-body">{children}</div>
        {footer ? <div className="modal-footer">{footer}</div> : null}
      </form>
    </div>
  );
}

type FormRowProps = {
  children: ReactNode;
  controlsClassName?: string;
  label: string;
  labelHidden?: boolean;
};

export function FormRow({ children, controlsClassName = "tab-form-controls", label, labelHidden = false }: FormRowProps) {
  return (
    <div className="tab-form-row">
      <div className="tab-form-label" aria-hidden={labelHidden || undefined}>
        {label}
      </div>
      <div className={controlsClassName}>{children}</div>
    </div>
  );
}
