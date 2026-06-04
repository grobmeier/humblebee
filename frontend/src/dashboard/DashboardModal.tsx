import type { FormEventHandler, ReactNode } from "react";

type DashboardModalProps = {
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
  onSubmit: FormEventHandler<HTMLFormElement>;
  title: string;
};

export function DashboardModal({ children, footer, onClose, onSubmit, title }: DashboardModalProps) {
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

type DashboardFormRowProps = {
  children: ReactNode;
  controlsClassName?: string;
  label: string;
  labelHidden?: boolean;
};

export function DashboardFormRow({ children, controlsClassName = "tab-form-controls", label, labelHidden = false }: DashboardFormRowProps) {
  return (
    <div className="tab-form-row">
      <label className="tab-form-label" aria-hidden={labelHidden || undefined}>
        {label}
      </label>
      <div className={controlsClassName}>{children}</div>
    </div>
  );
}
