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
