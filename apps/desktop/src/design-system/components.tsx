import type {
  ButtonHTMLAttributes,
  HTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from "react";

export type ButtonVariant = "primary" | "secondary" | "ghost";
export type BadgeTone =
  | "neutral"
  | "accent"
  | "success"
  | "warning"
  | "danger"
  | "muted"
  | "info";
export type CardTone = "default" | "raised" | "inset" | "accent";

function classNames(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export function Button({
  variant = "secondary",
  className,
  type = "button",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
}) {
  return (
    <button
      type={type}
      className={classNames("button", `button--${variant}`, className)}
      {...props}
    />
  );
}

export function Card({
  tone = "default",
  className,
  ...props
}: HTMLAttributes<HTMLElement> & {
  tone?: CardTone;
}) {
  return <section className={classNames("card", `card--${tone}`, className)} {...props} />;
}

export function Panel({
  title,
  eyebrow,
  description,
  actions,
  footer,
  className,
  children,
  ...props
}: HTMLAttributes<HTMLElement> & {
  title: string;
  eyebrow?: string;
  description?: string;
  actions?: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <section className={classNames("panel", className)} {...props}>
      <div className="panel__header">
        <div className="panel__copy">
          {eyebrow && <p className="panel__eyebrow">{eyebrow}</p>}
          <h2 className="panel__title">{title}</h2>
          {description && <p className="panel__description">{description}</p>}
        </div>
        {actions && <div className="panel__actions">{actions}</div>}
      </div>
      <div className="panel__body">{children}</div>
      {footer && <div className="panel__footer">{footer}</div>}
    </section>
  );
}

export function Badge({
  tone = "neutral",
  className,
  children,
  ...props
}: HTMLAttributes<HTMLSpanElement> & {
  tone?: BadgeTone;
}) {
  return (
    <span className={classNames("badge", `badge--${tone}`, className)} {...props}>
      {children}
    </span>
  );
}

export function StatusDot({
  tone = "neutral",
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement> & { tone?: BadgeTone }) {
  return <span className={classNames("status-dot", `status-dot--${tone}`, className)} {...props} />;
}

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={classNames("input", className)} {...props} />;
}

export function Select({ className, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={classNames("select", className)} {...props} />;
}

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className={classNames("textarea", className)} {...props} />;
}

export function SectionHeader({
  title,
  subtitle,
  actions,
  kicker,
  className,
  ...props
}: HTMLAttributes<HTMLElement> & {
  title: string;
  subtitle?: string;
  actions?: ReactNode;
  kicker?: string;
}) {
  return (
    <header className={classNames("section-header", className)} {...props}>
      <div className="section-header__copy">
        {kicker && <p className="section-header__kicker">{kicker}</p>}
        <h2 className="section-header__title">{title}</h2>
        {subtitle && <p className="section-header__subtitle">{subtitle}</p>}
      </div>
      {actions && <div className="section-header__actions">{actions}</div>}
    </header>
  );
}

export function MetricCard({
  label,
  value,
  detail,
  tone = "default",
  className,
  children,
  ...props
}: HTMLAttributes<HTMLElement> & {
  label: string;
  value: string;
  detail?: string;
  tone?: CardTone;
}) {
  return (
    <article className={classNames("metric-card", `metric-card--${tone}`, className)} {...props}>
      <p className="metric-card__label">{label}</p>
      <div className="metric-card__value">{value}</div>
      {detail && <p className="metric-card__detail">{detail}</p>}
      {children}
    </article>
  );
}

export function ProgressBar({
  value,
  tone = "accent",
  label,
  detail,
  className,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  value: number;
  tone?: BadgeTone;
  label?: string;
  detail?: string;
}) {
  const clamped = Math.max(0, Math.min(100, Math.round(value)));
  const bucket = Math.round(clamped / 10) * 10;

  return (
    <div className={classNames("progress", className)} {...props}>
      {(label || detail) && (
        <div className="progress__meta">
          {label && <span className="progress__label">{label}</span>}
          {detail && <span className="progress__detail">{detail}</span>}
        </div>
      )}
      <div className="progress__track">
        <div className={classNames("progress__fill", `progress__fill--${tone}`, `progress__fill--${bucket}`)} />
      </div>
    </div>
  );
}

export function CodeBlock({
  title,
  subtitle,
  children,
  className,
  ...props
}: HTMLAttributes<HTMLElement> & {
  title?: string;
  subtitle?: string;
}) {
  return (
    <section className={classNames("code-block", className)} {...props}>
      {(title || subtitle) && (
        <div className="code-block__header">
          {title && <p className="code-block__title">{title}</p>}
          {subtitle && <p className="code-block__subtitle">{subtitle}</p>}
        </div>
      )}
      <pre className="code-block__body">
        <code>{children}</code>
      </pre>
    </section>
  );
}

export function CommandBar({
  value,
  onChange,
  onSubmit,
  placeholder = "Type a command...",
  status,
  prompt = "/",
  hint,
  className,
  ...props
}: Omit<HTMLAttributes<HTMLFormElement>, "onChange" | "onSubmit"> & {
  value: string;
  onChange: (nextValue: string) => void;
  onSubmit?: () => void;
  placeholder?: string;
  status?: string;
  prompt?: string;
  hint?: string;
}) {
  return (
    <form
      className={classNames("command-bar", className)}
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit?.();
      }}
      {...props}
    >
      <div className="command-bar__prompt">{prompt}</div>
      <input
        className="command-bar__input"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
      />
      {(status || hint) && (
        <div className="command-bar__meta">
          {status && <span className="command-bar__status">{status}</span>}
          {hint && <span className="command-bar__hint">{hint}</span>}
        </div>
      )}
    </form>
  );
}
